import {
  identity = {
    space_id        = var.contentful_space_id
    environment_id  = var.contentful_environment_id
    content_type_id = var.content_type_id
  }
  to = contentful_editor_interface.this
}
