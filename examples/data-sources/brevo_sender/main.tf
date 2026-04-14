data "brevo_sender" "example" {
  email = "noreply@mail.example.com"
}

output "sender_id" {
  value = data.brevo_sender.example.id
}
