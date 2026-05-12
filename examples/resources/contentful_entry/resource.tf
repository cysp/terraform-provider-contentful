resource "contentful_entry" "example" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = "blogPost"

  fields = {
    title = {
      "en-AU" = jsonencode("My First Blog Post")
      "en-US" = jsonencode("My First Blog Post")
    }
    body = {
      "en-AU" = jsonencode("This is the content of my first blog post.")
      "en-US" = jsonencode("This is the content of my first blog post.")
    }
    slug = {
      "en-AU" = jsonencode("my-first-blog-post")
      "en-US" = jsonencode("my-first-blog-post")
    }
  }

  metadata = {
    tags = ["blog", "example"]
  }
}
