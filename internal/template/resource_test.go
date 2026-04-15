package template_test

import (
	"fmt"
	"os"
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

// Tests use sender id=1 which must already exist and be active in the Brevo account.
// Newly created senders require OTP email verification and cannot be used immediately.
func testAccEmailTemplateResourceConfig(name, subject string, isActive bool) string {
	senderName := os.Getenv("BREVO_SENDER_NAME")
	senderEmail := os.Getenv("BREVO_SENDER_EMAIL")
	return fmt.Sprintf(`
resource "brevo_email_template" "test" {
  name         = %q
  subject      = %q
  html_content = "<html><body><h1>Hello</h1></body></html>"
  sender_name  = %q
  sender_email = %q
  is_active    = %t
}
`, name, subject, senderName, senderEmail, isActive)
}
