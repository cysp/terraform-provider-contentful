resource "contentful_content_type" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = var.content_type_id

  name          = "Author"
  description   = "Author"
  display_field = "name"

  fields = [
    {
      id        = "name"
      name      = "Name"
      type      = "Symbol"
      localized = false
      required  = true
    },
    {
      id        = "biography"
      name      = "Biography"
      type      = "Text"
      localized = false
      required  = false
    },
  ]
}

resource "contentful_editor_interface" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = contentful_content_type.test.content_type_id

  controls = [
    {
      field_id         = "name"
      widget_id        = "singleLine"
      widget_namespace = "builtin"
    },
    {
      field_id         = "biography"
      widget_id        = "multipleLine"
      widget_namespace = "builtin"
    },
  ]
}
