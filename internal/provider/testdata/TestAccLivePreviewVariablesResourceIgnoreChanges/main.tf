variable "space_id" {
  type    = string
  default = "space"
}
variable "environment_id" {
  type    = string
  default = "environment"
}
variable "variables" {
  type = string
}
variable "read_timeout" {
  type    = string
  default = "3m"
}
resource "contentful_live_preview_variables" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id
  variables      = var.variables
  lifecycle {
    ignore_changes = [variables]
  }
  timeouts = {
    create = "3m"
    read   = var.read_timeout
    update = "3m"
    delete = "3m"
  }
}
