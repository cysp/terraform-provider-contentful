resource "contentful_entry" "example" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = "blogPost"

  fields = {
    title = jsonencode({
      "en-AU" = "My First Blog Post"
      "en-US" = "My First Blog Post"
    })
    body = jsonencode({
      "en-AU" = "This is the content of my first blog post."
      "en-US" = "This is the content of my first blog post."
    })
    slug = jsonencode({
      "en-AU" = "my-first-blog-post"
      "en-US" = "my-first-blog-post"
    })
  }

  metadata = {
    tags = ["blog", "example"]
  }
}
