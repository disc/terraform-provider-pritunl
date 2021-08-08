package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestGetServer_basic(t *testing.T) {
	var serverId string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testGetServerSimpleConfig("tfacc-server1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server1"),

					// extract serverId for future use
					func(s *terraform.State) error {
						serverId = s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["id"]
						return nil
					},
				),
			},
			importStep("pritunl_server.test"),
			{
				Config: testGetServerConfig("tfacc-server2", "10.0.0.0/24", 12345),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server2"),
					resource.TestCheckResourceAttr("pritunl_server.test", "network", "10.0.0.0/24"),
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

func TestGetServer_with_attached_organization(t *testing.T) {
	var serverId string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testGetServerSimpleConfigWithAttachedOrganization("tfacc-server1", "tfacc-org1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server1"),
					resource.TestCheckResourceAttr("pritunl_organization.test", "name", "tfacc-org1"),

					func(s *terraform.State) error {
						attachedOrganizationId := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["organization_ids.0"]
						organizationId := s.RootModule().Resources["pritunl_organization.test"].Primary.Attributes["id"]
						if attachedOrganizationId != organizationId {
							return fmt.Errorf("organization_id is invalid or empty")
						}
						return nil
					},

					// extract serverId for future use
					func(s *terraform.State) error {
						serverId = s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["id"]
						return nil
					},
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

func TestGetServer_with_attached_route(t *testing.T) {
	var serverId string

	expectedRouteNetwork := "10.1.0.0/24"
	expectedRouteComment := "tfacc-route"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testGetServerSimpleConfigWithAttachedRoute("tfacc-server1", expectedRouteNetwork),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server1"),

					func(s *terraform.State) error {
						routeNetwork := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.0.network"]
						routeComment := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.0.comment"]
						if routeNetwork != expectedRouteNetwork {
							return fmt.Errorf("route network is invalid: expected is %s, but actual is %s", expectedRouteNetwork, routeNetwork)
						}
						if routeComment != expectedRouteComment {
							return fmt.Errorf("route comment is invalid: expected is %s, but actual is %s", expectedRouteComment, routeComment)
						}
						return nil
					},

					// extract serverId for future use
					func(s *terraform.State) error {
						serverId = s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["id"]
						return nil
					},
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

func TestGetServer_with_a_few_attached_route(t *testing.T) {
	var serverId string

	expectedRoute1Network := "10.1.0.0/24"
	expectedRoute2Network := "10.2.0.0/24"
	expectedRouteComment := "tfacc-route"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testGetServerSimpleConfigWithAFewAttachedRoutes("tfacc-server1", expectedRoute1Network, expectedRoute2Network),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server1"),

					func(s *terraform.State) error {
						routeNetwork1 := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.0.network"]
						routeNetwork2 := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.1.network"]
						routeComment1 := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.0.comment"]
						routeComment2 := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.1.comment"]
						if routeNetwork1 != expectedRoute1Network {
							return fmt.Errorf("first route network is invalid: expected is %s, but actual is %s", expectedRoute1Network, routeNetwork1)
						}
						if routeNetwork2 != expectedRoute2Network {
							return fmt.Errorf("second route network is invalid: expected is %s, but actual is %s", expectedRoute2Network, routeNetwork1)
						}
						if routeComment1 != expectedRouteComment {
							return fmt.Errorf("first route comment is invalid: expected is %s, but actual is %s", expectedRouteComment, routeComment1)
						}
						if routeComment2 != expectedRouteComment {
							return fmt.Errorf("second route comment is invalid: expected is %s, but actual is %s", expectedRouteComment, routeComment2)
						}
						return nil
					},

					// extract serverId for future use
					func(s *terraform.State) error {
						serverId = s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["id"]
						return nil
					},
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

func testGetServerSimpleConfigWithAttachedOrganization(name, organizationName string) string {
	return fmt.Sprintf(`
resource "pritunl_organization" "test" {
	name    = "%[2]s"
}

resource "pritunl_server" "test" {
	name    = "%[1]s"
	organization_ids = [
		pritunl_organization.test.id
	]
}
`, name, organizationName)
}

func testGetServerSimpleConfigWithAttachedRoute(name, route string) string {
	return fmt.Sprintf(`
resource "pritunl_server" "test" {
	name = "%[1]s"
	
	route {
		network = "%[2]s"
		comment = "tfacc-route"
  	}
}
`, name, route)
}
func testGetServerSimpleConfigWithAFewAttachedRoutes(name, route1, route2 string) string {
	return fmt.Sprintf(`
resource "pritunl_server" "test" {
	name = "%[1]s"
	
	route {
		network = "%[2]s"
		comment = "tfacc-route"
  	}

	route {
		network = "%[3]s"
		comment = "tfacc-route"
  	}
}
`, name, route1, route2)
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

func testGetServerDestroy(s *terraform.State) error {
	serverId := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["id"]

	servers, err := testClient.GetServers()
	if err != nil {
		return err
	}
	for _, server := range servers {
		if server.ID == serverId {
			return fmt.Errorf("a server is not destroyed")
		}
	}
	return nil
}
