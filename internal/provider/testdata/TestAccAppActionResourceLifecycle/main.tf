variable "organization_id" {
  type = string
}

variable "custom_input" {
  type = object({
    parameters        = optional(string)
    parameters_schema = optional(string)
    result_schema     = optional(string)
    description       = optional(string)
  })
  default = {
    parameters_schema = "{\"properties\":{\"message\":{\"type\":\"string\"}},\"type\":\"object\"}"
    result_schema     = "{\"type\":\"object\"}"
    description       = "Temporary acceptance action"
  }
}

variable "builtin_category" {
  type    = string
  default = "Entries.v1.0"
}

variable "function_target" {
  type = object({
    type        = string
    url         = optional(string)
    function_id = optional(string)
  })
  default = {
    type        = "function-invocation"
    function_id = "acceptancefunction"
  }
}

resource "contentful_app_definition" "test" {
  organization_id = var.organization_id
  name            = "Terraform App Action acceptance"
  locations       = []
}

resource "contentful_app_action" "test" {
  organization_id   = contentful_app_definition.test.organization_id
  app_definition_id = contentful_app_definition.test.app_definition_id
  name              = "Action"
  category          = "Custom"
  type              = "endpoint"
  url               = "https://example.invalid/terraform-action"
  parameters        = var.custom_input.parameters
  parameters_schema = var.custom_input.parameters_schema
  result_schema     = var.custom_input.result_schema
  description       = var.custom_input.description
}

resource "contentful_app_action" "builtin" {
  organization_id   = contentful_app_definition.test.organization_id
  app_definition_id = contentful_app_definition.test.app_definition_id
  name              = "Builtin"
  category          = var.builtin_category
  type              = "endpoint"
  url               = "https://example.invalid/terraform-action"
}

# The undeployed Function link exercises definition management, not execution.
resource "contentful_app_action" "function" {
  organization_id   = contentful_app_definition.test.organization_id
  app_definition_id = contentful_app_definition.test.app_definition_id
  name              = "Function action"
  category          = "Custom"
  type              = var.function_target.type
  function_id       = var.function_target.function_id
  url               = var.function_target.url
  parameters = jsonencode([
    { id = "text", name = "Text", type = "Symbol" },
    { id = "number", name = "Number", type = "Number", required = false, default = 0 },
    { id = "flag", name = "Flag", type = "Boolean", default = false },
    { id = "choice", name = "Choice", type = "Enum", options = ["a", "b"] }
  ])
  description = ""
}

data "contentful_app_action" "one" {
  organization_id   = contentful_app_action.test.organization_id
  app_definition_id = contentful_app_action.test.app_definition_id
  app_action_id     = contentful_app_action.test.app_action_id
}

data "contentful_app_actions" "all" {
  organization_id   = contentful_app_definition.test.organization_id
  app_definition_id = contentful_app_definition.test.app_definition_id
  depends_on        = [contentful_app_action.test, contentful_app_action.builtin, contentful_app_action.function]
}
