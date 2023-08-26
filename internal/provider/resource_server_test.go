package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestGetServer_basic(t *testing.T) {
	var serverId string

	resource.Test(t, resource.TestCase{
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
				Config: testGetServerConfig("tfacc-server2", "10.4.0.0/24", 12345),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server2"),
					resource.TestCheckResourceAttr("pritunl_server.test", "network", "10.4.0.0/24"),
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

	resource.Test(t, resource.TestCase{
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

func TestGetServer_with_a_few_attached_organizations(t *testing.T) {
	var serverId string

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testGetServerSimpleConfigWithAFewAttachedOrganization("tfacc-server1", "tfacc-org1", "tfacc-org2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server1"),
					resource.TestCheckResourceAttr("pritunl_organization.test", "name", "tfacc-org1"),
					resource.TestCheckResourceAttr("pritunl_organization.test2", "name", "tfacc-org2"),

					func(s *terraform.State) error {
						attachedOrganization1Id := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["organization_ids.0"]
						attachedOrganization2Id := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["organization_ids.1"]
						organization1Id := s.RootModule().Resources["pritunl_organization.test"].Primary.Attributes["id"]
						organization2Id := s.RootModule().Resources["pritunl_organization.test2"].Primary.Attributes["id"]
						expectedOrganizationIds := map[string]struct{}{
							organization1Id: {},
							organization2Id: {},
						}

						if attachedOrganization1Id == attachedOrganization2Id {
							return fmt.Errorf("first and seconds attached organization_id is the same")
						}

						if _, ok := expectedOrganizationIds[attachedOrganization1Id]; !ok {
							return fmt.Errorf("attached organization_id %s doesn't contain in expected organizations list", attachedOrganization1Id)
						}

						if _, ok := expectedOrganizationIds[attachedOrganization2Id]; !ok {
							return fmt.Errorf("attached organization_id %s doesn't contain in expected organizations list", attachedOrganization2Id)
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

	expectedRouteNetwork := "10.5.0.0/24"
	expectedRouteComment := "tfacc-route"

	resource.Test(t, resource.TestCase{
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

func TestGetServer_with_a_few_attached_routes(t *testing.T) {
	var serverId string

	expectedRoute1Network := "10.2.0.0/24"
	expectedRoute2Network := "10.3.0.0/24"
	expectedRoute3Network := "10.4.0.0/32"
	expectedRouteComment := "tfacc-route"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testGetServerSimpleConfigWithAFewAttachedRoutes("tfacc-server1", expectedRoute1Network, expectedRoute2Network, expectedRoute3Network),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server1"),

					func(s *terraform.State) error {
						routeNetwork1 := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.0.network"]
						routeNetwork2 := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.1.network"]
						routeNetwork3 := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.2.network"]
						routeComment1 := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.0.comment"]
						routeComment2 := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.1.comment"]
						routeComment3 := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.2.comment"]
						if routeNetwork1 != expectedRoute1Network {
							return fmt.Errorf("first route network is invalid: expected is %s, but actual is %s", expectedRoute1Network, routeNetwork1)
						}
						if routeNetwork2 != expectedRoute2Network {
							return fmt.Errorf("second route network is invalid: expected is %s, but actual is %s", expectedRoute2Network, routeNetwork2)
						}
						if routeNetwork3 != expectedRoute3Network {
							return fmt.Errorf("second route network is invalid: expected is %s, but actual is %s", expectedRoute3Network, routeNetwork3)
						}
						if routeComment1 != expectedRouteComment {
							return fmt.Errorf("first route comment is invalid: expected is %s, but actual is %s", expectedRouteComment, routeComment1)
						}
						if routeComment2 != expectedRouteComment {
							return fmt.Errorf("second route comment is invalid: expected is %s, but actual is %s", expectedRouteComment, routeComment2)
						}
						if routeComment3 != expectedRouteComment {
							return fmt.Errorf(" route comment is invalid: expected is %s, but actual is %s", expectedRouteComment, routeComment3)
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

func TestGetServer_with_invalid_route(t *testing.T) {
	invalidRouteNetwork := "10.100.0.2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testGetServerSimpleConfigWithAttachedRoute("tfacc-server1", invalidRouteNetwork),
				ExpectError: regexp.MustCompile(fmt.Sprintf("invalid CIDR address: %s", invalidRouteNetwork)),
			},
		},
	})
}

func TestCreateServer_with_invalid_network(t *testing.T) {
	missedSubnetNetwork := "10.100.0.2"
	invalidNetwork := "10.100.0"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testGetServerConfig("tfacc-server1", missedSubnetNetwork, 11111),
				ExpectError: regexp.MustCompile(fmt.Sprintf("invalid CIDR address: %s", missedSubnetNetwork)),
			},
			{
				Config:      testGetServerConfig("tfacc-server2", invalidNetwork, 22222),
				ExpectError: regexp.MustCompile(fmt.Sprintf("invalid CIDR address: %s", invalidNetwork)),
			},
		},
	})
}

func TestCreateServer_with_unsupported_network(t *testing.T) {
	unsupportedNetwork := "172.14.68.0/24"
	supportedNetwork := "172.16.68.0/24"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testGetServerConfig("tfacc-server1", unsupportedNetwork, 11111),
				ExpectError: regexp.MustCompile(fmt.Sprintf("provided subnet %s does not belong to expected subnets 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16", unsupportedNetwork)),
			},
			{
				Config: testGetServerConfig("tfacc-server2", supportedNetwork, 22222),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server2"),
					resource.TestCheckResourceAttr("pritunl_server.test", "network", supportedNetwork),
				),
			},
		},
	})
}

func TestCreateServer_with_invalid_bind_address(t *testing.T) {
	invalidBindAddress := "10.100.0.1/24"
	correctBindAddress := "10.100.0.1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testGetServerConfigWithBindAddress("tfacc-server1", "172.16.68.0/24", invalidBindAddress, 11111),
				ExpectError: regexp.MustCompile(fmt.Sprintf("expected bind_address to contain a valid IP, got: %s", invalidBindAddress)),
			},
			{
				Config: testGetServerConfigWithBindAddress("tfacc-server2", "172.16.68.0/24", correctBindAddress, 22222),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server2"),
					resource.TestCheckResourceAttr("pritunl_server.test", "bind_address", correctBindAddress),
				),
			},
		},
	})
}

func TestGetServer_with_default_host(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testGetServerSimpleConfig("tfacc-server1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server1"),

					func(s *terraform.State) error {
						attachedHostId := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["host_ids.0"]
						if attachedHostId == "" {
							return fmt.Errorf("attached host is empty")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestGetServer_without_hosts(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testGetServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testGetServerSimpleConfig("tfacc-server1"),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						attachedHost := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["host_ids.0"]
						if attachedHost == "" {
							return fmt.Errorf("attached hosts must not be empty")
						}
						return nil
					},
				),
			},
			{
				Config: testGetServerSimpleConfigWithoutHosts("tfacc-server1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pritunl_server.test", "name", "tfacc-server1"),

					func(s *terraform.State) error {
						attachedHost := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["host_ids.0"]
						if attachedHost != "" {
							return fmt.Errorf("attached hosts must be empty")
						}

						return nil
					},
				),
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

func testGetServerSimpleConfigWithAFewAttachedOrganization(name, organization1Name, organization2Name string) string {
	//testing net-gateway
	return fmt.Sprintf(`
resource "pritunl_organization" "test" {
	name    = "%[2]s"
}

resource "pritunl_organization" "test2" {
	name    = "%[3]s"
}

resource "pritunl_server" "test" {
	name    = "%[1]s"
	organization_ids = [
		pritunl_organization.test.id,
		pritunl_organization.test2.id
	]
}
`, name, organization1Name, organization2Name)
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

func testGetServerSimpleConfigWithAFewAttachedRoutes(name, route1, route2, route3 string) string {

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

	route {
		network = "%[4]s"
		comment = "tfacc-route"
		net_gateway = true
  	}	
}
`, name, route1, route2, route3)
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

func testGetServerConfigWithBindAddress(name, network, bindAddress string, port int) string {
	return fmt.Sprintf(`
resource "pritunl_server" "test" {
	name    		= "%[1]s"
    network  		= "%[2]s"
    bind_address  	= "%[3]s"
    port     		= %[4]d
    protocol 		= "tcp"
}
`, name, network, bindAddress, port)
}

func testGetServerSimpleConfigWithoutHosts(name string) string {
	return fmt.Sprintf(`
resource "pritunl_server" "test" {
	name    = "%[1]s"

	host_ids = []
}
`, name)
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
