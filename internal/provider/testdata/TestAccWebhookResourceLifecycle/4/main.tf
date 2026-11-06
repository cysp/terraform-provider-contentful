resource "contentful_webhook" "test" {
  space_id = var.space_id

  name = var.webhook_id

  url = "https://example.org/webhook"

  topics = ["Entry.save", "Entry.publish", "Entry.unpublish"]
  filters = [
    {
      equals = {
        doc   = "sys.environment.sys.id"
        value = "master"
      }
    },
  ]

  transformation = {
    method                 = "POST"
    content_type           = "application/vnd.contentful.management.v1+json"
    include_content_length = true
  }
}
