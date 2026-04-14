package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/bbieniek/terraform-provider-brevo/internal/acctest/testprovider"
)

func TestProviderConfigure_fromEnv(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testprovider.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `provider "brevo" {}`,
			},
		},
	})
}
