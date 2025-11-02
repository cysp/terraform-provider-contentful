data "contentful_app_definition" "this" {
  organization_id = var.organization_id

  app_definition_id = "app-definition-id"
}
