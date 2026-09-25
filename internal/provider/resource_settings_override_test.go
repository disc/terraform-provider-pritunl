package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPritunlSettingsOverride(t *testing.T) {

	t.Run("manages basic settings without error", func(t *testing.T) {
		username := "tfacc-admin"
		theme := "dark"
		pinMode := "optional"
		reverseProxy := true

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testSettingsOverrideBasicConfig(username, theme, pinMode, reverseProxy),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "username", username),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "theme", theme),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "pin_mode", pinMode),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "reverse_proxy", strconv.FormatBool(reverseProxy)),
					),
				},
			},
		})
	})

	t.Run("manages email settings without error", func(t *testing.T) {
		emailFrom := "tfacc-noreply@example.com"
		emailServer := "smtp.example.com"
		emailUsername := "tfacc-user"
		emailPassword := "tfacc-password"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testSettingsOverrideEmailConfig(emailFrom, emailServer, emailUsername, emailPassword),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "email_from", emailFrom),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "email_server", emailServer),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "email_username", emailUsername),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "email_password", emailPassword),
					),
				},
			},
		})
	})

	t.Run("manages influxdb settings without error", func(t *testing.T) {
		influxdbUrl := "http://tfacc-influxdb.example.com:8086"
		influxdbToken := "tfacc-token"
		influxdbOrg := "tfacc-org"
		influxdbBucket := "tfacc-bucket"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testSettingsOverrideInfluxdbConfig(influxdbUrl, influxdbToken, influxdbOrg, influxdbBucket),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "influxdb_url", influxdbUrl),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "influxdb_org", influxdbOrg),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "influxdb_bucket", influxdbBucket),
					),
				},
			},
		})
	})

	t.Run("manages network settings without error", func(t *testing.T) {
		publicAddress := "10.0.0.1"
		publicAddress6 := "::1"
		routedSubnet6 := "fd00::/64"
		routedSubnet6Wg := "fd01::/64"
		ipv6 := true

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testSettingsOverrideNetworkConfig(publicAddress, publicAddress6, routedSubnet6, routedSubnet6Wg, ipv6),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "public_address", publicAddress),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "public_address6", publicAddress6),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "routed_subnet6", routedSubnet6),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "routed_subnet6_wg", routedSubnet6Wg),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "ipv6", strconv.FormatBool(ipv6)),
					),
				},
			},
		})
	})

	t.Run("manages user restriction settings without error", func(t *testing.T) {
		restrictImport := true
		clientReconnect := false
		dropPermissions := true
		restrictClient := true

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testSettingsOverrideUserConfig(restrictImport, clientReconnect, dropPermissions, restrictClient),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "restrict_import", strconv.FormatBool(restrictImport)),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "client_reconnect", strconv.FormatBool(clientReconnect)),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "drop_permissions", strconv.FormatBool(dropPermissions)),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "restrict_client", strconv.FormatBool(restrictClient)),
					),
				},
			},
		})
	})

	t.Run("imports settings without error", func(t *testing.T) {
		username := "tfacc-admin"
		theme := "dark"
		pinMode := "optional"
		reverseProxy := true

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testSettingsOverrideBasicConfig(username, theme, pinMode, reverseProxy),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "username", username),
					),
				},
				importStep("pritunl_settings_override.test", "email_password", "influxdb_token", "server_key"),
			},
		})
	})

	t.Run("updates settings without error", func(t *testing.T) {
		username1 := "tfacc-admin1"
		theme1 := "dark"
		pinMode1 := "optional"
		reverseProxy1 := true

		username2 := "tfacc-admin2"
		theme2 := "light"
		pinMode2 := "disabled"
		reverseProxy2 := false

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testSettingsOverrideBasicConfig(username1, theme1, pinMode1, reverseProxy1),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "username", username1),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "theme", theme1),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "pin_mode", pinMode1),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "reverse_proxy", strconv.FormatBool(reverseProxy1)),
					),
				},
				{
					Config: testSettingsOverrideBasicConfig(username2, theme2, pinMode2, reverseProxy2),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "username", username2),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "theme", theme2),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "pin_mode", pinMode2),
						resource.TestCheckResourceAttr("pritunl_settings_override.test", "reverse_proxy", strconv.FormatBool(reverseProxy2)),
					),
				},
			},
		})
	})
}

func testSettingsOverrideBasicConfig(username, theme, pinMode string, reverseProxy bool) string {
	return fmt.Sprintf(`
		resource "pritunl_settings_override" "test" {
			username      = "%[1]s"
			theme         = "%[2]s"
			pin_mode      = "%[3]s"
			reverse_proxy = %[4]v
		}
	`, username, theme, pinMode, reverseProxy)
}

func testSettingsOverrideEmailConfig(emailFrom, emailServer, emailUsername, emailPassword string) string {
	return fmt.Sprintf(`
		resource "pritunl_settings_override" "test" {
			email_from     = "%[1]s"
			email_server   = "%[2]s"
			email_username = "%[3]s"
			email_password = "%[4]s"
		}
	`, emailFrom, emailServer, emailUsername, emailPassword)
}

func testSettingsOverrideInfluxdbConfig(influxdbUrl, influxdbToken, influxdbOrg, influxdbBucket string) string {
	return fmt.Sprintf(`
		resource "pritunl_settings_override" "test" {
			influxdb_url    = "%[1]s"
			influxdb_token  = "%[2]s"
			influxdb_org    = "%[3]s"
			influxdb_bucket = "%[4]s"
		}
	`, influxdbUrl, influxdbToken, influxdbOrg, influxdbBucket)
}

func testSettingsOverrideNetworkConfig(publicAddress, publicAddress6, routedSubnet6, routedSubnet6Wg string, ipv6 bool) string {
	return fmt.Sprintf(`
		resource "pritunl_settings_override" "test" {
			public_address   = "%[1]s"
			public_address6  = "%[2]s"
			routed_subnet6   = "%[3]s"
			routed_subnet6_wg = "%[4]s"
			ipv6             = %[5]v
		}
	`, publicAddress, publicAddress6, routedSubnet6, routedSubnet6Wg, ipv6)
}

func testSettingsOverrideUserConfig(restrictImport, clientReconnect, dropPermissions, restrictClient bool) string {
	return fmt.Sprintf(`
		resource "pritunl_settings_override" "test" {
			restrict_import  = %[1]v
			client_reconnect = %[2]v
			drop_permissions = %[3]v
			restrict_client  = %[4]v
		}
	`, restrictImport, clientReconnect, dropPermissions, restrictClient)
}
