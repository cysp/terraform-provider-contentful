data "contentful_environment_aliases" "existing" {
  space_id = var.contentful_space_id
}

locals {
  alias_targets = {
    for alias in data.contentful_environment_aliases.existing.environment_aliases :
    alias.environment_alias_id => alias.target_environment_id
  }
}
