resource "contentful_app_installation" "cool_app" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  app_definition_id = var.cool_app_definition_id
}

resource "contentful_app_installation" "cool_marketplace_app" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  app_definition_id = var.cool_marketplace_app_definition_id

  marketplace = [
    "i-accept-end-user-license-agreement",
    "i-accept-marketplace-terms-of-service",
    "i-accept-privacy-policy",
  ]
}
