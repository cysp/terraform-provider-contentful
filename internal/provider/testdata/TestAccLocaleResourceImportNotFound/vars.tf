variable "space_id" {
  type = string
}

variable "locale_id" {
  type = string
}

variable "name" {
  type = string
}

variable "code" {
  type = string
}

variable "fallback_code" {
  type    = string
  default = null
}

variable "optional" {
  type    = bool
  default = null
}

variable "content_delivery_api" {
  type    = bool
  default = null
}

variable "content_management_api" {
  type    = bool
  default = null
}
