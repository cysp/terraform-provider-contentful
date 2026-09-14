data "contentful_locales" "existing" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  lifecycle {
    postcondition {
      condition     = length([for locale in self.locales : locale if locale.default]) == 1
      error_message = "Exactly one default Locale is required."
    }
  }
}

locals {
  default_locale = one([for locale in data.contentful_locales.existing.locales : locale if locale.default])
}

# Use local.default_locale.code in localized content, or .locale_id in a Locale lookup.
