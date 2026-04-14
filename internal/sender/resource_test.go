package sender_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/bbieniek/terraform-provider-brevo/internal/acctest/testprovider"
)

func TestAccSenderResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	senderName := fmt.Sprintf("tf-test-%s", rName)
	senderEmail := fmt.Sprintf("tf-test-%s@gmail.com", rName)
	updatedName := fmt.Sprintf("tf-test-%s-updated", rName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testprovider.ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSenderResourceConfig(senderName, senderEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("brevo_sender.test", "name", senderName),
					resource.TestCheckResourceAttr("brevo_sender.test", "email", senderEmail),
					resource.TestCheckResourceAttrSet("brevo_sender.test", "id"),
				),
			},
			{
				Config: testAccSenderResourceConfig(updatedName, senderEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("brevo_sender.test", "name", updatedName),
					resource.TestCheckResourceAttr("brevo_sender.test", "email", senderEmail),
				),
			},
			{
				ResourceName:      "brevo_sender.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSenderResourceConfig(name, email string) string {
	return fmt.Sprintf(`
resource "brevo_sender" "test" {
  name  = %q
  email = %q
}
`, name, email)
}
