import {
  identity = {
    space_id        = var.contentful_space_id
    environment_id  = var.contentful_environment_id
    content_type_id = "author"
  }
  to = contentful_content_type.author
}
