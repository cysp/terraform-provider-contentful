resource "contentful_content_type" "test_a_a" {
  space_id        = "space-a"
  environment_id  = "environment-a-a"
  content_type_id = var.content_type_id

  name        = "Test (${var.content_type_id})"
  description = "Test (${var.content_type_id})"

  display_field = "name"

  fields = concat(
    [
      {
        id          = "name"
        name        = "Name"
        type        = "Symbol"
        required    = true
        localized   = false
        validations = []
      },
    ],
    [for id in var.content_type_additional_fields : {
      id        = id
      name      = upper(id)
      type      = "Symbol"
      localized = false
      disabled  = false
      omitted   = false
      required  = true
    }],
  )
}

resource "contentful_editor_interface" "test_a_a" {
  space_id        = "space-a"
  environment_id  = "environment-a-a"
  content_type_id = contentful_content_type.test_a_a.content_type_id

  controls = concat(
    [
      {
        field_id         = "name"
        widget_id        = "singleLine"
        widget_namespace = "builtin"
      },
    ],
    [for id in var.content_type_additional_fields : {
      field_id         = id
      widget_id        = "singleLine"
      widget_namespace = "builtin"
    }],
  )
}

resource "contentful_content_type" "test_a_b" {
  space_id        = "space-a"
  environment_id  = "environment-a-b"
  content_type_id = var.content_type_id

  name        = "Test (${var.content_type_id})"
  description = "Test (${var.content_type_id})"

  display_field = "name"

  fields = concat(
    [
      {
        id          = "name"
        name        = "Name"
        type        = "Symbol"
        required    = true
        localized   = false
        validations = []
      },
    ],
    [for id in var.content_type_additional_fields : {
      id        = id
      name      = upper(id)
      type      = "Symbol"
      localized = false
      disabled  = false
      omitted   = false
      required  = true
    }],
  )
}

resource "contentful_editor_interface" "test_a_b" {
  space_id        = "space-a"
  environment_id  = "environment-a-b"
  content_type_id = contentful_content_type.test_a_b.content_type_id

  controls = concat(
    [
      {
        field_id         = "name"
        widget_id        = "singleLine"
        widget_namespace = "builtin"
      },
    ],
    [for id in var.content_type_additional_fields : {
      field_id         = id
      widget_id        = "singleLine"
      widget_namespace = "builtin"
    }],
  )
}

resource "contentful_content_type" "test_b_a" {
  space_id        = "space-b"
  environment_id  = "environment-b-a"
  content_type_id = var.content_type_id

  name        = "Test (${var.content_type_id})"
  description = "Test (${var.content_type_id})"

  display_field = "name"

  fields = concat(
    [
      {
        id          = "name"
        name        = "Name"
        type        = "Symbol"
        required    = true
        localized   = false
        validations = []
      },
    ],
    [for id in var.content_type_additional_fields : {
      id        = id
      name      = upper(id)
      type      = "Symbol"
      localized = false
      disabled  = false
      omitted   = false
      required  = true
    }],
  )
}

resource "contentful_editor_interface" "test_b_a" {
  space_id        = "space-b"
  environment_id  = "environment-b-a"
  content_type_id = contentful_content_type.test_b_a.content_type_id

  controls = concat(
    [
      {
        field_id         = "name"
        widget_id        = "singleLine"
        widget_namespace = "builtin"
      },
    ],
    [for id in var.content_type_additional_fields : {
      field_id         = id
      widget_id        = "singleLine"
      widget_namespace = "builtin"
    }],
  )
}
