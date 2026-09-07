list "contentful_entry" "blog_posts" {
  provider = contentful

  config {
    space_id       = "SPACE_ID"
    environment_id = "master"
    content_type   = "blogPost"
  }
}
