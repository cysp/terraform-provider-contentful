resource "contentful_locale" "example" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  name          = "German"
  code          = "de-DE"
  fallback_code = "en-US"
  optional      = true
}
