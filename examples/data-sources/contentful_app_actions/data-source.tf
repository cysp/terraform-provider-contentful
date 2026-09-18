data "contentful_app_actions" "this" {
  organization_id   = var.contentful_organization_id
  app_definition_id = var.app_definition_id
}
