import {
  identity = {
    space_id   = var.contentful_space_id
    api_key_id = var.api_key_id
  }
  to = contentful_delivery_api_key.this
}
