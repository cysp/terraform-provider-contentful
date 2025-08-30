resource "contentful_app_signing_secret" "this" {
  organization_id   = var.organization_id
  app_definition_id = local.app_definition_id

  value = random_password.contentful_app_signing_secret.result
}

resource "random_password" "contentful_app_signing_secret" {
  override_special = "+/=_-"

  length = 64
}
