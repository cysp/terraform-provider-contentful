resource "contentful_editor_interface" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = var.content_type_id

  editor_layout = [{
    group_id = "content"
    name     = "Content"
    items = [
      jsonencode(
        {
          groupId = "name"
          name    = "name"
          items = [
            {
              fieldId = "name"
            },
          ]
        }
      ),
      jsonencode(
        {
          groupId = "bio"
          name    = "Bio"
          items = [
            {
              fieldId = "avatar"
            },
            {
              fieldId = "blurb"
            },
          ]
        }
      ),
    ]
  }]

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
      group_id         = "name"
      widget_namespace = "builtin"
      widget_id        = "fieldset"
      settings = jsonencode({
        collapsedByDefault = false
        helpText           = ""
      })
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
