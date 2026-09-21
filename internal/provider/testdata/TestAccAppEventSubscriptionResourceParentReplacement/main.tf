variable "organization_id" {
  type    = string
  default = "organization"
}

variable "app_definition_id" {
  type    = string
  default = "app"
}

resource "contentful_app_event_subscription" "test" {
  organization_id   = var.organization_id
  app_definition_id = var.app_definition_id
  topics            = ["Entry.publish", "Asset.publish"]
  target_url        = "https://example.invalid/events"
}
