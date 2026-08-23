resource "contentful_tag" "example" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id
  tag_id         = "example"

  name       = "Example"
  visibility = "private"
}
