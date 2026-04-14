resource "brevo_webhook" "delivery" {
  url         = "https://api.example.com/webhooks/brevo"
  description = "Track delivery events"
  events      = ["delivered", "hardBounce", "softBounce", "blocked", "spam"]
  type        = "transactional"
}
