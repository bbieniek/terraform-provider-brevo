data "brevo_webhook" "example" {
  id = 7
}

output "webhook_url" {
  value = data.brevo_webhook.example.url
}
