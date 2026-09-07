variable "contentful_space_id" {
  type = string
}

variable "webhook_id" {
  type = string
}

variable "webhook_secret" {
  description = "Shared secret expected by the webhook receiver."
  type        = string
  sensitive   = true
}
