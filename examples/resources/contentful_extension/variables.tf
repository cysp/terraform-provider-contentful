variable "contentful_space_id" {
  type = string
}

variable "contentful_environment_id" {
  type = string
}

variable "extension_id" {
  type = string
}

variable "extension_api_key" {
  description = "API key supplied to the extension installation."
  type        = string
  sensitive   = true
}
