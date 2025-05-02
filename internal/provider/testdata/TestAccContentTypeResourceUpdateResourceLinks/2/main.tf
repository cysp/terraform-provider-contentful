resource "contentful_content_type" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = var.test_content_type_id

  name        = "Test"
  description = "Test content type (${var.test_content_type_id})"

  display_field = "name"

  fields = [
    {
      id        = "name"
      name      = "Name"
      type      = "Symbol"
      required  = true
      localized = false
    },
    {
      id        = "slug"
      name      = "Slug"
      type      = "Symbol"
      required  = true
      localized = false
      validations = [
        jsonencode({
          regexp = {
            pattern = "^[a-z0-9]+(?:-[a-z0-9]+)*$"
            flags   = null
          }
        }),
      ]
    },
    {
      id   = "flags"
      name = "Flags"
      type = "Array"
      items = {
        type = "Symbol"
        validations = [
          jsonencode({
            in = ["abc", "def", "ghi"]
          }),
        ]
      }
      default_value = jsonencode({
        "en-AU" : ["def"],
      })
      required  = true
      localized = false
    },
    {
      id   = "external_resource"
      name = "External Resource"
      type = "ResourceLink"
      allowed_resources = [{
        external = {
          type = "external-resource-type"
        }
      }]
      required  = false
      localized = false
    },
    {
      id   = "cross_space_link"
      name = "Cross-space Link"
      type = "ResourceLink"
      allowed_resources = [{
        contentful_entry = {
          source = "cross-space"
          content_types = [
            "cross-space-content-type-abc",
            "cross-space-content-type-def",
          ]
        }
      }]
      required  = false
      localized = false
    },
  ]
}

resource "contentful_editor_interface" "test" {
  space_id        = var.space_id
  environment_id  = var.environment_id
  content_type_id = var.test_content_type_id

  depends_on = [contentful_content_type.test]
}
