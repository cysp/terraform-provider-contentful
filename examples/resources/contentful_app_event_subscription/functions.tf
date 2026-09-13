resource "contentful_app_event_subscription" "this" {
  organization_id   = var.contentful_organization_id
  app_definition_id = var.app_definition_id

  topics                     = ["Entry.publish"]
  filter_function_id         = "eventFilter"
  transformation_function_id = "eventTransformation"
  handler_function_id        = "eventHandler"
}
