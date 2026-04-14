terraform {
  required_providers {
    brevo = {
      source = "bbieniek/brevo"
    }
  }
}

provider "brevo" {
  api_key = var.brevo_api_key
}

variable "brevo_api_key" {
  type      = string
  sensitive = true
}
