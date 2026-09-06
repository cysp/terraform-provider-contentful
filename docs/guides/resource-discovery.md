---
page_title: "Discover and import existing resources"
description: |-
  Discover existing Contentful Entries and Content Types with Terraform list resources.
---

# Discover and import existing resources

The provider supports list resources for [`contentful_entry`](../list-resources/entry) and [`contentful_content_type`](../list-resources/content_type), using Terraform 1.14 or later. They discover existing objects for import. Managed resources control those objects' lifecycles; data sources provide the read-only lookups listed in the provider reference.

## Query Contentful

With the [provider configured](../index), put the query in a `.tfquery.hcl` file. This example finds Entries of an existing `blogPost` Content Type in the `master` environment of `SPACE_ID`:

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

For Content Types, use `list "contentful_content_type"` with the same space and environment arguments. The Entry list resource also accepts Contentful filters through `query` and ordering through `order`; see its schema for those arguments. The provider handles CMA pagination, while Terraform's list `limit` controls the maximum number of results.

`terraform query` returns matching identities. `terraform query -generate-config-out=generated.tf` also generates resource and import blocks. See HashiCorp's [bulk import workflow](https://developer.hashicorp.com/terraform/language/import/bulk) for initialization, query syntax, and configuration generation.

## Import behavior

Generated configuration needs review against the selected resource's schema and lifecycle. Import alone does not publish an external Entry draft or activate an external Content Type draft. Later managed changes follow the [`contentful_entry`](../resources/entry) and [`contentful_content_type`](../resources/content_type) lifecycle contracts.

For one known object, use the Import section on its resource page.
