resource "contentful_preview_environment" "test" {
  space_id               = var.space_id
  preview_environment_id = var.preview_environment_id
  name                   = var.name

  content_type_configurations = var.include_page ? {
    page = {
      url = "https://preview.example.invalid/{entry.sys.id}"
    }
  } : {}
}
