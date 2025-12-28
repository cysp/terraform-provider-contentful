resource "contentful_tag" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id
  tag_id         = var.tag_id
  name           = "my-tag"
  visibility     = var.visibility
}
