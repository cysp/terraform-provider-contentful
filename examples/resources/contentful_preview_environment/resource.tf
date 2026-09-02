resource "contentful_preview_environment" "this" {
  space_id    = var.contentful_space_id
  name        = "Website preview"
  description = "Draft website"

  content_type_configurations = {
    page = {
      url = "https://preview.example.com/{env_id}/pages/{entry.sys.id}"
    }
    article = {
      url = "https://preview.example.com/{env_id}/articles/{entry.fields.slug}"
    }
  }
}

resource "contentful_preview_environment" "selected_id" {
  space_id               = var.contentful_space_id
  preview_environment_id = var.preview_environment_id
  name                   = "Website preview with selected ID"

  content_type_configurations = {
    page = {
      url = "https://preview.example.com/{env_id}/pages/{entry.sys.id}"
    }
  }
}
