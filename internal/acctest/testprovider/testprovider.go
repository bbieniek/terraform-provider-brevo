package testprovider

import (
	"github.com/bbieniek/terraform-provider-brevo/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func ProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"brevo": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}
