variable "contentful_organization_id" {
  type = string
}

variable "app_definition_id" {
  type = string
}

variable "app_signing_secret" {
  type      = string
  sensitive = true
}
