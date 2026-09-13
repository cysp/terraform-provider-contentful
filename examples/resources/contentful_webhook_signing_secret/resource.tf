resource "contentful_webhook_signing_secret" "this" {
  space_id = var.contentful_space_id
  value    = var.webhook_signing_secret
}
