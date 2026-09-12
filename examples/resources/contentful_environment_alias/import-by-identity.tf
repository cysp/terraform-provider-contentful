import {
  identity = {
    space_id             = var.contentful_space_id
    environment_alias_id = "staging"
  }
  to = contentful_environment_alias.example
}
