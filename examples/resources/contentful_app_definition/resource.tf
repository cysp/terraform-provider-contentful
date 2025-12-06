resource "contentful_app_definition" "this" {
  organization_id = var.contentful_organization_id

  name = "My app"

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
