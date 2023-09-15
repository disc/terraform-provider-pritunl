package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccPritunlOrganization(t *testing.T) {

	t.Run("creates organizations without error", func(t *testing.T) {
		orgName := "tfacc-org1"

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("pritunl_organization.test", "name", orgName),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testPritunlOrganizationConfig(orgName),
					Check:  check,
				},
				// import test
				{
					ResourceName:      "pritunl_organization.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func testPritunlOrganizationConfig(name string) string {
	return fmt.Sprintf(`
		resource "pritunl_organization" "test" {
			name    = "%[1]s"
		}
	`, name)
}
