list "contentful_content_type" "content_types" {
  provider = contentful

  config {
    space_id       = var.space_id
    environment_id = var.environment_id
  }
}
