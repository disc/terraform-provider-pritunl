package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestDataSourceHost(t *testing.T) {
	// pritunl.local sets in Makefile's "test" target
	existsHostname := "pritunl.local"
	notExistHostname := "not-exist-hostname"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() {},
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataHostSimpleConfig(existsHostname),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config:      testDataHostSimpleConfig(notExistHostname),
				ExpectError: regexp.MustCompile(fmt.Sprintf("could not find host with a hostname %s. Previous error message: could not find a host with specified parameters", notExistHostname)),
			},
		},
	})
}

func testDataHostSimpleConfig(name string) string {
	return fmt.Sprintf(`
data "pritunl_host" "test" {
	hostname    = "%[1]s"
}
`, name)
}
