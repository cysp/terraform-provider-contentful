list "contentful_entry" "entries" {
  provider = contentful

  config {
    space_id       = var.contentful_space_id
    environment_id = var.contentful_environment_id
    content_type   = var.content_type
  }
}
