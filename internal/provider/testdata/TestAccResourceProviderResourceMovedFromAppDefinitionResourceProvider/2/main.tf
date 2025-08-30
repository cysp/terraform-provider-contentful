moved {
  from = contentful_app_definition_resource_provider.test
  to   = contentful_resource_provider.test
}

resource "contentful_resource_provider" "test" {
  organization_id   = var.organization_id
  app_definition_id = var.app_definition_id

  resource_provider_id = "test"
  function_id          = "resourceProvider"
}
