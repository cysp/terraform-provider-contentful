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
