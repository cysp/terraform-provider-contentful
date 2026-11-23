terraform {
  required_providers {
    contentful = {
      source = "cysp/contentful"
    }
  }
}

resource "contentful_webhook" "test" {
  space_id = "space"
  name     = "Edited externally"
  url      = "https://example.com/webhook"
  topics   = ["Entry.publish"]
  headers  = {}
  filters = [{
    equals = {
      doc   = "sys.type"
      value = "Entry"
    }
  }]
}
