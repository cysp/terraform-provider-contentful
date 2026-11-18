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
  description = "Randomly generated Basic password expected by the webhook receiver."
  type        = string
  ephemeral   = true
  sensitive   = true
}
