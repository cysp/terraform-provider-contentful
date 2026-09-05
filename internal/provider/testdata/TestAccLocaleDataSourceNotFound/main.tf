data "contentful_locale" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id
  locale_id      = var.locale_id
}
