data "contentful_content_types" "existing" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  lifecycle {
    postcondition {
      condition     = length([for content_type in self.content_types : content_type if content_type.name == "Article"]) == 1
      error_message = "Exactly one Content Type named Article is required."
    }
  }
}

locals {
  article = one([for content_type in data.contentful_content_types.existing.content_types : content_type if content_type.name == "Article"])
}

# Use local.article.content_type_id to address the selected Content Type.
