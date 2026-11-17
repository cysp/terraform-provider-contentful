resource "contentful_app_signing_secret" "this" {
  organization_id   = var.contentful_organization_id
  app_definition_id = var.app_definition_id
  value             = var.app_signing_secret
}
