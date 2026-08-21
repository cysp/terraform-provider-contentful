resource "contentful_content_type" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = var.test_content_type_id

  name          = "Published field removal test"
  description   = "Published field removal test (${var.test_content_type_id})"
  display_field = "title"

  fields = [
    {
      id        = "title"
      name      = "Title"
      type      = "Symbol"
      required  = true
      localized = false
    }
  ]
}
