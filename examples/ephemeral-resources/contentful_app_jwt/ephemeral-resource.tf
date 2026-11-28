variable "app_private_key" {
  type      = string
  sensitive = true
  ephemeral = true
}

ephemeral "contentful_app_jwt" "editorial" {
  app_definition_id = var.app_definition_id
  private_key       = var.app_private_key
}
