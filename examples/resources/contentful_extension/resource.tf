resource "contentful_extension" "example" {
  space_id       = var.space_id
  environment_id = var.environment_id
  extension_id   = "custom-field-extension"

  extension = {
    name = "Custom Field Extension"
    src  = "https://example.com/custom-field-extension.html"

    field_types = [
      {
        type = "Symbol"
      }
    ]

    sidebar = false

    parameters = {
      installation = [
        {
          id   = "apiKey"
          name = "API Key"
          type = "Secret"
        }
      ]
      instance = [
        {
          id   = "placeholder"
          name = "Placeholder Text"
          type = "Symbol"
        }
      ]
    }
  }

  parameters = jsonencode({
    apiKey = "your-api-key-here"
  })
}
