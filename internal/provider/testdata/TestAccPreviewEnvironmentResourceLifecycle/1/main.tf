resource "contentful_preview_environment" "test" {
  space_id = var.space_id
  name     = var.name

  content_type_configurations = {
    page = {
      url = "https://preview.example.invalid/{env_id}/pages/{entry.sys.id}"
    }
  }
}
