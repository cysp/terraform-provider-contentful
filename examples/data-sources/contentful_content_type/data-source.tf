data "contentful_content_type" "article" {
  space_id        = var.contentful_space_id
  environment_id  = var.contentful_environment_id
  content_type_id = var.contentful_content_type_id
}

# Use data.contentful_content_type.article.content_type_id in an Entry or
# Editor Interface configuration. The fields describe the current CMA model.
