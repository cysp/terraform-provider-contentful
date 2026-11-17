list "contentful_entry" "authors" {
  provider = contentful

  config {
    space_id       = "SPACE_ID"
    environment_id = "master"
    content_type   = "author"
  }
}
