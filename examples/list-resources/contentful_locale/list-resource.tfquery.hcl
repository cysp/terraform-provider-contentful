list "contentful_locale" "locales" {
  provider = contentful

  config {
    space_id       = var.contentful_space_id
    environment_id = var.contentful_environment_id
  }

  include_resource = true
}
