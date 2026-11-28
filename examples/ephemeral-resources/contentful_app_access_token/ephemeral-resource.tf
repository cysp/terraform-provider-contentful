variable "app_private_key" {
  type      = string
  sensitive = true
  ephemeral = true
}

ephemeral "contentful_app_jwt" "editorial" {
  app_definition_id = var.app_definition_id
  private_key       = var.app_private_key
}

ephemeral "contentful_app_access_token" "editorial" {
  space_id          = var.contentful_space_id
  environment_id    = var.contentful_environment_id
  app_definition_id = var.app_definition_id
  jwt               = ephemeral.contentful_app_jwt.editorial.jwt
}

provider "contentful" {
  alias        = "editorial"
  access_token = ephemeral.contentful_app_access_token.editorial.token
}

resource "contentful_entry" "author" {
  provider        = contentful.editorial
  space_id        = var.contentful_space_id
  environment_id  = var.contentful_environment_id
  content_type_id = "author"

  fields = {
    name = jsonencode({ (var.default_locale_code) = "Ursula Le Guin" })
  }
}
