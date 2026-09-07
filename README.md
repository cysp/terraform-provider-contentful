# Terraform Provider Contentful

A Terraform provider for managing Contentful content and configuration in existing spaces, plus supported organization and account configuration.

[![Version](https://img.shields.io/github/v/release/cysp/terraform-provider-contentful?display_name=tag&sort=semver&logo=terraform&label=version)](https://registry.terraform.io/providers/cysp/contentful)
[![Tests](https://github.com/cysp/terraform-provider-contentful/actions/workflows/test.yml/badge.svg)](https://github.com/cysp/terraform-provider-contentful/actions/workflows/test.yml)
[![Code Coverage](https://codecov.io/gh/cysp/terraform-provider-contentful/graph/badge.svg)](https://codecov.io/gh/cysp/terraform-provider-contentful)

The provider does not create or delete Contentful spaces.

## Get started

Use an existing Contentful space and a [Content Management API access
token](https://www.contentful.com/developers/docs/references/authentication/#the-content-management-api)
with permission to manage the objects in your configuration. Set the token in
the `CONTENTFUL_MANAGEMENT_ACCESS_TOKEN` environment variable, then declare the
provider:

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

Add resources from the [provider reference](https://registry.terraform.io/providers/cysp/contentful/latest/docs),
then run `terraform init` and `terraform plan` to review the changes before
applying. To adopt existing objects, follow the Import section on their resource
pages. Entry and Content Type resources publish or activate their managed
changes; review their lifecycle guidance before applying.

## Documentation

| You want to… | Start here |
| --- | --- |
| Configure the released provider or look up a resource | [Terraform Registry reference](https://registry.terraform.io/providers/cysp/contentful/latest/docs) |
| Read documentation matching this checkout | [Provider overview](docs/index.md), [resources](docs/resources/), and [data sources](docs/data-sources/) |
| Discover existing Entries and Content Types for import | [Resource discovery guide](docs/guides/resource-discovery.md) |
| Understand credentials, redaction, and imported secrets | [Secrets and Terraform state](docs/guides/secrets-and-state.md) |
| Build, test, or change the provider and its docs | [Development workflow](DEVELOPMENT.md) |
| Understand implementation contracts and their evidence | [Design documentation map](docs/design/README.md) |
| Publish a release | [Release workflow](docs/releasing.md) |

The Registry documents released versions. Documentation and
[reference examples](examples/) in this checkout can include unreleased changes.

## License

Licensed under the Mozilla Public License 2.0. See [LICENSE](LICENSE).
