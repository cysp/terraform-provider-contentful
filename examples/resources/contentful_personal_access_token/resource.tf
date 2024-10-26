resource "contentful_personal_access_token" "this" {
  name       = "management access token"
  scopes     = ["content_management_manage"]
  expires_in = 90 * 24 * 60 * 60
}
