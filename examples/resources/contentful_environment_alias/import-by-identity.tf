import {
  identity = {
    space_id             = var.space_id
    environment_alias_id = var.environment_alias_id
  }
  to = contentful_environment_alias.this
}
