package provider

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"regexp"
	"terraform-pritunl/internal/pritunl"
	"testing"
)

func TestAccOrganization_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckOrganizationBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists("pritunl_organization.test_org"),
					resource.TestCheckResourceAttr("pritunl_organization.test_org", "name", "test_org"),
				),
			},
		},
	})
}

//func TestAccOrganization_Update(t *testing.T) {
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		Providers:    testAccProviders,
//		CheckDestroy: testAccCheckItemDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccCheckOrganizationUpdatePre(),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckOrganizationExists("example_item.test_update"),
//					resource.TestCheckResourceAttr(
//						"example_item.test_update", "name", "test_update"),
//					resource.TestCheckResourceAttr(
//						"example_item.test_update", "description", "hello"),
//					resource.TestCheckResourceAttr(
//						"example_item.test_update", "tags.#", "2"),
//					resource.TestCheckResourceAttr("example_item.test_update", "tags.1931743815", "tag1"),
//					resource.TestCheckResourceAttr("example_item.test_update", "tags.1477001604", "tag2"),
//				),
//			},
//			{
//				Config: testAccCheckOrganizationUpdatePost(),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckOrganizationExists("example_item.test_update"),
//					resource.TestCheckResourceAttr(
//						"example_item.test_update", "name", "test_update"),
//					resource.TestCheckResourceAttr(
//						"example_item.test_update", "description", "updated description"),
//					resource.TestCheckResourceAttr(
//						"example_item.test_update", "tags.#", "2"),
//					resource.TestCheckResourceAttr("example_item.test_update", "tags.1931743815", "tag1"),
//					resource.TestCheckResourceAttr("example_item.test_update", "tags.1477001604", "tag2"),
//				),
//			},
//		},
//	})
//}

func testAccCheckOrganizationBasic() string {
	return fmt.Sprintf(`
        resource "pritunl_organization" "test_org" {
          name = "test_org"
        }
    `)
}

func testAccCheckOrganizationExists(resource string) resource.TestCheckFunc {
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
			return fmt.Errorf("error fetching organization with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("PRITUNL_URL"); v == "" {
		t.Fatal("PRITUNL_URL must be set for acceptance tests")
	}
	if v := os.Getenv("PRITUNL_TOKEN"); v == "" {
		t.Fatal("PRITUNL_TOKEN must be set for acceptance tests")
	}
	if v := os.Getenv("PRITUNL_SECRET"); v == "" {
		t.Fatal("PRITUNL_SECRET must be set for acceptance tests")
	}
}

func testAccCheckItemDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(pritunl.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "pritunl_organization" {
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

//func TestAccAlert_importBasic(t *testing.T) {
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		Providers:    testAccProviders,
//		CheckDestroy: testAccCheckItemDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccCheckExampleItemImporter_basic(),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckExampleItemExists("pritunl_organization.organization_import"),
//				),
//			},
//			{
//				ResourceName:      "pritunl_organization.organization_import",
//				ImportState:       true,
//				ImportStateVerify: true,
//			},
//		},
//	})
//}

func testAccCheckOrganizationImporter_basic() string {
	return fmt.Sprintf(`
		resource "pritunl_organization" "organization_import" {
			name = "organization_import"
		}
	`)
}
