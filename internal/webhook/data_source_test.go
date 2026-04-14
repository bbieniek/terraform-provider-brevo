package webhook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/bbieniek/terraform-provider-brevo/internal/acctest/testprovider"
)

func TestAccWebhookDataSource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	webhookUrl := fmt.Sprintf("https://%s.example.com/webhook", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testprovider.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookDataSourceConfig(webhookUrl),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.brevo_webhook.test", "url", webhookUrl),
					resource.TestCheckResourceAttr("data.brevo_webhook.test", "type", "transactional"),
					resource.TestCheckResourceAttrSet("data.brevo_webhook.test", "id"),
					resource.TestCheckResourceAttrSet("data.brevo_webhook.test", "batched"),
				),
			},
		},
	})
}

func testAccWebhookDataSourceConfig(url string) string {
	return fmt.Sprintf(`
resource "brevo_webhook" "test" {
  url    = %q
  events = ["delivered", "hardBounce"]
  type   = "transactional"
}

data "brevo_webhook" "test" {
  id = brevo_webhook.test.id
}
`, url)
}
