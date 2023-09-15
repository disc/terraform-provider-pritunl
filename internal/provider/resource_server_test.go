package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPritunlServer(t *testing.T) {

	t.Run("creates a server with default configuration", func(t *testing.T) {
		serverName := "tfacc-server1"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testPritunlServerDestroy,
			Steps: []resource.TestStep{
				{
					Config: testPritunlServerSimpleConfig(serverName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_server.test", "name", serverName),
					),
				},
				// import test
				importStep("pritunl_server.test"),
			},
		})
	})

	t.Run("creates a server with sso_auth attribute", func(t *testing.T) {
		serverName := "tfacc-server1"

		testCase := func(t *testing.T, ssoAuth bool) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preCheck(t) },
				ProviderFactories: providerFactories,
				CheckDestroy:      testPritunlServerDestroy,
				Steps: []resource.TestStep{
					{
						Config: testPritunlServerConfigWithSsoAuth(serverName, ssoAuth),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("pritunl_server.test", "name", serverName),
							resource.TestCheckResourceAttr("pritunl_server.test", "sso_auth", strconv.FormatBool(ssoAuth)),
						),
					},
					// import test
					importStep("pritunl_server.test"),
				},
			})
		}

		t.Run("with enabled option", func(t *testing.T) {
			testCase(t, true)
		})

		t.Run("with disabled option", func(t *testing.T) {
			testCase(t, false)
		})

		t.Run("without an option", func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preCheck(t) },
				ProviderFactories: providerFactories,
				CheckDestroy:      testPritunlServerDestroy,
				Steps: []resource.TestStep{
					{
						Config: testPritunlServerSimpleConfig(serverName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("pritunl_server.test", "name", serverName),
							resource.TestCheckResourceAttr("pritunl_server.test", "sso_auth", "false"),
						),
					},
					// import test
					importStep("pritunl_server.test"),
				},
			})
		})
	})

	t.Run("creates a server with an attached organization", func(t *testing.T) {
		serverName := "tfacc-server1"
		orgName := "tfacc-org1"

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testPritunlServerDestroy,
			Steps: []resource.TestStep{
				{
					Config: testPritunlServerConfigWithAttachedOrganization(serverName, orgName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_server.test", "name", serverName),
						resource.TestCheckResourceAttr("pritunl_organization.test", "name", orgName),

						func(s *terraform.State) error {
							attachedOrganizationId := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["organization_ids.0"]
							organizationId := s.RootModule().Resources["pritunl_organization.test"].Primary.Attributes["id"]
							if attachedOrganizationId != organizationId {
								return fmt.Errorf("organization_id is invalid or empty")
							}
							return nil
						},
					),
				},
				// import test
				importStep("pritunl_server.test"),
			},
		})
	})

	t.Run("creates a server with a few attached organizations", func(t *testing.T) {
		serverName := "tfacc-server1"
		org1Name := "tfacc-org1"
		org2Name := "tfacc-org2"

		expectedOrganizationIds := make(map[string]struct{})

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { preCheck(t) },
			ProviderFactories: providerFactories,
			CheckDestroy:      testPritunlServerDestroy,
			Steps: []resource.TestStep{
				{
					Config: testPritunlServerConfigWithAFewAttachedOrganizations(serverName, org1Name, org2Name),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("pritunl_server.test", "name", serverName),
						resource.TestCheckResourceAttr("pritunl_organization.test", "name", org1Name),
						resource.TestCheckResourceAttr("pritunl_organization.test2", "name", org2Name),

						func(s *terraform.State) error {
							attachedOrganization1Id := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["organization_ids.0"]
							attachedOrganization2Id := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["organization_ids.1"]
							organization1Id := s.RootModule().Resources["pritunl_organization.test"].Primary.Attributes["id"]
							organization2Id := s.RootModule().Resources["pritunl_organization.test2"].Primary.Attributes["id"]
							expectedOrganizationIds = map[string]struct{}{
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
					),
				},
				// import test (custom import that ignores order of organization IDs)
				{
					ResourceName: "pritunl_server.test",
					ImportStateCheck: func(states []*terraform.InstanceState) error {
						importedOrganization1Id := states[0].Attributes["organization_ids.0"]
						importedOrganization2Id := states[0].Attributes["organization_ids.1"]

						if _, ok := expectedOrganizationIds[importedOrganization1Id]; !ok {
							return fmt.Errorf("imported organization_id %s doesn't contain in expected organizations list", importedOrganization1Id)
						}

						if _, ok := expectedOrganizationIds[importedOrganization2Id]; !ok {
							return fmt.Errorf("imported organization_id %s doesn't contain in expected organizations list", importedOrganization2Id)
						}

						return nil
					},
					ImportState:       true,
					ImportStateVerify: false,
				},
			},
		})
	})

	t.Run("creates a server with routes", func(t *testing.T) {
		t.Run("with one attached route", func(t *testing.T) {
			serverName := "tfacc-server1"
			routeNetwork := "10.5.0.0/24"
			routeComment := "tfacc-route"

			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preCheck(t) },
				ProviderFactories: providerFactories,
				CheckDestroy:      testPritunlServerDestroy,
				Steps: []resource.TestStep{
					{
						Config: testPritunlServerConfigWithAttachedRoute(serverName, routeNetwork),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("pritunl_server.test", "name", serverName),

							func(s *terraform.State) error {
								actualRouteNetwork := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.0.network"]
								actualRouteComment := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.0.comment"]
								if actualRouteNetwork != routeNetwork {
									return fmt.Errorf("route network is invalid: expected is %s, but actual is %s", routeNetwork, actualRouteNetwork)
								}
								if actualRouteComment != routeComment {
									return fmt.Errorf("route comment is invalid: expected is %s, but actual is %s", routeComment, actualRouteComment)
								}
								return nil
							},
						),
					},
					// import test
					importStep("pritunl_server.test"),
				},
			})
		})

		t.Run("with a few attached routes", func(t *testing.T) {
			serverName := "tfacc-server1"
			route1Network := "10.2.0.0/24"
			route2Network := "10.3.0.0/24"
			route3Network := "10.4.0.0/32"
			routeComment := "tfacc-route"

			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preCheck(t) },
				ProviderFactories: providerFactories,
				CheckDestroy:      testPritunlServerDestroy,
				Steps: []resource.TestStep{
					{
						Config: testPritunlServerConfigWithAFewAttachedRoutes(serverName, route1Network, route2Network, route3Network),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("pritunl_server.test", "name", serverName),

							func(s *terraform.State) error {
								actualRoute1Network := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.0.network"]
								actualRoute2Network := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.1.network"]
								actualRoute3Network := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.2.network"]
								actualRoute1Comment := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.0.comment"]
								actualRoute2Comment := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.1.comment"]
								actualRoute3Comment := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["route.2.comment"]
								if actualRoute1Network != route1Network {
									return fmt.Errorf("first route network is invalid: expected is %s, but actual is %s", route1Network, actualRoute1Network)
								}
								if actualRoute2Network != route2Network {
									return fmt.Errorf("second route network is invalid: expected is %s, but actual is %s", route2Network, actualRoute2Network)
								}
								if actualRoute3Network != route3Network {
									return fmt.Errorf("second route network is invalid: expected is %s, but actual is %s", route3Network, actualRoute3Network)
								}
								if actualRoute1Comment != routeComment {
									return fmt.Errorf("first route comment is invalid: expected is %s, but actual is %s", routeComment, actualRoute1Comment)
								}
								if actualRoute2Comment != routeComment {
									return fmt.Errorf("second route comment is invalid: expected is %s, but actual is %s", routeComment, actualRoute2Comment)
								}
								if actualRoute3Comment != routeComment {
									return fmt.Errorf(" route comment is invalid: expected is %s, but actual is %s", routeComment, actualRoute3Comment)
								}
								return nil
							},
						),
					},
					// import test
					importStep("pritunl_server.test"),
				},
			})
		})
	})

	t.Run("creates a server with error", func(t *testing.T) {
		t.Run("due to an invalid network", func(t *testing.T) {
			serverName := "tfacc-server1"
			port := 11111
			missedSubnetNetwork := "10.100.0.2"
			invalidNetwork := "10.100.0"

			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preCheck(t) },
				ProviderFactories: providerFactories,
				CheckDestroy:      testPritunlServerDestroy,
				Steps: []resource.TestStep{
					{
						Config:      testGetServerConfigWithNetworkAndPort(serverName, missedSubnetNetwork, port),
						ExpectError: regexp.MustCompile(fmt.Sprintf("invalid CIDR address: %s", missedSubnetNetwork)),
					},
					{
						Config:      testGetServerConfigWithNetworkAndPort(serverName, invalidNetwork, port),
						ExpectError: regexp.MustCompile(fmt.Sprintf("invalid CIDR address: %s", invalidNetwork)),
					},
				},
			})
		})

		t.Run("due to an unsupported network", func(t *testing.T) {
			serverName := "tfacc-server1"
			port := 11111
			unsupportedNetwork := "172.14.68.0/24"
			supportedNetwork := "172.16.68.0/24"

			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preCheck(t) },
				ProviderFactories: providerFactories,
				CheckDestroy:      testPritunlServerDestroy,
				Steps: []resource.TestStep{
					{
						Config:      testGetServerConfigWithNetworkAndPort(serverName, unsupportedNetwork, port),
						ExpectError: regexp.MustCompile(fmt.Sprintf("provided subnet %s does not belong to expected subnets 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16", unsupportedNetwork)),
					},
					{
						Config: testGetServerConfigWithNetworkAndPort(serverName, supportedNetwork, port),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("pritunl_server.test", "name", serverName),
							resource.TestCheckResourceAttr("pritunl_server.test", "network", supportedNetwork),
						),
					},
				},
			})
		})

		t.Run("due to an invalid route", func(t *testing.T) {
			serverName := "tfacc-server1"
			invalidRouteNetwork := "10.100.0.2"

			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preCheck(t) },
				ProviderFactories: providerFactories,
				CheckDestroy:      testPritunlServerDestroy,
				Steps: []resource.TestStep{
					{
						Config:      testPritunlServerConfigWithAttachedRoute(serverName, invalidRouteNetwork),
						ExpectError: regexp.MustCompile(fmt.Sprintf("invalid CIDR address: %s", invalidRouteNetwork)),
					},
				},
			})
		})

		t.Run("due to an invalid bind_address", func(t *testing.T) {
			serverName := "tfacc-server1"
			network := "172.16.68.0/24"
			port := 11111
			invalidBindAddress := "10.100.0.1/24"
			correctBindAddress := "10.100.0.1"

			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preCheck(t) },
				ProviderFactories: providerFactories,
				CheckDestroy:      testPritunlServerDestroy,
				Steps: []resource.TestStep{
					{
						Config:      testGetServerConfigWithBindAddress(serverName, network, invalidBindAddress, port),
						ExpectError: regexp.MustCompile(fmt.Sprintf("expected bind_address to contain a valid IP, got: %s", invalidBindAddress)),
					},
					{
						Config: testGetServerConfigWithBindAddress(serverName, network, correctBindAddress, port),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("pritunl_server.test", "name", serverName),
							resource.TestCheckResourceAttr("pritunl_server.test", "bind_address", correctBindAddress),
						),
					},
				},
			})
		})
	})

	t.Run("creates a server with groups attribute", func(t *testing.T) {
		serverName := "tfacc-server1"

		t.Run("with correct group name", func(t *testing.T) {
			correctGroupName := "Group-Has-No-Spaces"
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preCheck(t) },
				ProviderFactories: providerFactories,
				CheckDestroy:      testPritunlServerDestroy,
				Steps: []resource.TestStep{
					{
						Config: testPritunlServerConfigWithGroups(serverName, correctGroupName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("pritunl_server.test", "name", serverName),

							func(s *terraform.State) error {
								groupName := s.RootModule().Resources["pritunl_server.test"].Primary.Attributes["groups.0"]
								if groupName != correctGroupName {
									return fmt.Errorf("group name mismatch")
								}

								return nil
							},
						),
					},
					// import test
					importStep("pritunl_server.test"),
				},
			})
		})

		t.Run("with invalid group name", func(t *testing.T) {
			invalidGroupName := "Group Name With Spaces"
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { preCheck(t) },
				ProviderFactories: providerFactories,
				CheckDestroy:      testPritunlServerDestroy,
				Steps: []resource.TestStep{
					{
						Config:      testPritunlServerConfigWithGroups(serverName, invalidGroupName),
						ExpectError: regexp.MustCompile(fmt.Sprintf("%s contains spaces", invalidGroupName)),
					},
				},
			})
		})
	})
}

func testPritunlServerSimpleConfig(name string) string {
	return fmt.Sprintf(`
		resource "pritunl_server" "test" {
			name    = "%[1]s"
		}
	`, name)
}

func testPritunlServerConfigWithSsoAuth(name string, ssoAuth bool) string {
	return fmt.Sprintf(`
		resource "pritunl_server" "test" {
			name     = "%[1]s"
			sso_auth = %[2]v
		}
	`, name, ssoAuth)
}

func testPritunlServerConfigWithAttachedOrganization(name, organizationName string) string {
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

func testPritunlServerConfigWithAFewAttachedOrganizations(name, organization1Name, organization2Name string) string {
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

func testPritunlServerConfigWithAttachedRoute(name, route string) string {
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

func testPritunlServerConfigWithAFewAttachedRoutes(name, route1, route2, route3 string) string {
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

func testGetServerConfigWithNetworkAndPort(name, network string, port int) string {
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

func testPritunlServerConfigWithGroups(name string, groupName string) string {
	return fmt.Sprintf(`
		resource "pritunl_server" "test" {
			name    = "%[1]s"
			groups    = ["%[2]s"]
		}
	`, name, groupName)
}

func testPritunlServerDestroy(s *terraform.State) error {
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
