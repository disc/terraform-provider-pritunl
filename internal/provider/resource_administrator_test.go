package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/lfventura/terraform-provider-pritunl/internal/pritunl"
)

// These tests manage the very kind of account the whole suite authenticates
// with, so every one of them creates a disposable administrator of its own and
// never touches the one the instance was seeded with, the one holding the
// credentials of PRITUNL_TOKEN and PRITUNL_SECRET. Writing to that account
// would revoke the credentials of the test client, of the provider under test
// and of every test that runs afterwards, and recovering from it takes a fresh
// container. testAccCheckSeedAdministratorIntact guards against getting that
// wrong by accident: it is part of the CheckDestroy of every case below and
// fails the test as soon as the seeded account has moved at all.

func TestAccPritunlAdministrator(t *testing.T) {
	t.Run("creates an administrator with the minimal configuration", func(t *testing.T) {
		username := "tfacc-admin-minimal"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckAdministratorDestroyed(username),
				testAccCheckSeedAdministratorIntact(t),
			),
			Steps: []resource.TestStep{
				{
					Config: testPritunlAdministratorConfig(username, `password = "tfacc-Passw0rd!"`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "username", username),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "super_user", "false"),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "auth_api", "false"),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "disabled", "false"),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "otp_auth", "false"),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "default", "false"),
						// the credentials of an account without api access
						// neither authenticate nor survive it being turned on
						resource.TestCheckResourceAttr("pritunl_administrator.test", "token", ""),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "secret", ""),
						testAccCheckAdministrator(username, map[string]interface{}{
							"super_user": false,
							"auth_api":   false,
							"disabled":   false,
						}),
					),
				},
				// the password is never read back from the api
				importStep("pritunl_administrator.test", "password"),
			},
		})
	})

	// The username Pritunl stores is not always the one it was given: it drops
	// every character it does not consider safe and lowercases the rest, so a
	// configuration carrying one of those would never settle without the
	// resource normalising it the same way.
	t.Run("normalises the username the way pritunl does", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckAdministratorDestroyed("tfaccadminnormalised"),
				testAccCheckSeedAdministratorIntact(t),
			),
			Steps: []resource.TestStep{
				{
					Config: testPritunlAdministratorConfig("TFACC Admin Normalised", `password = "tfacc-Passw0rd!"`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "username", "tfaccadminnormalised"),
						testAccCheckAdministrator("tfaccadminnormalised", nil),
					),
				},
			},
		})
	})

	t.Run("creates a service account administrator with api credentials", func(t *testing.T) {
		username := "tfacc-admin-service"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckAdministratorDestroyed(username),
				testAccCheckSeedAdministratorIntact(t),
			),
			Steps: []resource.TestStep{
				{
					Config: testPritunlAdministratorConfig(username, `
						password   = "tfacc-Passw0rd!"
						super_user = true
						auth_api   = true
					`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "super_user", "true"),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "auth_api", "true"),
						resource.TestCheckResourceAttrSet("pritunl_administrator.test", "token"),
						resource.TestCheckResourceAttrSet("pritunl_administrator.test", "secret"),
						testAccCheckAdministratorCredentials(username),
						testAccCheckAdministrator(username, map[string]interface{}{
							"super_user": true,
							"auth_api":   true,
						}),
					),
				},
				importStep("pritunl_administrator.test", "password"),
			},
		})
	})

	t.Run("updates the username, the yubikey id and the disabled flag", func(t *testing.T) {
		username := "tfacc-admin-updated"
		renamed := "tfacc-admin-renamed"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckAdministratorDestroyed(username),
				testAccCheckAdministratorDestroyed(renamed),
				testAccCheckSeedAdministratorIntact(t),
			),
			Steps: []resource.TestStep{
				{
					Config: testPritunlAdministratorConfig(username, `password = "tfacc-Passw0rd!"`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "username", username),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "yubikey_id", ""),
					),
				},
				{
					// a whole OTP is truncated to the public id of the key by
					// Pritunl, and by the resource so that the plan settles
					Config: testPritunlAdministratorConfig(renamed, `
						password   = "tfacc-Passw0rd!"
						yubikey_id = "ccccccbcdefgtrljvnvtdjhbvrhlchgcbeghdthflukj"
						disabled   = true
					`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "username", renamed),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "yubikey_id", "ccccccbcdefg"),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "disabled", "true"),
						testAccCheckAdministrator(renamed, map[string]interface{}{
							"yubikey_id": "ccccccbcdefg",
							"disabled":   true,
						}),
					),
				},
				{
					// the empty id is the key taken off the account, a value
					// of its own the SDK drops from the plan without help
					Config: testPritunlAdministratorConfig(renamed, `
						password   = "tfacc-Passw0rd!"
						yubikey_id = ""
						disabled   = false
					`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "yubikey_id", ""),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "disabled", "false"),
						testAccCheckAdministrator(renamed, map[string]interface{}{
							"yubikey_id": nil,
							"disabled":   false,
						}),
					),
				},
			},
		})
	})

	// The reason the resource writes the complete administrator back, and the
	// regression this whole design exists for.
	//
	// PUT /admin/<id> is a full replace and the request body the Pritunl web
	// frontend builds carries every field of it, zeroed for whatever the
	// caller left out. A resource that wrote only the attribute that changed
	// would therefore take the super user flag off this account, turn its API
	// access off and, because Pritunl discards the credentials of an account
	// that loses API access and generates a fresh pair on the next write,
	// hand out a different token and secret, all from an apply that asked for
	// nothing but a disabled flag. Every one of those is checked against the
	// API itself rather than against the state, so that a resource reporting
	// values it did not actually write cannot pass.
	t.Run("keeps the other attributes on an unrelated write", func(t *testing.T) {
		username := "tfacc-admin-roundtrip"
		credentials := &administratorCredentials{}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckAdministratorDestroyed(username),
				testAccCheckSeedAdministratorIntact(t),
			),
			Steps: []resource.TestStep{
				{
					Config: testPritunlAdministratorConfig(username, `
						password   = "tfacc-Passw0rd!"
						super_user = true
						auth_api   = true
						yubikey_id = "ccccccbcdefg"
					`),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAdministratorCredentials(username),
						testAccRecordAdministratorCredentials(username, credentials),
					),
				},
				{
					// the attributes that were applied above are deliberately
					// gone from the configuration: they are optional and
					// computed, so they keep the value the account holds, and
					// the only thing this apply asks for is the disabled flag.
					// A resource writing nothing but what changed would take
					// them all off the account here.
					Config: testPritunlAdministratorConfig(username, `
						password = "tfacc-Passw0rd!"
						disabled = true
					`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "disabled", "true"),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "super_user", "true"),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "auth_api", "true"),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "yubikey_id", "ccccccbcdefg"),
						testAccCheckAdministrator(username, map[string]interface{}{
							"disabled":   true,
							"super_user": true,
							"auth_api":   true,
							"yubikey_id": "ccccccbcdefg",
						}),
						testAccCheckAdministratorCredentialsKept(username, credentials),
					),
				},
				{
					// and back, so that the account is left enabled: Pritunl
					// counts only the enabled super users and refuses to
					// remove a super user while it is the only other one, see
					// the destroy protection case below
					Config: testPritunlAdministratorConfig(username, `
						password = "tfacc-Passw0rd!"
						disabled = false
					`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "disabled", "false"),
						testAccCheckAdministrator(username, map[string]interface{}{
							"disabled":   false,
							"super_user": true,
							"auth_api":   true,
							"yubikey_id": "ccccccbcdefg",
						}),
						testAccCheckAdministratorCredentialsKept(username, credentials),
					),
				},
			},
		})
	})

	// An attribute the configuration never mentions is left to the account,
	// the same way pritunl_settings leaves an unmanaged setting to the
	// instance. Here the flags are set out of band, the way an operator would
	// from the web console, and have to still be there once Terraform has
	// written the one attribute it manages.
	t.Run("leaves the attributes it does not manage untouched", func(t *testing.T) {
		if os.Getenv("TF_ACC") == "" {
			t.Skip("TF_ACC is not set, skipping the acceptance test")
		}

		username := "tfacc-admin-unmanaged"
		renamed := "tfacc-admin-unmanaged-renamed"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckAdministratorDestroyed(username),
				testAccCheckAdministratorDestroyed(renamed),
				testAccCheckSeedAdministratorIntact(t),
			),
			Steps: []resource.TestStep{
				{
					Config: testPritunlAdministratorConfig(username, `password = "tfacc-Passw0rd!"`),
				},
				{
					Config:    testPritunlAdministratorConfig(username, `password = "tfacc-Passw0rd!"`),
					PreConfig: func() { configureAdministratorOutOfBand(t, username) },
					// only the username is managed and only the username
					// changes, the flags set out of band have to survive
					Check: testAccCheckAdministrator(username, map[string]interface{}{
						"super_user": true,
						"auth_api":   true,
						"yubikey_id": "ccccccbcdefg",
					}),
				},
				{
					Config: testPritunlAdministratorConfig(renamed, `password = "tfacc-Passw0rd!"`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "username", renamed),
						testAccCheckAdministrator(renamed, map[string]interface{}{
							"super_user": true,
							"auth_api":   true,
							"yubikey_id": "ccccccbcdefg",
						}),
					),
				},
			},
		})
	})

	// Pritunl answers an account carrying both two-step authentication modes
	// with a 400 admin_invalid_otp. The plan refuses the pair before the
	// request is ever built, which is what turns it into a plan time error.
	t.Run("refuses both two-step authentication modes", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testPritunlAdministratorConfig("tfacc-admin-otp", `
						password       = "tfacc-Passw0rd!"
						otp_auth       = true
						local_otp_auth = true
					`),
					ExpectError: regexp.MustCompile(`otp_auth and local_otp_auth cannot both be enabled`),
					PlanOnly:    true,
				},
			},
		})
	})

	// Only enabling both is refused. Turning both off is an ordinary
	// configuration, and so is turning one on while turning the other off,
	// which is the only way to move an account from one mode to the other, so
	// the plan time check must let both through.
	t.Run("takes the two-step authentication modes it does accept", func(t *testing.T) {
		username := "tfacc-admin-otp-modes"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckAdministratorDestroyed(username),
				testAccCheckSeedAdministratorIntact(t),
			),
			Steps: []resource.TestStep{
				{
					Config: testPritunlAdministratorConfig(username, `
						password       = "tfacc-Passw0rd!"
						otp_auth       = false
						local_otp_auth = false
					`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "otp_auth", "false"),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "local_otp_auth", "false"),
					),
				},
				{
					// one on and the other off, which is what moving an
					// account from one mode to the other looks like
					Config: testPritunlAdministratorConfig(username, `
						password       = "tfacc-Passw0rd!"
						otp_auth       = true
						local_otp_auth = false
					`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "otp_auth", "true"),
						testAccCheckAdministrator(username, map[string]interface{}{"otp_auth": true}),
						testAccCheckAdministratorOtpSecret(username),
					),
				},
				{
					// an account with the two-step authentication turned on
					// carries the shared secret itself in the response, a
					// string, while the request only takes a boolean there:
					// writing anything else to such an account is what an
					// unnormalised round trip fails on, with an empty 400
					Config: testPritunlAdministratorConfig(username, `
						password       = "tfacc-Passw0rd!"
						otp_auth       = true
						local_otp_auth = false
						yubikey_id     = "ccccccbcdefg"
					`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "yubikey_id", "ccccccbcdefg"),
						resource.TestCheckResourceAttr("pritunl_administrator.test", "otp_auth", "true"),
						testAccCheckAdministrator(username, map[string]interface{}{
							"otp_auth":   true,
							"yubikey_id": "ccccccbcdefg",
						}),
						testAccCheckAdministratorOtpSecret(username),
					),
				},
			},
		})
	})

	// The destroy protection of Pritunl, reached without going anywhere near
	// the account the suite authenticates as.
	//
	// The guard counts the super users of the whole instance and only the ones
	// that are enabled, so an account that is both a super user and disabled
	// cannot be removed while the seeded account is the only other super user,
	// even though removing it could not lock anybody out. That makes it the
	// one way to reach the guard with a disposable account of our own: the
	// seeded account is never written to, it merely happens to be the other
	// super user the guard counts. The destroy is expected to fail with the
	// diagnostic this provider builds rather than with a bare 400, and the
	// account is enabled again afterwards so that the framework can clean it
	// up.
	t.Run("explains why the last super user cannot be destroyed", func(t *testing.T) {
		username := "tfacc-admin-protected"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckAdministratorDestroyed(username),
				testAccCheckSeedAdministratorIntact(t),
			),
			Steps: []resource.TestStep{
				{
					Config: testPritunlAdministratorConfig(username, `
						password   = "tfacc-Passw0rd!"
						super_user = true
						disabled   = true
					`),
					Check: testAccCheckAdministrator(username, map[string]interface{}{
						"super_user": true,
						"disabled":   true,
					}),
				},
				{
					Config:      testPritunlAdministratorConfig(username, `password = "tfacc-Passw0rd!"`),
					Destroy:     true,
					ExpectError: regexp.MustCompile(`At least one super administrator must exist`),
				},
				{
					// enabled again, which is what lets the framework destroy
					// it once the case is over
					Config: testPritunlAdministratorConfig(username, `
						password   = "tfacc-Passw0rd!"
						super_user = true
						disabled   = false
					`),
					Check: testAccCheckAdministrator(username, map[string]interface{}{
						"super_user": true,
						"disabled":   false,
					}),
				},
			},
		})
	})

	// Pritunl has no answer of its own for an administrator that does not
	// exist: the handler looks it up, gets nothing back and fails while
	// building the response, which reaches the client as a plain 500 and not
	// as a 404. Telling that apart from an instance in trouble is what keeps
	// an account removed from the web console from failing every later plan
	// instead of being created again.
	t.Run("creates the administrator again when it is removed out of band", func(t *testing.T) {
		if os.Getenv("TF_ACC") == "" {
			t.Skip("TF_ACC is not set, skipping the acceptance test")
		}

		username := "tfacc-admin-removed"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckAdministratorDestroyed(username),
				testAccCheckSeedAdministratorIntact(t),
			),
			Steps: []resource.TestStep{
				{
					Config: testPritunlAdministratorConfig(username, `password = "tfacc-Passw0rd!"`),
				},
				{
					Config:    testPritunlAdministratorConfig(username, `password = "tfacc-Passw0rd!"`),
					PreConfig: func() { deleteAdministratorOutOfBand(t, username) },
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_administrator.test", "username", username),
						testAccCheckAdministrator(username, nil),
					),
				},
			},
		})
	})

	// The seeded administrator is the natural thing to adopt with an import,
	// and the one nobody has the object id of at hand.
	t.Run("adopts an existing administrator by its username", func(t *testing.T) {
		username := "tfacc-admin-imported"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy: resource.ComposeTestCheckFunc(
				testAccCheckAdministratorDestroyed(username),
				testAccCheckSeedAdministratorIntact(t),
			),
			Steps: []resource.TestStep{
				{
					Config: testPritunlAdministratorConfig(username, `password = "tfacc-Passw0rd!"`),
				},
				{
					Config:                  testPritunlAdministratorConfig(username, `password = "tfacc-Passw0rd!"`),
					ResourceName:            "pritunl_administrator.test",
					ImportState:             true,
					ImportStateId:           username,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"password"},
				},
			},
		})
	})

	t.Run("reports an unknown username on import", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config:        testPritunlAdministratorConfig("tfacc-admin-unknown", `password = "tfacc-Passw0rd!"`),
					ResourceName:  "pritunl_administrator.test",
					ImportState:   true,
					ImportStateId: "tfacc-admin-does-not-exist",
					ExpectError:   regexp.MustCompile(`no administrator named "tfacc-admin-does-not-exist" exists`),
				},
			},
		})
	})
}

type administratorCredentials struct {
	token  string
	secret string
}

// findAdministrator looks an administrator up by username through the list
// endpoint, which is the only lookup that answers properly for an account that
// is not there: GET /admin/<id> fails while building the response of an
// unknown id and reaches the client as a 500.
func findAdministrator(username string) (pritunl.Administrator, error) {
	administrators, err := testClient.GetAdministrators()
	if err != nil {
		return nil, err
	}

	for _, administrator := range administrators {
		if administrator.String("username") == username {
			return administrator, nil
		}
	}

	return nil, nil
}

// testAccCheckAdministrator compares the account on the instance against the
// expected fields, reading it from the API rather than from the Terraform
// state so that a resource reporting what it did not write cannot pass. A nil
// expectation stands for a field the API reports as null.
func testAccCheckAdministrator(username string, expected map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		administrator, err := findAdministrator(username)
		if err != nil {
			return err
		}

		if administrator == nil {
			return fmt.Errorf("no administrator named %q exists on the pritunl instance", username)
		}

		for key, value := range expected {
			if fmt.Sprintf("%v", administrator[key]) != fmt.Sprintf("%v", value) {
				return fmt.Errorf("the administrator %q carries %q=%v, want %v",
					username, key, administrator[key], value)
			}
		}

		return nil
	}
}

func testAccCheckAdministratorCredentials(username string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		administrator, err := findAdministrator(username)
		if err != nil {
			return err
		}

		if administrator == nil {
			return fmt.Errorf("no administrator named %q exists on the pritunl instance", username)
		}

		if administrator.String("token") == "" || administrator.String("secret") == "" {
			return fmt.Errorf("the administrator %q has no api credentials", username)
		}

		return nil
	}
}

func testAccRecordAdministratorCredentials(username string, into *administratorCredentials) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		administrator, err := findAdministrator(username)
		if err != nil {
			return err
		}

		if administrator == nil {
			return fmt.Errorf("no administrator named %q exists on the pritunl instance", username)
		}

		into.token = administrator.String("token")
		into.secret = administrator.String("secret")

		return nil
	}
}

func testAccCheckAdministratorCredentialsKept(username string, expected *administratorCredentials) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		administrator, err := findAdministrator(username)
		if err != nil {
			return err
		}

		if administrator == nil {
			return fmt.Errorf("no administrator named %q exists on the pritunl instance", username)
		}

		if administrator.String("token") != expected.token || administrator.String("secret") != expected.secret {
			return fmt.Errorf("the api credentials of the administrator %q have been rotated by an unrelated write",
				username)
		}

		return nil
	}
}

// testAccCheckAdministratorOtpSecret checks the account kept the shared secret
// of its two-step authentication, which Pritunl discards as soon as a request
// turns the mode off and regenerates as soon as one turns it back on.
func testAccCheckAdministratorOtpSecret(username string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		administrator, err := findAdministrator(username)
		if err != nil {
			return err
		}

		if administrator == nil {
			return fmt.Errorf("no administrator named %q exists on the pritunl instance", username)
		}

		if administrator.String("otp_secret") == "" {
			return fmt.Errorf("the administrator %q has no two-step authentication secret", username)
		}

		return nil
	}
}

func testAccCheckAdministratorDestroyed(username string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		administrator, err := findAdministrator(username)
		if err != nil {
			return err
		}

		if administrator != nil {
			return fmt.Errorf("the administrator %q still exists on the pritunl instance", username)
		}

		return nil
	}
}

// configureAdministratorOutOfBand sets a few attributes straight through the
// API, the way an operator would from the web console, on the account the test
// manages. It is a full object write, the only kind the endpoint takes.
func configureAdministratorOutOfBand(t *testing.T, username string) {
	t.Helper()

	administrator, err := findAdministrator(username)
	if err != nil {
		t.Fatalf("failed to read the administrators of the test instance: %s", err)
	}

	if administrator == nil {
		t.Fatalf("no administrator named %q exists on the pritunl instance", username)
	}

	id := administrator.String("id")
	administrator["super_user"] = true
	administrator["auth_api"] = true
	administrator["yubikey_id"] = "ccccccbcdefg"

	if err = testClient.UpdateAdministrator(id, administrator); err != nil {
		t.Fatalf("failed to configure the administrator out of band: %s", err)
	}
}

// deleteAdministratorOutOfBand removes the account straight through the API,
// the way an operator would from the web console.
func deleteAdministratorOutOfBand(t *testing.T, username string) {
	t.Helper()

	administrator, err := findAdministrator(username)
	if err != nil {
		t.Fatalf("failed to read the administrators of the test instance: %s", err)
	}

	if administrator == nil {
		t.Fatalf("no administrator named %q exists on the pritunl instance", username)
	}

	if err = testClient.DeleteAdministrator(administrator.String("id")); err != nil {
		t.Fatalf("failed to delete the administrator out of band: %s", err)
	}
}

// testAccCheckSeedAdministratorIntact fails as soon as the administrator the
// suite authenticates as has moved, whichever test moved it.
//
// It is the safety net of this file. The account holding PRITUNL_TOKEN and
// PRITUNL_SECRET is a super user with API access, exactly what these tests
// create and destroy, and a resource writing to the wrong id would revoke the
// credentials of the test client and of the provider under test in one go,
// leaving every later test unable to authenticate at all. Checking it here
// turns that into a failure of the test that caused it rather than into a
// cascade of unexplained 401s.
//
// It is also why no test here tries to trip the guards Pritunl puts on the
// last super user of the instance. Those count the super users of the whole
// instance, and this account is one of them and always enabled, so a test
// could only reach the guards by demoting, disabling or deleting it, which is
// the one thing that must not happen. The guards themselves are backend
// behaviour and the diagnostics this provider builds out of them are covered
// by TestAdministratorDiagnostics instead.
func testAccCheckSeedAdministratorIntact(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if os.Getenv("TF_ACC") == "" {
			return nil
		}

		token := os.Getenv("PRITUNL_TOKEN")
		secret := os.Getenv("PRITUNL_SECRET")

		administrators, err := testClient.GetAdministrators()
		if err != nil {
			return fmt.Errorf("failed to read the administrators of the test instance: %s", err)
		}

		for _, administrator := range administrators {
			if administrator.String("token") != token {
				continue
			}

			if administrator.String("secret") != secret {
				return fmt.Errorf("the api secret of the administrator the test suite authenticates as has been changed")
			}

			if !administrator.Bool("super_user") {
				return fmt.Errorf("the administrator the test suite authenticates as is no longer a super user")
			}

			if !administrator.Bool("auth_api") {
				return fmt.Errorf("the administrator the test suite authenticates as no longer has api access")
			}

			if administrator.Bool("disabled") {
				return fmt.Errorf("the administrator the test suite authenticates as has been disabled")
			}

			return nil
		}

		return fmt.Errorf("the administrator the test suite authenticates as is gone from the pritunl instance")
	}
}

func testPritunlAdministratorConfig(username, attributes string) string {
	return fmt.Sprintf(`
		resource "pritunl_administrator" "test" {
			username = %[1]q
			%[2]s
		}
	`, username, attributes)
}
