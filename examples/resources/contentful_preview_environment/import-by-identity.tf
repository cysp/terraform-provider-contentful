import {
  identity = {
    space_id               = var.contentful_space_id
    preview_environment_id = var.preview_environment_id
  }
  to = contentful_preview_environment.this
}
