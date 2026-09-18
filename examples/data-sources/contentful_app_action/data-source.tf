data "contentful_app_action" "this" {
  organization_id   = var.contentful_organization_id
  app_definition_id = var.app_definition_id
  app_action_id     = var.app_action_id
}
