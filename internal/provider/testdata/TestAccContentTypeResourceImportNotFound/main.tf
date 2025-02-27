resource "contentful_content_type" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = "nonexistent"

  name        = ""
  description = ""

  display_field = ""

  fields = []
}
