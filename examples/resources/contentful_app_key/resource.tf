resource "contentful_app_key" "this" {
  organization_id   = var.organization_id
  app_definition_id = var.app_definition_id
}

resource "contentful_app_key" "provided" {
  organization_id   = var.organization_id
  app_definition_id = var.app_definition_id

  # Public key material for a key pair managed outside this provider.
  # private_key is null for provided JWKs.
  jwk = {
    kid = var.key_kid
    x5c = var.x5c
    x5t = var.x5t
  }
}
