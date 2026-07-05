resource "contentful_app_key" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"

  jwk = null
}
