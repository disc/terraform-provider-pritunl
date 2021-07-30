package provider

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"pritunl-terraform/internal/pritunl"
	"regexp"
	"testing"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"pritunl": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("SERVICE_URL"); v == "" {
		t.Fatal("SERVICE_URL must be set for acceptance tests")
	}
	if v := os.Getenv("SERVICE_TOKEN"); v == "" {
		t.Fatal("SERVICE_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("SERVICE_SECRET"); v == "" {
		t.Fatal("SERVICE_SECRET must be set for acceptance tests")
	}
}

func TestAccOrganization_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckItemBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExampleItemExists("pritunl_org.test_org"),
					resource.TestCheckResourceAttr("pritunl_org.test_org", "name", "test_org"),
				),
			},
		},
	})
}

func testAccCheckExampleItemExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		name := rs.Primary.ID
		apiClient := testAccProvider.Meta().(pritunl.Client)
		_, err := apiClient.GetOrganization(name)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckItemBasic() string {
	return fmt.Sprintf(`
        resource "pritunl_org" "test_org" {
          name = "test_org"
        }
    `)
}

func testAccCheckItemDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(pritunl.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pritunl_org" {
			continue
		}

		_, err := apiClient.GetOrganization(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Alert still exists")
		}
		notFoundErr := "not found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
