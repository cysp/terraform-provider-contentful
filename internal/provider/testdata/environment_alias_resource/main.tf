resource "contentful_environment_alias" "test" {
  space_id             = var.space_id
  environment_alias_id = var.test_environment_alias_id

  target_environment_id = var.target_environment_id
}
