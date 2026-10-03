terraform {
  required_version = ">= 1.11.1"
}

variable "signing_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "contentful_app_signing_secret" "this" {
  organization_id   = var.contentful_organization_id
  app_definition_id = var.app_definition_id

  value_wo = var.signing_secret
}
