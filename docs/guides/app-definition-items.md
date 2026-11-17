---
page_title: "App Definition array items"
subcategory: ""
description: |-
  Read App Definition array item definitions and update list-based expressions.
---

# App Definition array items

When upgrading from v0.0.34–v0.0.68 to v0.0.69 or later, update expressions that read field types from the [App Definition](../data-sources/app_definition) and [Marketplace App Definition](../data-sources/marketplace_app_definition) data sources. Each field type now has a single `items` object. Contentful can omit it, so check `field.items != null` before accessing its attributes.

| v0.0.34–v0.0.68 | v0.0.69 and later |
| --- | --- |
| `field.items[0].type` | `field.items.type` |
| `field.items[0].link_type` | `field.items.link_type` |

The data-source examples show how to traverse nullable `locations` and `field_types` lists. This migration applies only to the two data sources.
