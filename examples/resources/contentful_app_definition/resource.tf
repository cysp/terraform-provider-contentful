# Deploy the app separately and replace src with its HTTPS URL.
resource "contentful_app_definition" "this" {
  organization_id = var.contentful_organization_id

  name = "Editorial tools"
  src  = "https://app.example.com"

  locations = [
    { location = "app-config" },
  ]

  parameters = {
    installation = [
      {
        id   = "accessToken"
        name = "Access Token"
        type = "Secret"
      },
    ]
  }
}
