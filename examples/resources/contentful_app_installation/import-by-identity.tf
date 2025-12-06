import {
  identity = {
    space_id          = var.contentful_space_id
    environment_id    = var.contentful_environment_id
    app_definition_id = var.app_definition_id
  }
  to = contentful_app_installation.this
}
