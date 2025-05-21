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
    annotations = {
      content_type = [
        {
          id = "aaa"
        },
        {
          id = "bbb"
          parameters = jsonencode({
            b = "c"
          })
        },
      ]
      content_type_field = {
        "name" = [
          {
            id = "ccc"
          },
          {
            id = "ddd"
            parameters = jsonencode({
              d = "e"
            })
          },
        ]
      }
    }
    taxonomy = [
      {
        concept_scheme = {
          id = "a"
        }
      },
      {
        concept = {
          id       = "b"
          required = true
        }
      },
    ]
  }
}
