resource "contentful_environment" "test" {
  space_id       = var.space_id
  environment_id = var.test_environment_id

  name = var.environment_name
}
