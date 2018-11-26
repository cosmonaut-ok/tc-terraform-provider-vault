package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/terraform-providers/tc-terraform-provider-vault/vault"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vault.Provider})
}
