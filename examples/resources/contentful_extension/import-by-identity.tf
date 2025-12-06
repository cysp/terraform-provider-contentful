import {
  identity = {
    space_id       = var.space_id
    environment_id = var.environment_id
    extension_id   = var.extension_id
  }
  to = contentful_extension.this
}
