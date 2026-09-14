data "contentful_space" "test" {
  space_id = "0p38pssr0fi3"
}

data "contentful_spaces" "test" {
  organization_id = data.contentful_space.test.organization_id
  lifecycle {
    postcondition {
      condition     = contains([for space in self.spaces : space.space_id], data.contentful_space.test.space_id)
      error_message = "The organization collection must contain the addressed Space."
    }
  }
}

data "contentful_environment_aliases" "test" {
  space_id = data.contentful_space.test.space_id
}

data "contentful_environment" "test" {
  space_id       = data.contentful_space.test.space_id
  environment_id = "master"
}

data "contentful_environments" "test" {
  space_id = data.contentful_space.test.space_id
  lifecycle {
    postcondition {
      condition     = contains([for environment in self.environments : environment.environment_id], data.contentful_environment.test.environment_id)
      error_message = "The Environment collection must contain the addressed Environment."
    }
  }
}

data "contentful_locales" "test" {
  space_id       = data.contentful_space.test.space_id
  environment_id = data.contentful_environment.test.environment_id
  lifecycle {
    postcondition {
      condition     = length([for locale in self.locales : locale if locale.default]) == 1
      error_message = "The acceptance environment must have exactly one default Locale."
    }
  }
}

locals {
  default_locale = one([for locale in data.contentful_locales.test.locales : locale if locale.default])
}

data "contentful_locale" "test" {
  space_id       = data.contentful_space.test.space_id
  environment_id = data.contentful_environment.test.environment_id
  locale_id      = local.default_locale.locale_id
  lifecycle {
    postcondition {
      condition     = self.code == local.default_locale.code && self.fallback_code == local.default_locale.fallback_code
      error_message = "Locale collection and detail must agree on the default Locale's code and fallback."
    }
  }
}
