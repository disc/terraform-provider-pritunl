package main

import (
	"github.com/hashicorp/terraform/plugin"
	"pritunl-terraform/internal/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.Provider,
	})
}
