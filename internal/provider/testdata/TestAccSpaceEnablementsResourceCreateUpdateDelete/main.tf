resource "contentful_space_enablements" "test" {
  space_id = var.space_id

  cross_space_links = var.cross_space_links
  space_templates   = var.space_templates
}
