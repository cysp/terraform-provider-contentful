import {
  identity = {
    space_id        = var.space_id
    environment_id  = var.environment_id
    content_type_id = var.content_type_id
  }
  to = contentful_content_type.this
}
