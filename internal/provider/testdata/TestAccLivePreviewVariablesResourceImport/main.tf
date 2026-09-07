variable "variables" {
  type = string
}
resource "contentful_live_preview_variables" "test" {
  space_id       = "space"
  environment_id = "environment"
  variables      = var.variables
}
