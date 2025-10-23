variable "space_id" {
  type = string
}

variable "environment_id" {
  type = string
}

variable "entry_id" {
  type = string
}

variable "content_type_id" {
  type = string
}

variable "fields" {
  type    = map(string)
  default = {}
}

resource "contentful_entry" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id
  entry_id       = var.entry_id

  content_type_id = var.content_type_id

  fields = var.fields
}
