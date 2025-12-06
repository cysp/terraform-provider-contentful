resource "contentful_resource_provider" "this" {
  organization_id      = var.contentful_organization_id
  app_definition_id    = var.app_definition_id
  resource_provider_id = "ResourceProviderName"
  function_id          = "resourceProvider"
}
