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
			resource.TestCheckNoResourceAttr("pritunl_user.test", "pin"),
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
	t.Run("creates users with PIN without error", func(t *testing.T) {
		username := "tfacc-user2"
		orgName := "tfacc-org2"
		pin := "123456"

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("pritunl_user.test", "name", username),
			resource.TestCheckResourceAttr("pritunl_user.test", "pin", pin),
			resource.TestCheckResourceAttr("pritunl_organization.test", "name", orgName),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testPritunlUserConfigWithPin(username, orgName, pin),
					Check:  check,
				},
				// import test
				pritunlUserImportStep("pritunl_user.test"),
			},
		})
	})
}

func testPritunlUserConfig(username, orgName string) string {
	return testPritunlUserConfigWithPin(username, orgName, "")
}

func testPritunlUserConfigWithPin(username, orgName, pin string) string {
	resources := fmt.Sprintf(`
resource "pritunl_organization" "test" {
    name = "%[2]s"
}

resource "pritunl_user" "test" {
    name = "%[1]s"
    organization_id = pritunl_organization.test.id
    `, username, orgName)

	if pin != "" {
		resources += fmt.Sprintf("pin = \"%[1]s\"\n", pin)
	}

	resources += "}\n"

	return resources
}
