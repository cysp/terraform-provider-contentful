variable "space_id" {
  type = string
}

variable "environment_id" {
  type = string
}

variable "test_id" {
  type = string
}

variable "entry_concepts" {
  type    = list(string)
  default = null
}

resource "contentful_content_type" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id

  content_type_id = var.test_id

  name        = "Test Content Type ${var.test_id}"
  description = "TestAccEntryResourceMetadataConcepts"

  display_field = "name"

  fields = [
    {
      id        = "name"
      name      = "Name"
      type      = "Symbol"
      localized = false
      required  = false
    },
  ]
}

resource "contentful_entry" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id

  content_type_id = contentful_content_type.test.content_type_id

  fields = {
    name = jsonencode({
      "en-AU" = "Test Entry with Taxonomy Concepts"
    })
  }

  metadata = {
    concepts = var.entry_concepts
  }
}
