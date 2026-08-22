resource "contentful_preview_environment" "test" {
  space_id    = var.space_id
  name        = "${var.name} updated"
  description = "updated description"

  content_type_configurations = {}
}
