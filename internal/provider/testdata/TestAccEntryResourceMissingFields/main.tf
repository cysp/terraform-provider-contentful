variable "space_id" {
  type = string
}

variable "environment_id" {
  type = string
}

variable "test_id" {
  type = string
}

variable "entry_fields" {
  type    = map(string)
  default = {}
}

resource "contentful_content_type" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id

  content_type_id = var.test_id

  name        = "Test Content Type ${var.test_id}"
  description = "TestAccEntryResourceMissingFields"

  display_field = "a"

  fields = [
    {
      id        = "a"
      name      = "A"
      type      = "Symbol"
      localized = false
      required  = false
    },
    {
      id        = "b"
      name      = "B"
      type      = "Symbol"
      localized = false
      required  = false
    },
    {
      id   = "c"
      name = "C"
      type = "Array"
      items = {
        type = "Symbol"
      }
      localized = false
      required  = false
    },
  ]
}

resource "contentful_entry" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id

  content_type_id = contentful_content_type.test.content_type_id

  fields = var.entry_fields
}
