---
page_title: "Reference existing Contentful configuration"
description: |-
  Use Contentful data sources to compose configuration with objects managed elsewhere.
---

# Reference existing Contentful configuration

Look up existing Contentful configuration by ID, or select an object from a scoped collection:

| Object | Exact lookup | Discovery scope |
| --- | --- | --- |
| Space | [`contentful_space`](../data-sources/space) | [`contentful_spaces`](../data-sources/spaces): one organization or all accessible organizations |
| Environment | [`contentful_environment`](../data-sources/environment) | [`contentful_environments`](../data-sources/environments): one Space |
| Environment Alias | [`contentful_environment_alias`](../data-sources/environment_alias) | [`contentful_environment_aliases`](../data-sources/environment_aliases): one Space |
| Locale | [`contentful_locale`](../data-sources/locale) | [`contentful_locales`](../data-sources/locales): one Space and Environment |
| Role | [`contentful_role`](../data-sources/role) | [`contentful_roles`](../data-sources/roles): one Space |
| Content Type | [`contentful_content_type`](../data-sources/content_type) | [`contentful_content_types`](../data-sources/content_types): one Space and Environment |

## Read existing configuration

Set `contentful_space_id` to your space ID. This configuration looks up the default Locale in `master`. Set `contentful_environment_id` to select another environment.

```terraform
terraform {
  required_providers {
    contentful = {
      source = "cysp/contentful"
    }
  }
}

provider "contentful" {}

variable "contentful_space_id" { type = string }

variable "contentful_environment_id" {
  type    = string
  default = "master"
}

data "contentful_locales" "existing" {
  space_id       = var.contentful_space_id
  environment_id = var.contentful_environment_id

  lifecycle {
    postcondition {
      condition     = length([for locale in self.locales : locale if locale.default]) == 1
      error_message = "Expected one default Locale in the selected environment."
    }
  }
}

locals {
  selected_locale = one([for locale in data.contentful_locales.existing.locales : locale if locale.default])
}

output "locale_id" {
  value = local.selected_locale.locale_id
}

output "locale_code" {
  value = local.selected_locale.code
}
```

Use `locale_id` for a Locale lookup and `code` for localized content. To manage a discovered Locale, import it into the [`contentful_locale` resource](../resources/locale). For bulk import and configuration generation, use the [Locale list resource](../list-resources/locale).

When looking up a Space by name, set `organization_id` if the same name occurs in multiple organizations. Use `role_id` from a Role data source when assigning roles with [`contentful_team_space_membership`](../resources/team_space_membership).

## Reference a Content Type

If you already know the Content Type ID, pass it directly to the Entry or Editor Interface. Use the [Content Type data source](../data-sources/content_type) to read field definitions, or [look up a Content Type by name](../data-sources/content_types) when you need its ID.

Both Content Type data sources may return unactivated changes. `published_version` reports the most recently activated version when present; it does not mean the returned fields match that version. To manage a discovered Content Type, import it into the [`contentful_content_type` resource](../resources/content_type).

## Use environment aliases

An Environment lookup can return an alias's target as `aliased_environment_id`. The Environment Alias data source exposes it as `target_environment_id`. A lookup through the alias can resolve to a different target after someone repoints it. Use a concrete environment ID to keep lookups tied to one environment.

Locale and Content Type lookups check the environment ID in the response. If a lookup through an alias reports an identity error, repeat it with the target environment ID.

## Collection reads

The two-minute `timeouts.read` default covers the whole collection read. Concurrent edits or alias changes can affect results between pages, so a collection is not an atomic snapshot. See [Operation timeouts](operation-timeouts) to adjust the read deadline.
