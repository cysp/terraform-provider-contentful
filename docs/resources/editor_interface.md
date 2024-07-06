---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "contentful_editor_interface Resource - terraform-provider-contentful"
subcategory: ""
description: |-
  
---

# contentful_editor_interface (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `content_type_id` (String)
- `environment_id` (String)
- `space_id` (String)

### Optional

- `controls` (Attributes List) (see [below for nested schema](#nestedatt--controls))
- `sidebar` (Attributes List) (see [below for nested schema](#nestedatt--sidebar))

<a id="nestedatt--controls"></a>
### Nested Schema for `controls`

Required:

- `field_id` (String)

Optional:

- `settings` (String)
- `widget_id` (String)
- `widget_namespace` (String)


<a id="nestedatt--sidebar"></a>
### Nested Schema for `sidebar`

Required:

- `widget_id` (String)
- `widget_namespace` (String)

Optional:

- `disabled` (Boolean)
- `settings` (String)

