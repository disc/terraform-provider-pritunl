package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccPritunlHost(t *testing.T) {
	// pritunl.local sets in Makefile's "test" target
	existsHostname := "pritunl.local"
	notExistHostname := "not-exist-hostname"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() {},
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testPritunlHostSimpleConfig(existsHostname),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config:      testPritunlHostSimpleConfig(notExistHostname),
				ExpectError: regexp.MustCompile(fmt.Sprintf("could not a find host with a hostname %s. Previous error message: could not find a host with specified parameters", notExistHostname)),
			},
		},
	})
}

func testPritunlHostSimpleConfig(name string) string {
	return fmt.Sprintf(`
		data "pritunl_host" "test" {
			hostname    = "%[1]s"
		}
	`, name)
}
