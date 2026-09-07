resource "contentful_live_preview_variables" "this" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  variables = jsonencode({
    host = "https://preview.example.invalid"
    baseurl = {
      # Use a locale enabled in this environment.
      "en-US" = "https://preview.example.invalid/en/"
    }
    optional_value = null
    empty_value    = ""
    localized      = {}
  })
}
