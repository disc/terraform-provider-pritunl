package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"
)

func TestAccPritunlUser(t *testing.T) {

	t.Run("creates users without error", func(t *testing.T) {
		username := "tfacc-user1"
		orgName := "tfacc-org1"

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("pritunl_user.test", "name", username),
			resource.TestCheckResourceAttr("pritunl_organization.test", "name", orgName),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testPritunlUserConfig(username, orgName),
					Check:  check,
				},
			},
		})
	})

	t.Run("updates users without error", func(t *testing.T) {
		username := "tfacc-user1"
		orgName := "tfacc-org1"

		newUsername := "tfacc-user1-new"

		initialConfig := testPritunlUserConfig(username, orgName)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("pritunl_user.test", "name", username),
				resource.TestCheckResourceAttr("pritunl_organization.test", "name", orgName),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("pritunl_user.test", "name", newUsername),
				resource.TestCheckResourceAttr("pritunl_organization.test", "name", orgName),
			),
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: initialConfig,
					Check:  checks["before"],
				},
				{
					Config: strings.Replace(initialConfig, username, newUsername, 1),
					Check:  checks["after"],
				},
			},
		})
	})

	t.Run("imports users without error", func(t *testing.T) {
		username := "tfacc-user1"
		orgName := "tfacc-org1"

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("pritunl_user.test", "name", username),
			resource.TestCheckResourceAttr("pritunl_organization.test", "name", orgName),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testPritunlUserConfig(username, orgName),
					Check:  check,
				},
				{
					ResourceName:      "pritunl_user.test",
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateIdFunc: func(state *terraform.State) (string, error) {
						userId := state.RootModule().Resources["pritunl_user.test"].Primary.Attributes["id"]
						orgId := state.RootModule().Resources["pritunl_organization.test"].Primary.Attributes["id"]

						return fmt.Sprintf("%s-%s", orgId, userId), nil
					},
				},
			},
		})
	})
}

func testPritunlUserConfig(username, orgName string) string {
	return fmt.Sprintf(`
		resource "pritunl_organization" "test" {
			name = "%[2]s"
		}

		resource "pritunl_user" "test" {
			name    			= "%[1]s"
			organization_id		= pritunl_organization.test.id
		}
	`, username, orgName)
}
