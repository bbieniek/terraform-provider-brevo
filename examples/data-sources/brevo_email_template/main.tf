data "brevo_email_template" "example" {
  id = 42
}

output "template_name" {
  value = data.brevo_email_template.example.name
}
