---
page_title: "Upgrade App Definition array item expressions"
subcategory: ""
description: |-
  Migrate App Definition data-source expressions from provider v0.0.34–v0.0.68.
---

# Upgrade App Definition array item expressions

The `contentful_app_definition` and `contentful_marketplace_app_definition`
data sources declare Array `items` as a list in v0.0.34–v0.0.68 and as an object
from v0.0.69. When upgrading from an affected
version to v0.0.69 or later, update consuming expressions:

| v0.0.34–v0.0.68 | v0.0.69 and later |
| --- | --- |
| `field.items[0].type` | `field.items.type` |
| `field.items[0].link_type` | `field.items.link_type` |

Replace iteration over `items` with object access, retaining the null check.
See the [App Definition example](../data-sources/app_definition) for nullable
locations and field types.
The managed resource's `items` object is unchanged.
