data "contentful_environment" "existing" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id
}

# This lookup reports status immediately. For a dependency that must wait:
data "contentful_environment_status_ready" "existing" {
  space_id       = data.contentful_environment.existing.space_id
  environment_id = data.contentful_environment.existing.environment_id
}
