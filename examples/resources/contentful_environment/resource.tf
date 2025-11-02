resource "contentful_environment" "example" {
  space_id              = "your-space-id"
  environment_id        = "staging-yyyy-mm-dd"
  name                  = "Staging (YYYY-MM-DD)"
  source_environment_id = "master"
}
