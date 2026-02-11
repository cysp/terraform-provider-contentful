locals {
  staging_environments = [
    "2026-02-11",
  ]

  active_staging_environment = local.staging_environments[0]
}

resource "contentful_environment" "staging" {
  for_each = toset(local.staging_environments)

  space_id       = var.contentful_space_id
  environment_id = "staging-${each.key}"
  name           = "Staging (${each.key})"
}

data "contentful_environment_status_ready" "staging_active" {
  space_id       = contentful_environment.staging[local.active_staging_environment].space_id
  environment_id = contentful_environment.staging[local.active_staging_environment].environment_id
}

resource "contentful_environment_alias" "staging" {
  space_id              = var.contentful_space_id
  environment_alias_id  = "staging"
  target_environment_id = contentful_environment.staging[local.active_staging_environment].environment_id

  depends_on = [data.contentful_environment_status_ready.staging_active]
}
