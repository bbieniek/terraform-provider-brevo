package domain_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/bbieniek/terraform-provider-brevo/internal/acctest/testprovider"
)

func TestAccDomainDataSource(t *testing.T) {
	domainName := fmt.Sprintf("tf-test-%s.example.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testprovider.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDataSourceConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.brevo_domain.test", "name", domainName),
					resource.TestCheckResourceAttrSet("data.brevo_domain.test", "id"),
					resource.TestCheckResourceAttrSet("data.brevo_domain.test", "dkim_record_1"),
					resource.TestCheckResourceAttrSet("data.brevo_domain.test", "dkim_record_2"),
					resource.TestCheckResourceAttrSet("data.brevo_domain.test", "brevo_code"),
					resource.TestCheckResourceAttrSet("data.brevo_domain.test", "verified"),
				),
			},
		},
	})
}

func testAccDomainDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "brevo_domain" "test" {
  name = %q
}

data "brevo_domain" "test" {
  name = brevo_domain.test.name
}
`, name)
}
