import {
  identity = {
    space_id       = var.contentful_space_id
    environment_id = var.contentful_environment_id
    locale_id      = var.contentful_locale_id
  }
  to = contentful_locale.example
}
