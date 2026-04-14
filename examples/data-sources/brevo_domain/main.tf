data "brevo_domain" "example" {
  name = "mail.example.com"
}

output "verified" {
  value = data.brevo_domain.example.verified
}
