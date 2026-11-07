# Deploy the app separately and replace src with its HTTPS URL.
resource "contentful_app_definition" "this" {
  organization_id = var.contentful_organization_id

  name = "Editorial tools"
  src  = "https://app.example.com"

  locations = [
    { location = "app-config" },
    {
      location = "entry-field"
      field_types = [
        { type = "Symbol" },
        { type = "Link", link_type = "Entry" },
        { type = "Array", items = { type = "Symbol" } },
        { type = "Array", items = { type = "Link", link_type = "Asset" } },
      ]
    },
    {
      location        = "page"
      navigation_item = { name = "Editorial tools", path = "/editorial-tools" }
    },
  ]

  parameters = {
    installation = [
      {
        id      = "refreshInterval"
        name    = "Refresh interval"
        type    = "Number"
        default = jsonencode(42)
      },
      {
        id      = "enabled"
        name    = "Enabled"
        type    = "Boolean"
        default = jsonencode(false)
      },
      {
        id      = "theme"
        name    = "Theme"
        type    = "Enum"
        options = [jsonencode({ light = "Light" }), jsonencode("dark")]
        default = jsonencode("light")
      },
      {
        id   = "accessToken"
        name = "Access Token"
        type = "Secret"
      },
    ]
  }
}
