variable "signing_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "contentful_webhook_signing_secret" "this" {
  space_id = var.contentful_space_id

  value_wo = var.signing_secret
}
