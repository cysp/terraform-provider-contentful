list "contentful_content_type" "content_types" {
  provider = contentful

  config {
    space_id       = var.contentful_space_id
    environment_id = var.contentful_environment_id
  }
}
