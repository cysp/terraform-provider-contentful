import {
  identity = {
    space_id = var.contentful_space_id
    role_id  = var.role_id
  }
  to = contentful_role.this
}
