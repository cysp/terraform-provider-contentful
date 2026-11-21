resource "contentful_entry" "example" {
  space_id        = contentful_content_type.author.space_id
  environment_id  = contentful_content_type.author.environment_id
  content_type_id = contentful_content_type.author.content_type_id

  fields = {
    name = { "en-US" = jsonencode("Ursula Le Guin") }
  }
}
