resource "contentful_locale" "test" {
  space_id  = var.space_id
  locale_id = var.locale_id
  name      = var.name
  code      = var.code

  fallback_code          = var.fallback_code
  optional               = var.optional
  content_delivery_api   = var.content_delivery_api
  content_management_api = var.content_management_api
}
