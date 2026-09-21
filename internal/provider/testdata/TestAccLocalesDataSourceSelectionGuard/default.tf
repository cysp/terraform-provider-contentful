data "contentful_locales" "test" {
  space_id       = "space"
  environment_id = "master"

  lifecycle {
    postcondition {
      condition     = length([for locale in self.locales : locale if locale.default]) == 1
      error_message = "Exactly one selected Locale is required."
    }
  }
}

locals {
  selected = one([for locale in data.contentful_locales.test.locales : locale if locale.default])
}

resource "contentful_entry" "test" {
  space_id        = data.contentful_locales.test.space_id
  environment_id  = data.contentful_locales.test.environment_id
  content_type_id = "article"
  fields          = { title = jsonencode({ (local.selected.code) = "Hello" }) }
}
