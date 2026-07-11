resource "contentful_taxonomy_concept" "test" {
  organization_id = var.organization_id
  concept_id      = var.concept_id
  pref_label = {
    "en-US" = var.concept_label
  }
  notations = ["FURNITURE"]
}

resource "contentful_taxonomy_concept_scheme" "test" {
  organization_id  = var.organization_id
  concept_scheme_id = var.concept_scheme_id
  pref_label = {
    "en-US" = var.scheme_label
  }
  top_concept_ids = [contentful_taxonomy_concept.test.concept_id]
  concept_ids     = [contentful_taxonomy_concept.test.concept_id]
}
