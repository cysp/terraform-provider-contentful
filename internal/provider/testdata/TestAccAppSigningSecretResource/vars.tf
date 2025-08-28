variable "organization_id" {
  type = string
}

variable "app_definition_id" {
  type = string
}

variable "signing_secret_value" {
  type      = string
  sensitive = true
}
