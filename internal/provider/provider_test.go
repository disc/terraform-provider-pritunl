package provider

import (
	"fmt"
	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"strconv"
	"testing"
)

var providerFactories = map[string]func() (*schema.Provider, error){
	"pritunl": func() (*schema.Provider, error) {
		return Provider(), nil
	},
}

var testClient pritunl.Client

func TestMain(m *testing.M) {
	if os.Getenv("TF_ACC") == "" {
		// short circuit non-acceptance test runs
		os.Exit(m.Run())
	}

	url := os.Getenv("PRITUNL_URL")
	token := os.Getenv("PRITUNL_TOKEN")
	secret := os.Getenv("PRITUNL_SECRET")
	insecure, _ := strconv.ParseBool(os.Getenv("PRITUNL_INSECURE"))

	testClient = pritunl.NewClient(url, token, secret, insecure)
	err := testClient.TestApiCall()
	if err != nil {
		panic(err)
	}

	resource.TestMain(m)
}

func preCheck(t *testing.T) {
	variables := []string{
		"PRITUNL_URL",
		"PRITUNL_TOKEN",
		"PRITUNL_SECRET",
	}

	for _, variable := range variables {
		value := os.Getenv(variable)
		if value == "" {
			t.Fatalf("`%s` must be set for acceptance tests!", variable)
		}
	}
}

func importStep(name string, ignore ...string) resource.TestStep {
	step := resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
	}

	if len(ignore) > 0 {
		step.ImportStateVerifyIgnore = ignore
	}

	return step
}

func userImportStep(name string) resource.TestStep {
	step := resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateIdFunc: func(state *terraform.State) (string, error) {
			userId := state.RootModule().Resources["pritunl_user.test"].Primary.Attributes["id"]
			orgId := state.RootModule().Resources["pritunl_organization.test"].Primary.Attributes["id"]

			return fmt.Sprintf("%s-%s", orgId, userId), nil
		},
	}

	return step
}
