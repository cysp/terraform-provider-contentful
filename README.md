# Terraform Provider Contentful

A Terraform provider for managing Contentful spaces and related configuration in a consistent, repeatable way.

[![Version](https://img.shields.io/github/v/release/cysp/terraform-provider-contentful?display_name=tag&sort=semver&logo=terraform&label=version)](https://registry.terraform.io/providers/cysp/contentful)
[![Tests](https://github.com/cysp/terraform-provider-contentful/actions/workflows/test.yml/badge.svg)](https://github.com/cysp/terraform-provider-contentful/actions/workflows/test.yml)
[![Code Coverage](https://codecov.io/gh/cysp/terraform-provider-contentful/graph/badge.svg)](https://codecov.io/gh/cysp/terraform-provider-contentful)

## Scope and Coverage

- A broad set of Contentful resources, including content types, entries, environments, webhooks, and roles.
- Standard Terraform workflows for reviewable, versioned changes.

## Documentation

- Terraform Registry: [cysp/contentful](https://registry.terraform.io/providers/cysp/contentful)

## Getting Started

```terraform
terraform {
  required_providers {
    contentful = {
      source = "cysp/contentful"
    }
  }
}

provider "contentful" {
  access_token = var.contentful_access_token
}
```

## Resources

- [`contentful_app_definition`](docs/resources/app_definition.md)
- [`contentful_app_installation`](docs/resources/app_installation.md)
- [`contentful_app_signing_secret`](docs/resources/app_signing_secret.md)
- [`contentful_content_type`](docs/resources/content_type.md)
- [`contentful_delivery_api_key`](docs/resources/delivery_api_key.md)
- [`contentful_editor_interface`](docs/resources/editor_interface.md)
- [`contentful_entry`](docs/resources/entry.md)
- [`contentful_environment`](docs/resources/environment.md)
- [`contentful_environment_alias`](docs/resources/environment_alias.md)
- [`contentful_extension`](docs/resources/extension.md)
- [`contentful_personal_access_token`](docs/resources/personal_access_token.md)
- [`contentful_resource_provider`](docs/resources/resource_provider.md)
- [`contentful_resource_type`](docs/resources/resource_type.md)
- [`contentful_role`](docs/resources/role.md)
- [`contentful_space_enablements`](docs/resources/space_enablements.md)
- [`contentful_tag`](docs/resources/tag.md)
- [`contentful_team`](docs/resources/team.md)
- [`contentful_team_space_membership`](docs/resources/team_space_membership.md)
- [`contentful_webhook`](docs/resources/webhook.md)

## List Resources

- [`contentful_content_type`](docs/list-resources/content_type.md)
- [`contentful_entry`](docs/list-resources/entry.md)

## Data Sources

- [`contentful_app_definition`](docs/data-sources/app_definition.md)
- [`contentful_environment_status_ready`](docs/data-sources/environment_status_ready.md)
- [`contentful_marketplace_app_definition`](docs/data-sources/marketplace_app_definition.md)
- [`contentful_preview_api_key`](docs/data-sources/preview_api_key.md)

## License

Licensed under the Mozilla Public License 2.0. See [LICENSE](LICENSE).
