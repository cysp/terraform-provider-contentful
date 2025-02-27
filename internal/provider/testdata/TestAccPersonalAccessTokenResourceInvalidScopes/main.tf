resource "contentful_personal_access_token" "test" {
  name       = "terraform-provider-contentful-acctest-${var.personal_access_token_id}"
  scopes     = ["content_management_invalid"]
  expires_in = 5 * 60
}
