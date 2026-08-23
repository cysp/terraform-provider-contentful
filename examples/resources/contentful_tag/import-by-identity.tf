import {
  identity = {
    space_id       = var.contentful_space_id
    environment_id = var.contentful_environment_id
    tag_id         = var.tag_id
  }
  to = contentful_tag.example
}
