import {
  identity = {
    space_id       = var.space_id
    environment_id = var.environment_id
    entry_id       = var.entry_id
  }
  to = contentful_entry.this
}
