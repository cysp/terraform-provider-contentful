resource "contentful_resource_provider" "this" {
  organization_id      = var.organization_id
  app_definition_id    = local.app_definition_id
  resource_provider_id = "ResourceProviderName"
  function_id          = "resourceProvider"
}
