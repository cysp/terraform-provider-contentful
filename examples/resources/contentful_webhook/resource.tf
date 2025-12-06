resource "contentful_webhook" "this" {
  space_id = var.contentful_space_id

  name = "Example Webhook"
  url  = "https://example.org/webhook"

  active = false

  topics = ["Entry.save", "Entry.publish", "Entry.unpublish"]

  filters = [
    {
      equals = {
        doc   = "sys.environment.sys.id"
        value = "master"
      }
    },
  ]

  headers = {
    "X-Webhook-Secret" = {
      value  = "abcdef"
      secret = true
    },
  }

  transformation = {
    method                 = "POST"
    content_type           = "application/vnd.contentful.management.v1+json"
    include_content_length = true
  }
}
