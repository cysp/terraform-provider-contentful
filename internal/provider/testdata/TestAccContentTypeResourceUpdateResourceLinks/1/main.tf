resource "contentful_content_type" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = var.test_content_type_id

  name        = "Test"
  description = "Test content type (${var.test_content_type_id})"

  display_field = "name"

  fields = [
    {
      id        = "name"
      name      = "Name"
      type      = "Symbol"
      required  = true
      localized = false
    }
  ]
}

resource "contentful_editor_interface" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = var.test_content_type_id

  depends_on = [contentful_content_type.test]
}
