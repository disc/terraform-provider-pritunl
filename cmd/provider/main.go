package main

import (
	"github.com/hashicorp/terraform/plugin"
	"terraform-pritunl/internal/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.Provider,
	})
}
