resource "contentful_app_event_subscription" "test" {
  organization_id   = "organization"
  app_definition_id = "app"
  topics            = ["Entry.publish", "Asset.publish"]
  target_url        = "https://example.invalid/events"
}
