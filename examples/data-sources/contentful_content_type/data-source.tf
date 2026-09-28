data "contentful_content_type" "article" {
  space_id        = var.contentful_space_id
  environment_id  = var.contentful_environment_id
  content_type_id = var.contentful_content_type_id
}
