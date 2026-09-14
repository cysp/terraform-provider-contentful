data "contentful_spaces" "accessible" {
  organization_id = var.contentful_organization_id
}

# Omit organization_id to discover Spaces across all accessible organizations.
