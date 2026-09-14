data "contentful_locale" "existing" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id
  locale_id      = var.contentful_locale_id # CMA sys.id, not a locale code.
}

# Use data.contentful_locale.existing.code for localized content fields.
