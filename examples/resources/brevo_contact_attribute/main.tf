# Custom contact attribute "FIRMA" (company), used in templates as {{ contact.FIRMA }}.
resource "brevo_contact_attribute" "firma" {
  name     = "FIRMA"
  category = "normal"
  type     = "text"
}

# A list newcomers join when they sign up.
resource "brevo_list" "newsletter_signups" {
  name      = "Newsletter Signups"
  folder_id = 1
}

# A transactional template that references the standard FIRSTNAME/EMAIL fields
# plus the custom FIRMA attribute defined above.
resource "brevo_email_template" "welcome" {
  name         = "Welcome — list join"
  subject      = "Welcome to our list, {{ contact.FIRSTNAME }}"
  sender_name  = "Rys"
  sender_email = "hello@example.com"
  html_content = <<-EOT
    <p>Hello {{ contact.FIRSTNAME }},</p>
    <p>Thanks for joining. We have you on file as:</p>
    <ul>
      <li>Name: {{ contact.FIRSTNAME }}</li>
      <li>Firma: {{ contact.FIRMA }}</li>
      <li>Email: {{ contact.EMAIL }}</li>
    </ul>
  EOT
  is_active    = true
}

# Note: Brevo's marketing automation workflows ("when contact joins list X,
# send template Y") are not exposed for CRUD via the public REST API. Wire
# the workflow once in the Brevo UI: Automations → New workflow → trigger
# "Contact added to a list" (pick the list above) → action "Send an email"
# (pick the template above). Terraform manages the inputs; the trigger lives
# in the UI.
