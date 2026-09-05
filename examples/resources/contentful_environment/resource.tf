resource "contentful_environment" "staging" {
  space_id              = "your-space-id"
  environment_id        = "staging-yyyy-mm-dd"
  name                  = "Staging (YYYY-MM-DD)"
  source_environment_id = "master"
}

data "contentful_environment_status_ready" "staging" {
  space_id       = contentful_environment.staging.space_id
  environment_id = contentful_environment.staging.environment_id
}

resource "contentful_environment_alias" "staging" {
  space_id              = contentful_environment.staging.space_id
  environment_alias_id  = "staging"
  target_environment_id = contentful_environment.staging.environment_id

  depends_on = [data.contentful_environment_status_ready.staging]
}
