resource "contentful_app_definition" "test" {
  organization_id = var.organization_id

  name = "Test App Definition"

  locations = []

  parameters = {
    installation = [
      {
        id   = "abc"
        type = "Symbol"
        name = "Test Symbol"
      },
      {
        id   = "def"
        type = "Boolean"
        name = "Test Boolean"
        labels = {
          empty = "Empty"
          true  = "True"
          false = "False"
        }
        default = jsonencode(false)
      },
      {
        id   = "ghi"
        type = "Enum"
        name = "Test Enum"
        options = [
          jsonencode("Option 1"),
          jsonencode("Option 2"),
        ]
      },
    ]
  }
}
