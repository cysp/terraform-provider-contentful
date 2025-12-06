resource "contentful_editor_interface" "author" {
  space_id        = contentful_content_type.author.space_id
  environment_id  = contentful_content_type.author.environment_id
  content_type_id = contentful_content_type.author.content_type_id

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
  ]

  sidebar = [
    {
      widget_namespace = "sidebar-builtin"
      widget_id        = "publication-widget"
      settings         = jsonencode({})
      disabled         = false
    },
    {
      widget_namespace = "sidebar-builtin"
      widget_id        = "incoming-links-widget"
      settings         = jsonencode({})
      disabled         = false
    },
    {
      widget_namespace = "sidebar-builtin"
      widget_id        = "translation-widget"
      settings         = jsonencode({})
      disabled         = false
    },
    {
      widget_namespace = "sidebar-builtin"
      widget_id        = "versions-widget"
      settings         = jsonencode({})
      disabled         = false
    },
    {
      widget_namespace = "app"
      widget_id        = var.cool_app_definition_id
      settings = jsonencode({
        foo = "bar"
      })
    },
  ]
}
