terraform {
  required_providers {
    contentful = {
      source = "cysp/contentful"
    }
  }
}

provider "contentful" {}

variable "contentful_space_id" { type = string }

variable "contentful_environment_id" {
  type    = string
  default = "master"
}

data "contentful_locales" "existing" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  lifecycle {
    postcondition {
      condition     = length([for locale in self.locales : locale if locale.default]) == 1
      error_message = "Expected one default Locale in the selected environment."
    }
  }
}

locals {
  selected_locale = one([for locale in data.contentful_locales.existing.locales : locale if locale.default])
}

output "locale_id" {
  value = local.selected_locale.locale_id
}

output "locale_code" {
  value = local.selected_locale.code
}
