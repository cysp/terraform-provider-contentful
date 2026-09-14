data "contentful_environment_alias" "existing" {
  space_id             = var.contentful_space_id
  environment_alias_id = "master"
}

data "contentful_environment" "target" {
  space_id       = data.contentful_environment_alias.existing.space_id
  environment_id = data.contentful_environment_alias.existing.target_environment_id
}
