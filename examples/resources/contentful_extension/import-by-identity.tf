import {
  identity = {
    space_id       = var.contentful_space_id
    environment_id = var.contentful_environment_id
    extension_id   = "custom-field-extension"
  }
  to = contentful_extension.example
}
