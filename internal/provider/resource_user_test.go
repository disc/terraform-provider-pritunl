package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccUser_basic(t *testing.T) {
	var userId, orgId string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		//CheckDestroy:      testAccUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig("tfacc-user1", "tfacc-org1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_user.test", "name", "tfacc-user1"),
					resource.TestCheckResourceAttr("pritunl_organization.test", "name", "tfacc-org1"),

					// extract siteName for future use
					func(s *terraform.State) error {
						userId = s.RootModule().Resources["pritunl_user.test"].Primary.Attributes["id"]
						orgId = s.RootModule().Resources["pritunl_organization.test"].Primary.Attributes["id"]
						return nil
					},
				),
			},
			userImportStep("pritunl_user.test"),
			{
				Config: testAccUserConfig("tfacc-user2", "tfacc-org1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_user.test", "name", "tfacc-user2"),
				),
			},
			userImportStep("pritunl_user.test"),
			// test importing
			{
				ResourceName: "pritunl_user.test",
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return fmt.Sprintf("%s-%s", orgId, userId), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserConfig(username, orgName string) string {
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
