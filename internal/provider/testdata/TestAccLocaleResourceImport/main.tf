resource "contentful_locale" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id

  name          = var.name
  code          = var.code
  fallback_code = var.fallback_code
  optional      = var.optional
}
