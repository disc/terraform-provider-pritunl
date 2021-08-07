package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"
)

func TestAccOrganization_basic(t *testing.T) {
	var orgId string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig("tfacc-org1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_organization.test", "name", "tfacc-org1"),

					// extract siteName for future use
					func(s *terraform.State) error {
						orgId = s.RootModule().Resources["pritunl_organization.test"].Primary.Attributes["id"]
						return nil
					},
				),
			},
			importStep("pritunl_organization.test"),
			{
				Config: testAccOrganizationConfig("tfacc-org2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_organization.test", "name", "tfacc-org2"),
				),
			},
			importStep("pritunl_organization.test"),
			// test importing
			{
				ResourceName: "pritunl_organization.test",
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return orgId, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccOrganizationConfig(name string) string {
	return fmt.Sprintf(`
resource "pritunl_organization" "test" {
	name    = "%[1]s"
}
`, name)
}

func testAccOrganizationDestroy(s *terraform.State) error {
	organizations, err := testClient.GetOrganizations()
	if err != nil {
		return err
	}
	for _, org := range organizations {
		if strings.HasPrefix(org.Name, "tfacc-") {
			return fmt.Errorf("organization is not destroyed")
		}
	}
	return nil
}
