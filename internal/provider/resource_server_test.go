package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"
)

func TestGetServer_basic(t *testing.T) {
	var serverId string
	_ = serverId

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testGetServerSimpleConfig("tfacc-server1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server1"),

					// extract siteName for future use
					func(s *terraform.State) error {
						serverId = s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["id"]
						return nil
					},
				),
			},
			importStep("pritunl_server.test"),
			{
				Config: testGetServerConfig("tfacc-server2", "192.168.10.0/24", 12345),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server2"),
					resource.TestCheckResourceAttr("pritunl_server.test", "network", "192.168.10.0/24"),
					resource.TestCheckResourceAttr("pritunl_server.test", "port", "12345"),
					resource.TestCheckResourceAttr("pritunl_server.test", "protocol", "tcp"),
				),
			},
			importStep("pritunl_server.test"),
			// test importing
			{
				ResourceName: "pritunl_server.test",
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return serverId, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testGetServerSimpleConfig(name string) string {
	return fmt.Sprintf(`
resource "pritunl_server" "test" {
	name    = "%[1]s"
}
`, name)
}

func testGetServerConfig(name, network string, port int) string {
	return fmt.Sprintf(`
resource "pritunl_server" "test" {
	name    = "%[1]s"
    network  =  "%[2]s"
    port     = %[3]d
    protocol = "tcp"
}
`, name, network, port)
}

func testGetServerDestroy(_ *terraform.State) error {
	servers, err := testClient.GetServers()
	if err != nil {
		return err
	}
	for _, server := range servers {
		if strings.HasPrefix(server.Name, "tfacc-") {
			return fmt.Errorf("a server is not destroyed")
		}
	}
	return nil
}
