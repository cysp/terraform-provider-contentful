variable "contentful_space_id" {
  type = string
}

variable "webhook_signing_secret" {
  type      = string
  sensitive = true
}
