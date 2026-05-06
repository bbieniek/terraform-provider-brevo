data "brevo_list" "example" {
  id = 12
}

output "list_name" {
  value = data.brevo_list.example.name
}
