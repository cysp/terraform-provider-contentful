resource "contentful_environment_alias" "example" {
  space_id              = "space-id"
  environment_alias_id  = "staging"
  target_environment_id = "staging-yyyy-mm-dd"
}
