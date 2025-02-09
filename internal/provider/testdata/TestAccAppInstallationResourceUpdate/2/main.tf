resource "contentful_app_installation" "test" {
  space_id          = var.space_id
  environment_id    = var.environment_id
  app_definition_id = var.test_app_definition_id

  parameters = jsonencode({ foo = "bar" })
}
