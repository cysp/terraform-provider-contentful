resource "contentful_app_installation" "test" {
  space_id          = var.space_id
  environment_id    = var.environment_id
  app_definition_id = var.test_app_definition_id
}

import {
  id = "${var.space_id}/${var.environment_id}/${var.test_app_definition_id}"
  to = contentful_app_installation.test_dup
}

resource "contentful_app_installation" "test_dup" {
  space_id          = var.space_id
  environment_id    = var.environment_id
  app_definition_id = var.test_app_definition_id
}
