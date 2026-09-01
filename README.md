# Terraform Provider Contentful

A Terraform provider for managing Contentful spaces and related configuration in a consistent, repeatable way.

[![Version](https://img.shields.io/github/v/release/cysp/terraform-provider-contentful?display_name=tag&sort=semver&logo=terraform&label=version)](https://registry.terraform.io/providers/cysp/contentful)
[![Tests](https://github.com/cysp/terraform-provider-contentful/actions/workflows/test.yml/badge.svg)](https://github.com/cysp/terraform-provider-contentful/actions/workflows/test.yml)
[![Code Coverage](https://codecov.io/gh/cysp/terraform-provider-contentful/graph/badge.svg)](https://codecov.io/gh/cysp/terraform-provider-contentful)

## Scope and Coverage

- A broad set of Contentful resources, including content types, entries, environments, webhooks, and roles.
- Standard Terraform workflows for reviewable, versioned changes.

The provider does not transparently replay Contentful mutations after an
ambiguous transport error or ordinary 5xx response. A later operation may still
create a duplicate when Contentful generated an identity that Terraform never
received; see the [Contentful HTTP retry policy](docs/design/contentful-http-retry-policy.md)
for the exact boundary.

## Documentation

- Complete provider reference: [cysp/contentful on the Terraform Registry](https://registry.terraform.io/providers/cysp/contentful)
- Provider design: [Terraform value semantics](docs/design/terraform-value-semantics.md) and
  [Contentful HTTP retry policy](docs/design/contentful-http-retry-policy.md)

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

## License

Licensed under the Mozilla Public License 2.0. See [LICENSE](LICENSE).
