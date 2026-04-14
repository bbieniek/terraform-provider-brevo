resource "brevo_domain" "example" {
  name = "mail.example.com"
}

output "dkim_record_1" {
  value = brevo_domain.example.dkim_record_1
}

output "brevo_code" {
  value = brevo_domain.example.brevo_code
}
