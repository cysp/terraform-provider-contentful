import {
  identity = {
    space_id   = var.space_id
    webhook_id = var.webhook_id
  }
  to = contentful_webhook.this
}
