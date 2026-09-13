import {
  identity = {
    organization_id = var.contentful_organization_id
    concept_id      = "furniture"
  }
  to = contentful_taxonomy_concept.furniture
}
