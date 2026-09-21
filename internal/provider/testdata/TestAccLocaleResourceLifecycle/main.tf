variable "locale" {
  type = object({
    code                   = string
    fallback_code          = optional(string)
    content_delivery_api   = optional(bool)
    content_management_api = optional(bool)
    optional               = optional(bool)
  })
}

resource "contentful_locale" "test" {
  space_id               = "space"
  environment_id         = "environment"
  name                   = "German"
  code                   = var.locale.code
  fallback_code          = var.locale.fallback_code
  content_delivery_api   = var.locale.content_delivery_api
  content_management_api = var.locale.content_management_api
  optional               = var.locale.optional
}
