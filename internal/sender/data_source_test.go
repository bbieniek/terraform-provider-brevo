package sender_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/bbieniek/terraform-provider-brevo/internal/acctest/testprovider"
)

func TestAccSenderDataSource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	senderName := fmt.Sprintf("tf-test-%s", rName)
	senderEmail := fmt.Sprintf("tf-test-%s@gmail.com", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testprovider.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSenderDataSourceConfig(senderName, senderEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.brevo_sender.test", "email", senderEmail),
					resource.TestCheckResourceAttr("data.brevo_sender.test", "name", senderName),
					resource.TestCheckResourceAttrSet("data.brevo_sender.test", "id"),
				),
			},
		},
	})
}

func testAccSenderDataSourceConfig(name, email string) string {
	return fmt.Sprintf(`
resource "brevo_sender" "test" {
  name  = %q
  email = %q
}

data "brevo_sender" "test" {
  email = brevo_sender.test.email
}
`, name, email)
}
