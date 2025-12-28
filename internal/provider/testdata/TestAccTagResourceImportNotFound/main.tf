resource "contentful_tag" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id
  tag_id         = "nonexistent"
  name           = "Missing"
  visibility     = "private"
}
