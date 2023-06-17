package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestDataSourceHosts(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() {},
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataHostsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("num_hosts", "1"),
				),
			},
		},
	})
}

func testDataHostsConfig() string {
	return fmt.Sprintf(`
data "pritunl_hosts" "my-server-hosts" {}

output "num_hosts" {
  value = length(data.pritunl_hosts.my-server-hosts.hosts)
}
`)
}
