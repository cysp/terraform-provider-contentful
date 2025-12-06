import {
  identity = {
    space_id       = var.contentful_space_id
    environment_id = var.contentful_environment_id
    entry_id       = var.entry_id
  }
  to = contentful_entry.this
}
