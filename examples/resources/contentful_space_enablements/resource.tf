resource "contentful_space_enablements" "this" {
  space_id = var.contentful_space_id

  cross_space_links = true
}
