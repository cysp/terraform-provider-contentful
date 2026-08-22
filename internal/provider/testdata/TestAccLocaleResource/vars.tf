variable "space_id" {
  type = string
}

variable "environment_id" {
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
  default = false
}
