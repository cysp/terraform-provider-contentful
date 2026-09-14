resource "terraform_data" "scope" {
  input = "space"
}

data "contentful_spaces" "test" {
  organization_id = "org"
}
data "contentful_space" "test" {
  space_id = terraform_data.scope.output
}
data "contentful_environments" "test" {
  space_id = data.contentful_space.test.space_id
}
data "contentful_environment_aliases" "test" {
  space_id = data.contentful_space.test.space_id
}
data "contentful_environment_alias" "test" {
  space_id             = data.contentful_space.test.space_id
  environment_alias_id = "master"
}
data "contentful_environment" "test" {
  space_id       = data.contentful_space.test.space_id
  environment_id = data.contentful_environment_alias.test.target_environment_id
}
data "contentful_locales" "test" {
  space_id       = data.contentful_space.test.space_id
  environment_id = data.contentful_environment_alias.test.environment_alias_id
  lifecycle {
    postcondition {
      condition     = length([for locale in self.locales : locale if locale.default]) == 1
      error_message = "Exactly one default Locale is required."
    }
  }
}
locals {
  default_locale = one([for locale in data.contentful_locales.test.locales : locale if locale.default])
}
data "contentful_locale" "test" {
  space_id       = data.contentful_space.test.space_id
  environment_id = data.contentful_environment_alias.test.environment_alias_id
  locale_id      = local.default_locale.locale_id
}
resource "contentful_entry" "localized" {
  space_id        = data.contentful_space.test.space_id
  environment_id  = data.contentful_environment.test.environment_id
  content_type_id = "article"
  entry_id        = "example"
  fields = {
    title = jsonencode({ (local.default_locale.code) = "Hello" })
  }
}
