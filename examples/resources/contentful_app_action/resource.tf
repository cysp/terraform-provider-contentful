resource "contentful_app_action" "this" {
  organization_id   = var.contentful_organization_id
  app_definition_id = var.app_definition_id
  name              = "Build site"
  category          = "Custom"
  type              = "endpoint"
  url               = "https://example.com/build"
  parameters_schema = jsonencode({
    type = "object"
    properties = {
      branch = { type = "string" }
    }
    required             = ["branch"]
    additionalProperties = false
  })
  result_schema = jsonencode({
    type = "object"
    properties = {
      build_id = { type = "string" }
    }
  })
}

# To invoke this action, deploy a Function that accepts appaction.call.
resource "contentful_app_action" "function" {
  organization_id   = var.contentful_organization_id
  app_definition_id = var.app_definition_id
  name              = "Process entries"
  category          = "Entries.v1.0"
  type              = "function-invocation"
  function_id       = var.function_id
  # Contentful supplies this category's parameters; do not configure parameters.
}
