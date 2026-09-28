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

With the [provider configured](../index#authentication), this example reads an existing Space and Environment, selects the default Locale, and outputs its ID and content code.

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

data "contentful_space" "existing" {
  space_id = var.contentful_space_id
}

data "contentful_environment" "existing" {
  space_id       = data.contentful_space.existing.space_id
  environment_id = var.contentful_environment_id
}

data "contentful_locales" "existing" {
  space_id       = data.contentful_space.existing.space_id
  environment_id = data.contentful_environment.existing.environment_id

  lifecycle {
    postcondition {
      condition     = length([for locale in self.locales : locale if locale.default]) == 1
      error_message = "Exactly one Locale must match the selection."
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

Set `contentful_space_id` and, if needed, `contentful_environment_id`, then run `terraform init` and `terraform plan`. Run `terraform apply` to save the lookup results and outputs in state. These data sources do not change Contentful configuration.

The Environment data source reports status immediately. Use [`contentful_environment_status_ready`](../data-sources/environment_status_ready) when a lookup needs to wait for an Environment to become ready.

## Select exactly one discovered object

Filter collection results by name, locale code, or another returned attribute. To select a locale by code, use `locale.code == "en-GB"` in both the postcondition and the `one(...)` expression above. Comparisons are case-sensitive.

The postcondition requires exactly one match before dependent expressions use it: `one([])` returns null, while multiple matches produce an error. Apply the same pattern to `self.spaces` and `space.name` to select a Space by name, using `organization_id` to narrow the search when needed.

Use `locale_id` for a Locale lookup and `code` for localized content. To manage a discovered Locale, import it into the [`contentful_locale` resource](../resources/locale). For bulk import and configuration generation, use the [Locale list resource](../list-resources/locale).

Use `role_id` from a Role data source when assigning roles with [`contentful_team_space_membership`](../resources/team_space_membership).

## Reference a Content Type

When the ID is known, use `contentful_content_type` and pass its `content_type_id` to an Entry or Editor Interface configuration. To discover an ID by name, filter `contentful_content_types.content_types` and require exactly one match with a postcondition and `one(...)`, as in the [plural example](../data-sources/content_types). Names are mutable, so use the Contentful system ID for the dependent lookup.

Both data sources read the current CMA Content Type model, which may include unactivated changes. `published_version` reports the most recently activated version when present; it does not establish that the returned fields are the activated model. The lookups do not activate, publish, import, or change Content Types. Import a discovered Content Type into the [`contentful_content_type` resource](../resources/content_type) to manage it.

## Use environment aliases

An Environment lookup accepts an alias ID and returns its target as `aliased_environment_id` when Contentful supplies it. An Environment Alias lookup returns `target_environment_id`. Using the resolved target ID makes dependent requests address that environment directly.

Alias lookups refresh on later plans, so a resolved target can change. To keep lookups tied to one environment across alias changes, use a concrete environment ID.

Locale lookups also accept an environment alias ID as `environment_id`. The provider currently supports responses whose environment link identifies that alias; other responses produce an identity error. Use the target environment ID if you encounter this error.

Content Type lookups retain the requested `environment_id` and require the returned environment link to echo it. An alias-routed Content Type response with a different link produces an unsupported response-identity error. Content Type response-link behavior through aliases has not been verified live; use a concrete environment ID when the alias route cannot meet this check.

## Collection reads

The two-minute `timeouts.read` default covers the whole collection read. Concurrent edits or alias changes can affect results between pages, so a collection is not an atomic snapshot. See [Operation timeouts](operation-timeouts) to adjust the read deadline.
