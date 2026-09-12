# Wait for the existing target environment to be ready before routing requests to it.
resource "contentful_environment_alias" "example" {
  space_id              = var.contentful_space_id
  environment_alias_id  = "staging"
  target_environment_id = "staging-yyyy-mm-dd"
}
