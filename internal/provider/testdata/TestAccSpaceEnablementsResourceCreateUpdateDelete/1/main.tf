resource "contentful_space_enablements" "test" {
  space_id = var.space_id

  space_templates   = false
  cross_space_links = false
}
