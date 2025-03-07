resource "contentful_role" "test" {
  space_id    = var.space_id
  name        = "Test"
  permissions = {}
  policies    = []
}

import {
  id = "${var.space_id}/${contentful_role.test.role_id}"
  to = contentful_role.test_dup
}

resource "contentful_role" "test_dup" {
  space_id    = var.space_id
  name        = "Test"
  permissions = {}
  policies    = []
}
