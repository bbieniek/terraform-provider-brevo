package template_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/bbieniek/terraform-provider-brevo/internal/acctest/testprovider"
)

func TestAccEmailTemplateDataSource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	templateName := fmt.Sprintf("tf-test-ds-%s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testprovider.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailTemplateDataSourceConfig(templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.brevo_email_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.brevo_email_template.test", "subject", "Test Subject"),
					resource.TestCheckResourceAttrSet("data.brevo_email_template.test", "id"),
					resource.TestCheckResourceAttrSet("data.brevo_email_template.test", "html_content"),
					resource.TestCheckResourceAttrSet("data.brevo_email_template.test", "sender_name"),
					resource.TestCheckResourceAttrSet("data.brevo_email_template.test", "sender_email"),
					resource.TestCheckResourceAttrSet("data.brevo_email_template.test", "is_active"),
				),
			},
		},
	})
}

func testAccEmailTemplateDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "brevo_email_template" "test" {
  name         = %q
  subject      = "Test Subject"
  html_content = "<html><body><h1>Hello</h1></body></html>"
  sender_name  = "Sender"
  sender_email = "sender@example.com"
}

data "brevo_email_template" "test" {
  id = brevo_email_template.test.id
}
`, name)
}
