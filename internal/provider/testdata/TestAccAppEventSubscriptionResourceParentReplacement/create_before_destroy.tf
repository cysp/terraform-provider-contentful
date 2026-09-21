variable "subscription" {
  type = object({
    organization_id            = optional(string, "organization")
    app_definition_id          = optional(string, "app")
    topics                     = optional(list(string), ["Entry.publish", "Asset.publish"])
    target_url                 = optional(string)
    filter_function_id         = optional(string)
    transformation_function_id = optional(string)
    handler_function_id        = optional(string)
    timeouts                   = optional(object({ update = optional(string) }))
  })
  default = {
    target_url = "https://example.invalid/events"
  }
}

resource "contentful_app_event_subscription" "test" {
  organization_id            = var.subscription.organization_id
  app_definition_id          = var.subscription.app_definition_id
  topics                     = var.subscription.topics
  target_url                 = var.subscription.target_url
  filter_function_id         = var.subscription.filter_function_id
  transformation_function_id = var.subscription.transformation_function_id
  handler_function_id        = var.subscription.handler_function_id
  timeouts                   = var.subscription.timeouts

  lifecycle {
    create_before_destroy = true
  }
}
