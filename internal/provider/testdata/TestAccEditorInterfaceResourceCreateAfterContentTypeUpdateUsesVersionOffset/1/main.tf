resource "contentful_content_type" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = var.content_type_id

  name          = "Author"
  description   = "Author"
  display_field = "name"

  fields = [{
    id        = "name"
    name      = "Name"
    type      = "Symbol"
    localized = false
    required  = true
  }]
}
