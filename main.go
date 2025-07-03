package main

import (
	"context"
	"log"

	"example.com/me/komodo-provider/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "example.com/me/komodo-provider",
	}
	err := providerserver.Serve(context.Background(), provider.New(), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
	plugin.Serve(&plugin.ServeOpts{
    	ProviderFunc: provider.Provider,
    	ProviderAddr: "example.com/me/komodo-provider", // 👈 this must match your .tf and .terraformrc
	})
}

