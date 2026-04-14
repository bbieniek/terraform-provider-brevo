# Terraform Provider Brevo

The Brevo provider allows [Terraform](https://www.terraform.io) to manage [Brevo](https://www.brevo.com/) (formerly Sendinblue) resources such as sender domains, senders, email templates, and webhooks.

Visit the [Terraform Registry page](https://registry.terraform.io/providers/bbieniek/brevo/) for published versions and documentation.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.26 (to build the provider plugin)

## Authentication

The provider requires a Brevo API key. You can provide it via the provider configuration or the `BREVO_API_KEY` environment variable:

```hcl
provider "brevo" {
  api_key = var.brevo_api_key
}
```

```shell
export BREVO_API_KEY="your-api-key"
```

## Usage

```hcl
terraform {
  required_providers {
    brevo = {
      source = "bbieniek/brevo"
    }
  }
}

provider "brevo" {}

resource "brevo_domain" "example" {
  name = "mail.example.com"
}

resource "brevo_sender" "example" {
  name  = "My Application"
  email = "noreply@mail.example.com"
}

resource "brevo_email_template" "welcome" {
  name         = "welcome-email"
  subject      = "Welcome to {{ params.appName }}"
  html_content = "<html><body><h1>Welcome!</h1></body></html>"
  sender_name  = "My App"
  sender_email = "noreply@mail.example.com"
  is_active    = true
}

resource "brevo_webhook" "delivery_events" {
  url         = "https://api.example.com/webhooks/brevo"
  description = "Track delivery events"
  events      = ["delivered", "hardBounce", "softBounce", "blocked", "spam"]
  type        = "transactional"
}
```

## Resources

| Resource | Description |
| --- | --- |
| `brevo_domain` | Manages sender domains and exposes DNS records (DKIM, verification code) |
| `brevo_sender` | Manages sender email addresses with display names |
| `brevo_email_template` | Manages transactional email templates |
| `brevo_webhook` | Manages webhook subscriptions for event notifications |

## Data Sources

| Data Source | Lookup By | Description |
| --- | --- | --- |
| `brevo_domain` | Domain name | Reads an existing sender domain |
| `brevo_sender` | Email address | Reads an existing sender |
| `brevo_email_template` | Template ID | Reads an existing email template |
| `brevo_webhook` | Webhook ID | Reads an existing webhook |

All resources support `terraform import`. See the [documentation](https://registry.terraform.io/providers/bbieniek/brevo/latest/docs) for details.

## Building the Provider

Clone the repository and build:

```shell
git clone https://github.com/bbieniek/terraform-provider-brevo.git
cd terraform-provider-brevo
make build
```

To install the provider locally:

```shell
make install
```

## Developing the Provider

### Running Tests

Unit tests:

```shell
make test
```

Acceptance tests (requires a valid `BREVO_API_KEY`):

```shell
make testacc
```

### Linting

```shell
make lint
```

## License

This project is licensed under the [Mozilla Public License 2.0](LICENSE).
