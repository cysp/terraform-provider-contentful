resource "contentful_webhook" "test" {
  space_id = var.space_id

  name = var.webhook_id

  url = "https://example.org/webhook"

  topics = ["Entry.publish", "Entry.unpublish"]
}
