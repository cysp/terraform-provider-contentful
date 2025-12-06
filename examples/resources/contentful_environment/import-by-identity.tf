import {
  identity = {
    space_id       = var.space_id
    environment_id = var.environment_id
  }
  to = contentful_environment.this
}
