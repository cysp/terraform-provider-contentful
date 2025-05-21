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

  metadata = {
    annotations = jsonencode({
      ContentType = [
        {
          id = "aaa"
        },
        {
          id = "bbb"
          parameters = {
            b = "c"
          }
        },
      ]
      ContentTypeField = {
        "name" = [
          {
            id = "ccc"
          },
          {
            id = "ddd"
            parameters = {
              d = "e"
            }
          },
        ]
      }
    })
    taxonomy = [
      {
        taxonomy_concept_scheme = {
          id = "a"
        }
      },
      {
        taxonomy_concept = {
          id       = "b"
          required = true
        }
      },
    ]
  }
}
