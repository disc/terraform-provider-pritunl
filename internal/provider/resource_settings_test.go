package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// This test pushes a certificate to the very same Pritunl instance the whole
// acceptance suite runs against, which makes Pritunl restart the web server
// serving its API. A malformed or mismatched pair would leave that instance
// without a usable HTTPS endpoint and break every other test, so the pair is
// generated as a valid self-signed certificate, the provider waits for the web
// server to come back after each write, and destroying the resource hands the
// certificate back to Pritunl, which regenerates its self-signed default.
// Setting PRITUNL_SKIP_SETTINGS_ACC_TEST opts out of it without touching the
// rest of the suite.
func TestAccPritunlSettings(t *testing.T) {
	if os.Getenv("PRITUNL_SKIP_SETTINGS_ACC_TEST") != "" {
		t.Skip("PRITUNL_SKIP_SETTINGS_ACC_TEST is set, skipping the pritunl_settings acceptance test")
	}

	t.Run("applies a certificate without error", func(t *testing.T) {
		cert, key := generateSelfSignedCertificate(t)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("pritunl_settings.test", "id", "settings"),
			resource.TestCheckResourceAttr("pritunl_settings.test", "server_cert", cert),
			resource.TestCheckResourceAttr("pritunl_settings.test", "server_key", key),
			resource.TestCheckResourceAttrSet("pritunl_settings.test", "server_port"),
			testAccCheckSettingsCertificate(cert),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckSettingsCertificateReset(cert),
			Steps: []resource.TestStep{
				{
					Config: testPritunlSettingsConfig(cert, key),
					Check:  check,
				},
				// the private key is never read back from the api
				importStep("pritunl_settings.test", "server_key"),
			},
		})
	})

	// Managing only the port must never take ownership of the certificate the
	// instance already serves: the certificate is read back into the state, so
	// it has to stay out of both the plan and the destroy.
	t.Run("manages the port without touching the certificate", func(t *testing.T) {
		if os.Getenv("TF_ACC") == "" {
			t.Skip("TF_ACC is not set, skipping the acceptance test")
		}

		settings, err := testClient.GetSettings()
		if err != nil {
			t.Fatalf("failed to read the settings of the test instance: %s", err)
		}

		certificate := strings.TrimSpace(settings.String("server_cert"))

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckSettingsCertificateKept(certificate),
			Steps: []resource.TestStep{
				{
					// the port already in use, so the web server is left alone
					Config: testPritunlSettingsPortConfig(settings.Int("server_port")),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings.test", "server_port", strconv.Itoa(settings.Int("server_port"))),
						// the private key is never read back, so managing only
						// the port leaves it out of the state entirely
						resource.TestCheckNoResourceAttr("pritunl_settings.test", "server_key"),
						testAccCheckSettingsCertificateKept(certificate),
					),
				},
			},
		})
	})

	// The reason the resource writes the complete settings object back: a
	// Terraform run that only manages the port must leave every other setting
	// of the instance alone. A few unrelated settings are configured out of
	// band, the way an operator would from the web console, and have to still
	// be there once Terraform has written the settings it manages.
	t.Run("leaves unrelated settings untouched", func(t *testing.T) {
		if os.Getenv("TF_ACC") == "" {
			t.Skip("TF_ACC is not set, skipping the acceptance test")
		}

		expected := configureUnrelatedSettings(t)

		settings, err := testClient.GetSettings()
		if err != nil {
			t.Fatalf("failed to read the settings of the test instance: %s", err)
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckUnrelatedSettings(expected),
			Steps: []resource.TestStep{
				{
					// the port already in use, so the web server is left alone
					Config: testPritunlSettingsPortConfig(settings.Int("server_port")),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings.test", "id", "settings"),
						testAccCheckUnrelatedSettings(expected),
					),
				},
			},
		})
	})
}

// Settings this provider does not manage, used to prove that writing the
// managed ones hands everything else back untouched. They are deliberately not
// sso_* settings: Pritunl clears all of those on any write while single sign-on
// itself is disabled, which would fail the check for a reason that has nothing
// to do with this provider. restrict_import and pin_mode are managed
// attributes, but ones the configuration of this test never mentions, which the
// round trip has to hand back just as untouched as the settings the resource
// knows nothing about.
var unrelatedSettings = map[string]interface{}{
	"restrict_import": true,
	"email_server":    "smtp.unrelated.example.com",
	"email_from":      "unrelated@example.com",
	"pin_mode":        "optional",
}

// configureUnrelatedSettings applies the unrelated settings through a full
// object write, restores the previous values once the test is over and returns
// the values the API reports for them, so that the checks compare against what
// Pritunl actually stored rather than against what was sent.
func configureUnrelatedSettings(t *testing.T) map[string]interface{} {
	t.Helper()

	settings, err := testClient.GetSettings()
	if err != nil {
		t.Fatalf("failed to read the settings of the test instance: %s", err)
	}

	previous := make(map[string]interface{}, len(unrelatedSettings))
	for key := range unrelatedSettings {
		previous[key] = settings[key]
	}

	for key, value := range unrelatedSettings {
		settings[key] = value
	}

	if err = testClient.UpdateSettings(settings); err != nil {
		t.Fatalf("failed to configure the unrelated settings: %s", err)
	}

	t.Cleanup(func() {
		current, err := testClient.GetSettings()
		if err != nil {
			return
		}

		for key, value := range previous {
			current[key] = value
		}

		testClient.UpdateSettings(current)
	})

	settings, err = testClient.GetSettings()
	if err != nil {
		t.Fatalf("failed to read the settings of the test instance: %s", err)
	}

	expected := make(map[string]interface{}, len(unrelatedSettings))
	for key := range unrelatedSettings {
		expected[key] = settings[key]

		if fmt.Sprintf("%v", settings[key]) != fmt.Sprintf("%v", unrelatedSettings[key]) {
			t.Fatalf("the unrelated setting %q has not been configured: got %v, want %v",
				key, settings[key], unrelatedSettings[key])
		}
	}

	return expected
}

func testAccCheckUnrelatedSettings(expected map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		settings, err := testClient.GetSettings()
		if err != nil {
			return err
		}

		for key, value := range expected {
			if fmt.Sprintf("%v", settings[key]) != fmt.Sprintf("%v", value) {
				return fmt.Errorf("the unmanaged setting %q has been modified: got %v, want %v",
					key, settings[key], value)
			}
		}

		return nil
	}
}

func testPritunlSettingsPortConfig(port int) string {
	return fmt.Sprintf(`
		resource "pritunl_settings" "test" {
			server_port = %[1]d
		}
	`, port)
}

func testAccCheckSettingsCertificateKept(certificate string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		settings, err := testClient.GetSettings()
		if err != nil {
			return err
		}

		if strings.TrimSpace(settings.String("server_cert")) != certificate {
			return fmt.Errorf("the certificate of the pritunl instance has been modified")
		}

		return nil
	}
}

func testPritunlSettingsConfig(cert, key string) string {
	return fmt.Sprintf(`
		resource "pritunl_settings" "test" {
			server_cert = %[1]q
			server_key  = %[2]q
		}
	`, cert, key)
}

func testAccCheckSettingsCertificate(cert string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		settings, err := testClient.GetSettings()
		if err != nil {
			return err
		}

		if strings.TrimSpace(settings.String("server_cert")) != cert {
			return fmt.Errorf("the certificate has not been applied on the pritunl instance")
		}

		return nil
	}
}

// Destroying the resource must leave the instance with a working certificate of
// its own, otherwise the following tests would not be able to reach the api.
func testAccCheckSettingsCertificateReset(cert string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		settings, err := testClient.GetSettings()
		if err != nil {
			return err
		}

		if strings.TrimSpace(settings.String("server_cert")) == "" {
			return fmt.Errorf("the pritunl instance has been left without a certificate")
		}

		if strings.TrimSpace(settings.String("server_cert")) == cert {
			return fmt.Errorf("the certificate is still applied on the pritunl instance")
		}

		return nil
	}
}

// The returned pair is trimmed to match what Pritunl stores and what the
// provider keeps in the state.
func generateSelfSignedCertificate(t *testing.T) (string, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate the test private key: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName:   "pritunl.local",
			Organization: []string{"terraform-provider-pritunl"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:              []string{"pritunl.local", "localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certificate, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to generate the test certificate: %s", err)
	}

	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate})
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return strings.TrimSpace(string(certPem)), strings.TrimSpace(string(keyPem))
}

// The instance wide toggles, the pin mode and the single sign-on configuration
// run against the very same Pritunl instance as the rest of the suite, so every
// one of them remembers the values it is about to change and puts them back
// once it is over. Single sign-on is the one that matters most: an instance
// left with it enabled would carry an organization the test deletes on its way
// out, and Pritunl refuses every further write to the settings of an instance
// whose single sign-on has no domain or no organization.
func TestAccPritunlSettingsToggles(t *testing.T) {
	if os.Getenv("PRITUNL_SKIP_SETTINGS_ACC_TEST") != "" {
		t.Skip("PRITUNL_SKIP_SETTINGS_ACC_TEST is set, skipping the pritunl_settings acceptance test")
	}

	for _, attribute := range settingsBoolAttributes {
		attribute := attribute

		t.Run(fmt.Sprintf("turns %s on and off again", attribute), func(t *testing.T) {
			if os.Getenv("TF_ACC") == "" {
				t.Skip("TF_ACC is not set, skipping the acceptance test")
			}

			restoreSettings(t, attribute)

			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preCheck(t) },
				ProviderFactories: providerFactories,
				Steps: []resource.TestStep{
					{
						Config: testPritunlSettingsAttributeConfig(attribute, "true"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("pritunl_settings.test", attribute, "true"),
							testAccCheckSettingsValues(map[string]interface{}{attribute: true}),
						),
					},
					{
						Config: testPritunlSettingsAttributeConfig(attribute, "false"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("pritunl_settings.test", attribute, "false"),
							testAccCheckSettingsValues(map[string]interface{}{attribute: false}),
						),
					},
				},
			})
		})
	}

	// sso_cache and sso_client_cache are two different caches with two very
	// similar names, one for every OpenVPN client and one for the Pritunl
	// client. Mixing them up in the schema or in the overlay would be invisible
	// from Terraform, both attributes would keep reporting the value they were
	// configured with, so each of them is written on its own and the other one
	// is read back from the instance to prove it stayed where it was.
	t.Run("keeps the two authentication caches apart", func(t *testing.T) {
		if os.Getenv("TF_ACC") == "" {
			t.Skip("TF_ACC is not set, skipping the acceptance test")
		}

		for _, cache := range []struct{ configured, other string }{
			{configured: "sso_cache", other: "sso_client_cache"},
			{configured: "sso_client_cache", other: "sso_cache"},
		} {
			cache := cache

			t.Run(fmt.Sprintf("writes %s without touching %s", cache.configured, cache.other), func(t *testing.T) {
				// both caches start off, so the only way the one left out of
				// the configuration can end up on is the write of the other one
				settingsBaseline(t, map[string]interface{}{
					cache.configured: false,
					cache.other:      false,
				})

				resource.Test(t, resource.TestCase{
					PreCheck:          func() { preCheck(t) },
					ProviderFactories: providerFactories,
					Steps: []resource.TestStep{
						{
							Config: testPritunlSettingsAttributeConfig(cache.configured, "true"),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr("pritunl_settings.test", cache.configured, "true"),
								resource.TestCheckResourceAttr("pritunl_settings.test", cache.other, "false"),
								testAccCheckSettingsValues(map[string]interface{}{
									cache.configured: true,
									cache.other:      false,
								}),
							),
						},
					},
				})
			})
		}
	})

	// A boolean cannot carry "not managed" in the state, so an attribute left
	// out of the configuration has to be told apart from one configured as
	// false through the raw configuration alone. Getting that wrong turns every
	// toggle of the instance off on the first apply of a configuration that
	// never mentions them.
	t.Run("leaves the toggles the configuration does not mention alone", func(t *testing.T) {
		if os.Getenv("TF_ACC") == "" {
			t.Skip("TF_ACC is not set, skipping the acceptance test")
		}

		expected := make(map[string]interface{}, len(settingsBoolAttributes))
		for _, attribute := range settingsBoolAttributes {
			expected[attribute] = true
		}

		settingsBaseline(t, expected)

		settings, err := testClient.GetSettings()
		if err != nil {
			t.Fatalf("failed to read the settings of the test instance: %s", err)
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckSettingsValues(expected),
			Steps: []resource.TestStep{
				{
					// the port already in use, so the web server is left alone
					Config: testPritunlSettingsPortConfig(settings.Int("server_port")),
					Check:  testAccCheckSettingsValues(expected),
				},
			},
		})
	})

	t.Run("applies every pin mode", func(t *testing.T) {
		if os.Getenv("TF_ACC") == "" {
			t.Skip("TF_ACC is not set, skipping the acceptance test")
		}

		restoreSettings(t, "pin_mode")

		steps := make([]resource.TestStep, 0, 4)
		for _, mode := range []string{"required", "disabled", "optional"} {
			steps = append(steps, resource.TestStep{
				Config: testPritunlSettingsAttributeConfig("pin_mode", fmt.Sprintf("%q", mode)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_settings.test", "pin_mode", mode),
					testAccCheckSettingsValues(map[string]interface{}{"pin_mode": mode}),
				),
			})
		}

		// importing reads every managed attribute back from the instance, which
		// is only ever going to match the state of an applied configuration
		// when the read covers all of them
		steps = append(steps, importStep("pritunl_settings.test", "server_key"))

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps:             steps,
		})
	})
}

// Pritunl clears sso_org, server_sso_url and the credentials of every single
// sign-on provider it knows as soon as the sso of a write is falsy, and the
// write of this resource always carries the complete settings object. A
// configuration that manages single sign-on and one that does not both have to
// come out of that with the single sign-on of the instance intact.
func TestAccPritunlSettingsSingleSignOn(t *testing.T) {
	if os.Getenv("PRITUNL_SKIP_SETTINGS_ACC_TEST") != "" {
		t.Skip("PRITUNL_SKIP_SETTINGS_ACC_TEST is set, skipping the pritunl_settings acceptance test")
	}

	// Okta is a SAML integration on this side of the API, so an Okta
	// configuration carries both the SAML settings of the identity provider and
	// the Okta ones. None of these credentials is real, the acceptance suite
	// never completes an authentication with them, it only checks that they
	// reach the instance and stay there. Pritunl neither parses nor validates
	// the SAML certificate, it hands it over to its SAML service as it comes,
	// so a self-signed one stands in for the one of an identity provider.
	samlCertificate, _ := generateSelfSignedCertificate(t)

	configured := map[string]interface{}{
		"sso":                 "saml_okta",
		"sso_saml_url":        "https://tfacc.okta.example.com/app/pritunl/sso/saml",
		"sso_saml_issuer_url": "https://www.okta.com/exktfaccsettings",
		"sso_saml_cert":       samlCertificate,
		"sso_okta_app_id":     "0oatfaccsettingsapp",
		"sso_okta_token":      "00tfaccsettingstoken",
		"sso_okta_mode":       "push",
		"server_sso_url":      "https://pritunl.local",
	}

	t.Run("configures okta and survives an unrelated write", func(t *testing.T) {
		if os.Getenv("TF_ACC") == "" {
			t.Skip("TF_ACC is not set, skipping the acceptance test")
		}

		restoreSettings(t, settingsSsoAttributes...)
		restoreSettings(t, "sso", "sso_okta_mode", "ipv6")

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			// destroying the resource never turns single sign-on off, it would
			// take the credentials of the instance down with it
			CheckDestroy: testAccCheckSettingsValues(configured),
			Steps: []resource.TestStep{
				{
					Config: testPritunlSettingsSsoConfig(configured, true),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSettingsAttributes(configured),
						resource.TestCheckResourceAttrPair(
							"pritunl_settings.test", "sso_org",
							"pritunl_organization.sso", "id",
						),
						testAccCheckSettingsSso(configured),
					),
				},
				{
					// nothing but a toggle that has no relation to single
					// sign-on changes here, and the write that carries it still
					// hands the whole settings object over
					Config: testPritunlSettingsSsoConfig(configured, false),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings.test", "ipv6", "false"),
						testAccCheckSettingsSso(configured),
					),
				},
				// the token is never read back from the api, the rest of the
				// single sign-on configuration is
				importStep("pritunl_settings.test", "server_key", "sso_okta_token"),
			},
		})
	})

	// The dangerous one: a configuration that never mentions single sign-on
	// still writes the whole settings object, so an empty sso slipping into
	// that write would wipe the configuration of an instance this resource was
	// never meant to touch.
	t.Run("keeps a single sign-on it does not manage", func(t *testing.T) {
		if os.Getenv("TF_ACC") == "" {
			t.Skip("TF_ACC is not set, skipping the acceptance test")
		}

		organization, err := testClient.CreateOrganization("tfacc-settings-sso")
		if err != nil {
			t.Fatalf("failed to create the organization of the single sign-on: %s", err)
		}

		// registered before the single sign-on is configured, so that it runs
		// after it has been turned off again: an organization referenced by a
		// single sign-on that is still enabled would be a dangling reference
		t.Cleanup(func() {
			testClient.DeleteOrganization(organization.ID)
		})

		unmanaged := settingsWith(configured, map[string]interface{}{
			"sso_org": organization.ID,
		})

		settingsBaseline(t, unmanaged)
		restoreSettings(t, "ipv6")

		settings, err := testClient.GetSettings()
		if err != nil {
			t.Fatalf("failed to read the settings of the test instance: %s", err)
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testAccCheckSettingsValues(unmanaged),
			Steps: []resource.TestStep{
				{
					Config: testPritunlSettingsAttributeConfig("ipv6", strconv.FormatBool(!settings.Bool("ipv6"))),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings.test", "ipv6", strconv.FormatBool(!settings.Bool("ipv6"))),
						testAccCheckSettingsValues(unmanaged),
					),
				},
			},
		})
	})

	// The secondary factor of Okta is the one managed setting whose empty value
	// is a value of its own, the factor turned off, which the overlay can only
	// tell apart from an unmanaged attribute through the raw configuration.
	t.Run("applies every okta secondary factor", func(t *testing.T) {
		if os.Getenv("TF_ACC") == "" {
			t.Skip("TF_ACC is not set, skipping the acceptance test")
		}

		restoreSettings(t, settingsSsoAttributes...)
		restoreSettings(t, "sso", "sso_okta_mode")

		steps := make([]resource.TestStep, 0, 5)
		for _, mode := range []string{"passcode", "push", "push_none", ""} {
			mode := mode

			modeConfigured := settingsWith(configured, map[string]interface{}{
				"sso_okta_mode": mode,
			})

			steps = append(steps, resource.TestStep{
				Config: testPritunlSettingsSsoConfig(modeConfigured, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_settings.test", "sso_okta_mode", mode),
					testAccCheckSettingsSso(modeConfigured),
				),
			})
		}

		// importing reads the whole single sign-on configuration back, the
		// token aside, which is only ever going to match the state of an
		// applied configuration when the read covers all of it
		steps = append(steps, importStep("pritunl_settings.test", "server_key", "sso_okta_token"))

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps:             steps,
		})
	})

	// The web console only offers the secondary factor of Okta while the
	// provider is exactly saml_okta, and the API turns out to enforce that
	// rather than merely hide it: the handler drops the setting for every other
	// provider. That is what the RequiredWith between sso_okta_mode and sso
	// stands on, so it is checked against the API itself rather than assumed.
	t.Run("proves the api ignores the okta secondary factor without okta", func(t *testing.T) {
		if os.Getenv("TF_ACC") == "" {
			t.Skip("TF_ACC is not set, skipping the acceptance test")
		}

		organization, err := testClient.CreateOrganization("tfacc-settings-sso-mode")
		if err != nil {
			t.Fatalf("failed to create the organization of the single sign-on: %s", err)
		}

		t.Cleanup(func() {
			testClient.DeleteOrganization(organization.ID)
		})

		restoreSettings(t, settingsSsoAttributes...)
		restoreSettings(t, "sso", "sso_okta_mode")

		// a mode that is neither the one about to be written nor the empty
		// default, applied while single sign-on is on and the handler takes it
		enabled := settingsWith(configured, map[string]interface{}{
			"sso_org":       organization.ID,
			"sso_okta_mode": "passcode",
		})

		settingsBaseline(t, enabled)

		settings, err := testClient.GetSettings()
		if err != nil {
			t.Fatalf("failed to read the settings of the test instance: %s", err)
		}

		if mode := settings.String("sso_okta_mode"); mode != "passcode" {
			t.Fatalf("the okta secondary factor has not been configured: got %q, want passcode", mode)
		}

		// single sign-on off, which also clears every credential of it
		writeSettings(t, map[string]interface{}{"sso": ""})
		writeSettings(t, map[string]interface{}{"sso_okta_mode": "push"})

		settings, err = testClient.GetSettings()
		if err != nil {
			t.Fatalf("failed to read the settings of the test instance: %s", err)
		}

		if mode := settings.String("sso_okta_mode"); mode != "passcode" {
			t.Fatalf("the pritunl api took the okta secondary factor over with single sign-on disabled: got %q, want the passcode it was left with", mode)
		}
	})
}

// restoreSettings remembers the settings a test is about to change and puts
// them back once it is over, so that the instance the whole suite shares is
// left the way it was found.
func restoreSettings(t *testing.T, keys ...string) {
	t.Helper()

	settings, err := testClient.GetSettings()
	if err != nil {
		t.Fatalf("failed to read the settings of the test instance: %s", err)
	}

	previous := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		previous[key] = settings[key]
	}

	t.Cleanup(func() {
		current, err := testClient.GetSettings()
		if err != nil {
			t.Errorf("failed to read the settings of the test instance back: %s", err)

			return
		}

		for key, value := range previous {
			current[key] = value
		}

		if err = testClient.UpdateSettings(current); err != nil {
			t.Errorf("failed to restore the settings of the test instance: %s", err)
		}
	})
}

// settingsBaseline configures the given settings the way an operator would from
// the web console, and restores the values it found once the test is over.
func settingsBaseline(t *testing.T, values map[string]interface{}) {
	t.Helper()

	restoreSettings(t, sortedKeys(values)...)
	writeSettings(t, values)
}

// writeSettings configures the given settings without remembering anything, for
// the steps of a test a restore registered earlier already covers. Restoring
// every step of a single sign-on configuration on its own would put the pieces
// back one by one, and Pritunl rejects the intermediate states: a provider
// without an organization is answered with a 400.
func writeSettings(t *testing.T, values map[string]interface{}) {
	t.Helper()

	settings, err := testClient.GetSettings()
	if err != nil {
		t.Fatalf("failed to read the settings of the test instance: %s", err)
	}

	for key, value := range values {
		settings[key] = value
	}

	if err = testClient.UpdateSettings(settings); err != nil {
		t.Fatalf("failed to configure the settings of the test instance: %s", err)
	}
}

func testAccCheckSettingsValues(expected map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		settings, err := testClient.GetSettings()
		if err != nil {
			return err
		}

		for key, value := range expected {
			if fmt.Sprintf("%v", settings[key]) != fmt.Sprintf("%v", value) {
				return fmt.Errorf("the setting %q of the pritunl instance is %v, want %v",
					key, settings[key], value)
			}
		}

		return nil
	}
}

// testAccCheckSettingsSso checks the single sign-on configuration of the
// instance against the organization Terraform created for it, whose id is only
// known once the configuration has been applied.
func testAccCheckSettingsSso(expected map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		organization, ok := s.RootModule().Resources["pritunl_organization.sso"]
		if !ok {
			return fmt.Errorf("the organization of the single sign-on is missing from the state")
		}

		values := map[string]interface{}{"sso_org": organization.Primary.ID}
		for key, value := range expected {
			values[key] = value
		}

		return testAccCheckSettingsValues(values)(s)
	}
}

func testPritunlSettingsAttributeConfig(attribute, value string) string {
	return fmt.Sprintf(`
		resource "pritunl_settings" "test" {
			%[1]s = %[2]s
		}
	`, attribute, value)
}

// testPritunlSettingsSsoConfig renders the whole single sign-on configuration,
// the organization it points at included, so that sso_org is a reference to a
// pritunl_organization the way an operator would write it.
func testPritunlSettingsSsoConfig(values map[string]interface{}, ipv6 bool) string {
	attributes := make([]string, 0, len(values))
	for _, key := range sortedKeys(values) {
		attributes = append(attributes, fmt.Sprintf("\t\t\t%s = %q", key, values[key]))
	}

	return fmt.Sprintf(`
		resource "pritunl_organization" "sso" {
			name = "tfacc-settings-sso-org"
		}

		resource "pritunl_settings" "test" {
%[1]s
			sso_org = pritunl_organization.sso.id
			ipv6    = %[2]t
		}
	`, strings.Join(attributes, "\n"), ipv6)
}

// settingsWith copies a set of settings with a few of them overridden, which is
// what keeps the subtests from having to spell the whole single sign-on
// configuration out again to change one of its values.
func settingsWith(base, overrides map[string]interface{}) map[string]interface{} {
	values := make(map[string]interface{}, len(base)+len(overrides))

	for key, value := range base {
		values[key] = value
	}

	for key, value := range overrides {
		values[key] = value
	}

	return values
}

func sortedKeys(values map[string]interface{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// testAccCheckSettingsAttributes checks the state of the resource against the
// values it was configured with.
func testAccCheckSettingsAttributes(values map[string]interface{}) resource.TestCheckFunc {
	checks := make([]resource.TestCheckFunc, 0, len(values))
	for _, key := range sortedKeys(values) {
		checks = append(checks, resource.TestCheckResourceAttr(
			"pritunl_settings.test", key, fmt.Sprintf("%v", values[key])))
	}

	return resource.ComposeTestCheckFunc(checks...)
}
