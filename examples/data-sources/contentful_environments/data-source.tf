data "contentful_environments" "existing" {
  space_id = var.contentful_space_id
}

locals {
  ready_environment_ids = [
    for environment in data.contentful_environments.existing.environments : environment.environment_id
    if environment.status == "ready"
  ]
}
