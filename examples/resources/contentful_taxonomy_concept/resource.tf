resource "contentful_taxonomy_concept" "furniture" {
  organization_id = var.contentful_organization_id
  concept_id      = "furniture"

  pref_label = {
    "en-US" = "Furniture"
  }

  alt_labels = {
    "en-US" = ["Furnishings"]
  }

  notations = ["FURNITURE"]
}
