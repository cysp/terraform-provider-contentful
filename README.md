# Terraform Provider Contentful

A Terraform provider for managing Contentful content and configuration in existing spaces, plus supported organization and account configuration.

[![Version](https://img.shields.io/github/v/release/cysp/terraform-provider-contentful?display_name=tag&sort=semver&logo=terraform&label=version)](https://registry.terraform.io/providers/cysp/contentful)
[![Tests](https://github.com/cysp/terraform-provider-contentful/actions/workflows/test.yml/badge.svg)](https://github.com/cysp/terraform-provider-contentful/actions/workflows/test.yml)
[![Code Coverage](https://codecov.io/gh/cysp/terraform-provider-contentful/graph/badge.svg)](https://codecov.io/gh/cysp/terraform-provider-contentful)

The provider does not create or delete Contentful spaces.

## Configuration

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

Set `CONTENTFUL_MANAGEMENT_ACCESS_TOKEN` in the environment for Contentful
Management API authentication. See the [latest released provider reference](https://registry.terraform.io/providers/cysp/contentful/latest/docs) for
configuration options and the [examples in this checkout](examples/resources/) for
configuration and import syntax.

## Documentation

- Latest released provider reference and guides: [cysp/contentful on the Terraform Registry](https://registry.terraform.io/providers/cysp/contentful/latest/docs)
- Provider design: [design documentation map](docs/design/README.md)
- Development workflow: [DEVELOPMENT.md](DEVELOPMENT.md)
- Release workflow: [docs/releasing.md](docs/releasing.md)

## License

Licensed under the Mozilla Public License 2.0. See [LICENSE](LICENSE).
