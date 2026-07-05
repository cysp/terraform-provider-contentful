resource "contentful_app_key" "test" {
  organization_id   = var.organization_id
  app_definition_id = var.app_definition_id

  jwk = {
    kid = var.key_kid
    x5c = ["certificate"]
    x5t = var.key_kid
  }
}
