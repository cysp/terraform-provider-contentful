resource "contentful_taxonomy_concept" "furniture" {
  organization_id = var.contentful_organization_id
  concept_id      = "furniture"

  pref_label = {
    "en-US" = "Furniture"
  }
}

resource "contentful_taxonomy_concept_scheme" "products" {
  organization_id   = var.contentful_organization_id
  concept_scheme_id = "products"

  pref_label = {
    "en-US" = "Products"
  }

  top_concept_ids = [contentful_taxonomy_concept.furniture.concept_id]
  concept_ids     = [contentful_taxonomy_concept.furniture.concept_id]
}
