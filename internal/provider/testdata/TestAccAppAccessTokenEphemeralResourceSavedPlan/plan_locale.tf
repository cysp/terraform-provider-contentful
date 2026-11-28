data "contentful_locales" "plan" {
  provider       = contentful.editorial
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  lifecycle {
    precondition {
      condition     = timecmp(ephemeral.contentful_app_access_token.editorial.expires_at, plantimestamp()) > 0
      error_message = "The app access token must be valid at plan time."
    }
  }
}

output "planned_locale" {
  value = one(data.contentful_locales.plan.locales).code
}
