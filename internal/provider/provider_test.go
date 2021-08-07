package provider

import (
	"github.com/disc/terraform-provider-pritunl/internal/pritunl"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
