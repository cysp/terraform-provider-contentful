resource "contentful_app_key" "this" {
  organization_id   = var.contentful_organization_id
  app_definition_id = var.app_definition_id

  jwk = {
    alg = "RS256"
    kty = "RSA"
    use = "sig"
    kid = var.app_key_kid
    x5c = [var.app_key_x5c]
    x5t = var.app_key_x5t
  }
}
