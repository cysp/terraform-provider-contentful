resource "contentful_app_event_subscription" "this" {
  organization_id   = var.contentful_organization_id
  app_definition_id = var.app_definition_id

  topics     = ["Entry.publish", "Asset.publish"]
  target_url = "https://example.com/contentful/events"
}
