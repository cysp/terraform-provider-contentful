resource "contentful_editor_interface" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = var.content_type_id

  editor_layout = [
    {
      group = {
        group_id = "content"
        name     = "Content"
        items = [
          {
            field = { field_id = "name" }
          },
          {
            group = {
              group_id = "bio"
              name     = "Bio"
              items = [
                { field = { field_id = "avatar" } },
                { field = { field_id = "blurb" } },
              ]
            }
          }
        ]
      }
    },
  ]

  controls = [
    {
      field_id         = "name",
      widget_namespace = "builtin",
      widget_id        = "singleLine",
    },
    {
      field_id         = "avatar",
      widget_namespace = "builtin",
      widget_id        = "assetLinkEditor",
    },
    {
      field_id         = "blurb",
      widget_namespace = "builtin",
      widget_id        = "richTextEditor",
    }
  ]

  group_controls = [
    {
      group_id         = "content"
      widget_namespace = "builtin"
      widget_id        = "topLevelTab"
    },
    {
      group_id         = "bio"
      widget_namespace = "builtin"
      widget_id        = "fieldset"
      settings = jsonencode({
        collapsedByDefault = false
        helpText           = ""
      })
    },
  ]
}
