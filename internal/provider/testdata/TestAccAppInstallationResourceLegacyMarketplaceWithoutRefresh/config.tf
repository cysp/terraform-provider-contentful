terraform {
  required_providers {
    contentful = {
      source = "cysp/contentful"
    }
  }
}

resource "contentful_app_installation" "test" {
  space_id          = "space"
  environment_id    = "master"
  app_definition_id = "app"
  parameters        = jsonencode({ secret = "alpha", visible = "one" })
}
