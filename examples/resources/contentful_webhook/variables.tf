variable "contentful_space_id" {
  type = string
}

variable "webhook_id" {
  type = string
}

variable "webhook_secret" {
  description = "Shared secret expected by the webhook receiver."
  type        = string
  ephemeral   = true
  sensitive   = true
}

variable "webhook_password" {
  description = "Basic authentication password supplied for webhook writes with Terraform 1.11.1 or later."
  type        = string
  ephemeral   = true
  sensitive   = true
}
