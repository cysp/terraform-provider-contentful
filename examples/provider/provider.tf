terraform {
  required_providers {
    contentful = {
      source = "cysp/contentful"
    }
  }
}

provider "contentful" {}

variable "contentful_space_id" {
  type = string
}

variable "contentful_environment_id" {
  type    = string
  default = "master"
}

resource "contentful_tag" "example" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id
  tag_id         = "terraform-example"

  name       = "Terraform example"
  visibility = "private"
}
