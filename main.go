package main

import (
	"github.com/bpicolo/terraform-provider-algolia/algolia"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return algolia.Provider()
		},
	})
}
