terraform {
  required_providers {
    contentful = {
      source = "cysp/contentful"
    }
  }
}

variable "parameters" {
  type = map(string)
  default = {
    retained = "original"
    visible  = "one"
  }
}

resource "contentful_app_installation" "test" {
  space_id          = "space"
  environment_id    = "master"
  app_definition_id = "app"
  parameters        = { for id, value in var.parameters : id => jsonencode(value) }

  lifecycle {
    ignore_changes = [parameters["retained"]]
  }
}
