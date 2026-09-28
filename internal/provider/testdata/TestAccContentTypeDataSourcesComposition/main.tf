resource "terraform_data" "scope" {
  input = "space"
}

data "contentful_content_types" "all" {
  space_id       = terraform_data.scope.output
  environment_id = "master"
}

locals {
  article = one([for content_type in data.contentful_content_types.all.content_types : content_type if content_type.content_type_id == "article"])
}

data "contentful_content_type" "article" {
  space_id        = terraform_data.scope.output
  environment_id  = "master"
  content_type_id = local.article.content_type_id
}

resource "contentful_entry" "using_lookup" {
  space_id        = terraform_data.scope.output
  environment_id  = "master"
  content_type_id = data.contentful_content_type.article.content_type_id
  entry_id        = "looked-up-entry"
  fields = {
    title = jsonencode({ "en-US" = "From discovery" })
  }
}
