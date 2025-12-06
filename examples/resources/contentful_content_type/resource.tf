resource "contentful_content_type" "author" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  content_type_id = "author"

  name        = "Author"
  description = "An author of content"

  display_field = "name"

  fields = [
    {
      id        = "name"
      name      = "Name"
      type      = "Symbol"
      disabled  = false
      localized = false
      omitted   = false
      required  = true
    },
    {
      id        = "avatar"
      name      = "Avatar"
      type      = "Link"
      link_type = "Asset"
      disabled  = false
      localized = false
      omitted   = false
      required  = false
    },
  ]
}
