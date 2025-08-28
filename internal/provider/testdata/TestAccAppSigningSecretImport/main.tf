resource "contentful_app_signing_secret" "test" {
  organization_id   = var.organization_id
  app_definition_id = var.app_definition_id

  value = "secret"
}
