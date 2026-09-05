# Terraform Provider Contentful

A Terraform provider for managing Contentful configuration and content within existing spaces in a consistent, repeatable way.

[![Version](https://img.shields.io/github/v/release/cysp/terraform-provider-contentful?display_name=tag&sort=semver&logo=terraform&label=version)](https://registry.terraform.io/providers/cysp/contentful)
[![Tests](https://github.com/cysp/terraform-provider-contentful/actions/workflows/test.yml/badge.svg)](https://github.com/cysp/terraform-provider-contentful/actions/workflows/test.yml)
[![Code Coverage](https://codecov.io/gh/cysp/terraform-provider-contentful/graph/badge.svg)](https://codecov.io/gh/cysp/terraform-provider-contentful)

## Scope and coverage

- Manages Contentful resources including content types, entries, environments, webhooks, roles, app configuration, and related space-scoped settings.
- Uses standard Terraform workflows for reviewable, versioned changes.
- Operates on existing Contentful spaces; it does not create or delete spaces.

Selected Contentful mutations are deliberately not replayed after an ambiguous
transport error, rate-limit response, or ordinary 5xx response when replay could
repeat a non-idempotent operation. If Contentful created an object with a
generated identity but Terraform never received that identity, a later apply can
still create a duplicate. See the
[Contentful HTTP retry policy](docs/design/contentful-http-retry-policy.md) for
the exact boundary.

## Documentation

- Practitioner reference: [cysp/contentful on the Terraform Registry](https://registry.terraform.io/providers/cysp/contentful)
- Provider design: [Terraform value semantics](docs/design/terraform-value-semantics.md) and
  [Contentful HTTP retry policy](docs/design/contentful-http-retry-policy.md)
- Development workflow: [DEVELOPMENT.md](DEVELOPMENT.md)
- Release workflow: [docs/releasing.md](docs/releasing.md)

## Getting started

Declare the provider in your Terraform configuration:

```terraform
terraform {
  required_providers {
    contentful = {
      source = "cysp/contentful"
    }
  }
}

provider "contentful" {}
```

Supply a Contentful Management API token through the environment rather than
hard-coding it in configuration:

```sh
export CONTENTFUL_MANAGEMENT_ACCESS_TOKEN='...'
terraform init
terraform plan
```

The provider also accepts `access_token` in its configuration. In production,
use an intentional provider version constraint appropriate to your upgrade
policy.

## License

Licensed under the Mozilla Public License 2.0. See [LICENSE](LICENSE).
