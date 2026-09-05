---
page_title: "Discover and import existing resources"
description: |-
  Use Terraform list resources to find existing Contentful Entries and Content Types and generate import configuration.
---

# Discover and import existing resources

The provider supports Terraform list resources for `contentful_entry` and `contentful_content_type`. List resources require Terraform 1.14 or later and are run with [`terraform query`](https://developer.hashicorp.com/terraform/cli/commands/query).

Use this workflow when you need to discover existing Contentful objects before bringing them under Terraform management. For one known object, the Import section on the corresponding resource page is usually simpler.

## Choose the appropriate Terraform surface

The provider exposes three different Terraform concepts for different jobs:

- **Managed resources** such as `contentful_entry` declare objects that Terraform creates, updates, imports into state, and may destroy according to the resource lifecycle.
- **Data sources** read existing Contentful information for use by a configuration without bringing the returned object under managed-resource lifecycle control. Use the provider's data-source pages for the specific read-only lookups it supports.
- **List resources** query collections of existing objects for discovery. They run through `terraform query` rather than ordinary resource planning and can produce configuration and import blocks that you review before adopting the discovered objects as managed resources.

Use a managed resource when Terraform should own an object's lifecycle, a data source when configuration only needs to read provider-supported existing information, and a list resource when the task is discovery or bulk import.

## Configure the provider

Declare and configure the provider in the normal Terraform `.tf` files in the root configuration directory. Set `CONTENTFUL_MANAGEMENT_ACCESS_TOKEN` in the environment before running Terraform unless you intentionally configure `access_token` directly.

Initialize the working directory before querying:

```shell
terraform init
```

## Define a query

Create a file such as `contentful.tfquery.hcl`. Terraform only accepts `list` blocks in files with the [`.tfquery.hcl` extension](https://developer.hashicorp.com/terraform/language/files/tfquery).

For example, to find Entries of one Content Type:

```terraform
list "contentful_entry" "blog_posts" {
  provider = contentful

  config {
    space_id       = "SPACE_ID"
    environment_id = "master"
    content_type   = "blogPost"
  }
}
```

To list Content Types instead, use `list "contentful_content_type"` with `space_id` and `environment_id`. See the [`contentful_entry`](../list-resources/entry) and [`contentful_content_type`](../list-resources/content_type) list-resource references for provider-specific query arguments.

## Inspect results

Run:

```shell
terraform query
```

Terraform prints the identities of matching resources. The Entry list resource can also pass Contentful collection filters through `query` and ordering expressions through `order`; `skip` and `limit` are controlled by the list operation rather than those provider-specific query parameters.

## Generate and review import configuration

To ask Terraform to generate managed-resource and import blocks for the query results, run:

```shell
terraform query -generate-config-out=generated.tf
```

Terraform requires the output file not to exist already. Review the generated configuration before applying it, including which objects were discovered and the resulting managed-resource arguments. The generated file is a starting point, not a reason to skip normal plan review.

Importing an existing Entry does not by itself publish an external draft, and importing an existing Content Type does not by itself activate an external draft. After import, later Terraform-managed changes follow the lifecycle behavior documented on the [`contentful_entry`](../resources/entry) and [`contentful_content_type`](../resources/content_type) resource pages.

After review, apply the import configuration using the normal Terraform workflow. HashiCorp's [bulk import documentation](https://developer.hashicorp.com/terraform/language/import/bulk) describes the Terraform-level query, generation, and import process in more detail.
