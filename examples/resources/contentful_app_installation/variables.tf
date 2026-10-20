variable "contentful_space_id" {
  type = string
}

variable "contentful_environment_id" {
  type = string
}

variable "app_definition_id" {
  type = string
}

variable "app_api_key" {
  description = "Secret installation parameter for the custom app."
  type        = string
  sensitive   = true
}

variable "marketplace_app_definition_id" {
  type = string
}
