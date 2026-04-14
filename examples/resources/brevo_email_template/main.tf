resource "brevo_email_template" "welcome" {
  name         = "welcome-email"
  subject      = "Welcome to {{ params.appName }}"
  html_content = file("${path.module}/templates/welcome.html")
  sender_name  = "My App"
  sender_email = "noreply@mail.example.com"
  is_active    = true
}
