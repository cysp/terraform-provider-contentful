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
    },
    {
      id        = "sections"
      name      = "Sections"
      type      = "Array"
      required  = false
      localized = false
      items = {
        type      = "Link"
        link_type = "Entry"
      }
    }
  ]

  metadata = {
    annotations = jsonencode({
      ContentType = [
        {
          sys = {
            id       = "Contentful:AggregateRoot"
            type     = "Link"
            linkType = "Annotation"
          }
        },
      ]
      ContentTypeField = {
        "sections" = [
          {
            sys = {
              id       = "Contentful:AggregateComponent"
              type     = "Link"
              linkType = "Annotation"
            }
          },
        ]
      }
    })
  }
}
