resource "contentful_webhook" "test" {
  space_id = var.space_id

  name = "test"

  url = "https://example.org/webhook"
}
