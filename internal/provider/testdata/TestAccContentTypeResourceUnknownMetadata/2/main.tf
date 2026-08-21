resource "terraform_data" "metadata" {
  input = {
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
        sections = [
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
    taxonomy = []
  }
}

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
      id            = "sections"
      name          = "Sections"
      type          = "Array"
      localized     = false
      required      = false
      disabled      = false
      omitted       = false
      link_type     = null
      default_value = null
      validations   = []
      items = {
        type        = "Link"
        link_type   = "Entry"
        validations = []
      }
    },
  ]

  metadata = terraform_data.metadata.output
}
