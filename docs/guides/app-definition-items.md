---
page_title: "App Definition array items"
subcategory: ""
description: |-
  Read App Definition array item definitions and update list-based expressions.
---

# App Definition array items

In the `contentful_app_definition` and `contentful_marketplace_app_definition`
data sources, each field type's `items` attribute is an object, or null when
Contentful omits the item definition. `locations` and `field_types` are lists
and can also be null.

Check `field.items != null` before accessing `field.items.type` or
`field.items.link_type`. The [App Definition](../data-sources/app_definition)
and [Marketplace App Definition](../data-sources/marketplace_app_definition)
examples show how to traverse nullable parent lists and item definitions.

## Migrating to v0.0.69

Provider versions v0.0.34 through v0.0.68 declare `items` as a list in both data
sources. Starting with v0.0.69, `items` is an object. When upgrading from an affected
version to v0.0.69 or later, update consuming expressions:

| v0.0.34–v0.0.68 | v0.0.69 and later |
| --- | --- |
| `field.items[0].type` | `field.items.type` |
| `field.items[0].link_type` | `field.items.link_type` |

Replace iteration over `items` with object access, retaining the null check.
The managed resource's `items` object is unchanged.
