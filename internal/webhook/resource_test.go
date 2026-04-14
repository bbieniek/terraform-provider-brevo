package webhook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/bbieniek/terraform-provider-brevo/internal/acctest/testprovider"
)

func TestAccWebhookResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	webhookUrl := fmt.Sprintf("https://%s.example.com/webhook", rName)
	updatedUrl := fmt.Sprintf("https://%s.example.com/webhook-updated", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testprovider.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookResourceConfig(webhookUrl, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("brevo_webhook.test", "url", webhookUrl),
					resource.TestCheckResourceAttr("brevo_webhook.test", "type", "transactional"),
					resource.TestCheckResourceAttrSet("brevo_webhook.test", "id"),
				),
			},
			{
				Config: testAccWebhookResourceConfig(updatedUrl, "Updated webhook"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("brevo_webhook.test", "url", updatedUrl),
					resource.TestCheckResourceAttr("brevo_webhook.test", "description", "Updated webhook"),
				),
			},
			{
				ResourceName:      "brevo_webhook.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWebhookResourceConfig(url, description string) string {
	descAttr := ""
	if description != "" {
		descAttr = fmt.Sprintf("\n  description = %q", description)
	}
	return fmt.Sprintf(`
resource "brevo_webhook" "test" {
  url    = %q%s
  events = ["delivered", "hardBounce"]
  type   = "transactional"
}
`, url, descAttr)
}
