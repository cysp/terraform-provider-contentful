resource "contentful_content_type" "author" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = "author"

  name        = "Author"
  description = "An author"

  display_field = "name"

  fields = []
}
