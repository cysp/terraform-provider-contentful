resource "contentful_editor_interface" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = var.content_type_id

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

  sidebar = [{
    widget_namespace = "app"
    widget_id        = "1WkQ2J9LERPtbMTdUfSHka"
  }]
}
