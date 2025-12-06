import {
  identity = {
    space_id   = var.contentful_space_id
    webhook_id = var.webhook_id
  }
  to = contentful_webhook.this
}
