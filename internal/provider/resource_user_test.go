package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				// import test
				pritunlUserImportStep("pritunl_user.test"),
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
