import {
  identity = {
    space_id       = var.contentful_space_id
    environment_id = var.contentful_environment_id
  }
  to = contentful_environment.this
}
