package domain_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/bbieniek/terraform-provider-brevo/internal/acctest/testprovider"
)

func TestAccDomainResource(t *testing.T) {
	domainName := fmt.Sprintf("tf-test-%s.example.com", acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testprovider.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("brevo_domain.test", "name", domainName),
					resource.TestCheckResourceAttrSet("brevo_domain.test", "id"),
					resource.TestCheckResourceAttrSet("brevo_domain.test", "dkim_record_1"),
					resource.TestCheckResourceAttrSet("brevo_domain.test", "dkim_record_2"),
					resource.TestCheckResourceAttrSet("brevo_domain.test", "brevo_code"),
					resource.TestCheckResourceAttrSet("brevo_domain.test", "verified"),
				),
			},
			{
				ResourceName:            "brevo_domain.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           domainName,
			},
		},
	})
}

func testAccDomainResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "brevo_domain" "test" {
  name = %q
}
`, name)
}
