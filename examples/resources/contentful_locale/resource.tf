resource "contentful_locale" "example" {
  space_id       = "your-space-id"
  environment_id = "master"

  name          = "German"
  code          = "de-DE"
  fallback_code = "en-US"
  optional      = true
}
