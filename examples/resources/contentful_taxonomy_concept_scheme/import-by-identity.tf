import {
  identity = {
    organization_id   = var.contentful_organization_id
    concept_scheme_id = var.concept_scheme_id
  }
  to = contentful_taxonomy_concept_scheme.products
}
