package template_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/bbieniek/terraform-provider-brevo/internal/acctest/testprovider"
)

func TestAccEmailTemplateResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	templateName := fmt.Sprintf("tf-test-%s", rName)
	updatedSubject := fmt.Sprintf("Updated Subject %s", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testprovider.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailTemplateResourceConfig(templateName, "Test Subject", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("brevo_email_template.test", "name", templateName),
					resource.TestCheckResourceAttr("brevo_email_template.test", "subject", "Test Subject"),
					resource.TestCheckResourceAttr("brevo_email_template.test", "is_active", "true"),
					resource.TestCheckResourceAttrSet("brevo_email_template.test", "id"),
					resource.TestCheckResourceAttr("brevo_email_template.test", "sender_name", "Test Sender"),
					resource.TestCheckResourceAttr("brevo_email_template.test", "sender_email", "test@example.com"),
				),
			},
			{
				Config: testAccEmailTemplateResourceConfig(templateName, updatedSubject, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("brevo_email_template.test", "subject", updatedSubject),
					resource.TestCheckResourceAttr("brevo_email_template.test", "is_active", "false"),
				),
			},
			{
				ResourceName:      "brevo_email_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccEmailTemplateResourceConfig(name, subject string, isActive bool) string {
	return fmt.Sprintf(`
resource "brevo_email_template" "test" {
  name         = %q
  subject      = %q
  html_content = "<html><body><h1>Hello</h1></body></html>"
  sender_name  = "Test Sender"
  sender_email = "test@example.com"
  is_active    = %t
}
`, name, subject, isActive)
}
