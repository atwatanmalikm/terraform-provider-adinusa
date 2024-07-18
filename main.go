package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"terraform-provider-adinusa/adinusa"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: adinusa.Provider,
	})
}