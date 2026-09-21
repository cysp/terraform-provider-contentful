variable "target_url" {
  type    = string
  default = "https://example.invalid/events"
}

resource "contentful_app_event_subscription" "test" {
  organization_id   = "organization"
  app_definition_id = "app"
  topics            = ["Entry.publish", "Asset.publish"]
  target_url        = var.target_url
}
