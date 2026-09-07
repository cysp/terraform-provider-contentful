resource "contentful_app_installation" "custom" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  app_definition_id = var.app_definition_id
}

resource "contentful_app_installation" "marketplace" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  app_definition_id = var.marketplace_app_definition_id

  marketplace = [
    "i-accept-end-user-license-agreement",
    "i-accept-marketplace-terms-of-service",
    "i-accept-privacy-policy",
  ]
}
