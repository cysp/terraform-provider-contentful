variable "space_id" {
  type = string
}

variable "environment_id" {
  type = string
}

variable "content_type_id" {
  type = string
}

variable "entry_fields" {
  type    = map(string)
  default = {}
}

resource "contentful_entry" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id

  content_type_id = var.content_type_id

  fields = var.entry_fields
}
