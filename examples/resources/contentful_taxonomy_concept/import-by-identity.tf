import {
  identity = {
    organization_id = var.contentful_organization_id
    concept_id      = var.concept_id
  }
  to = contentful_taxonomy_concept.furniture
}
