resource "contentful_delivery_api_key" "test" {
  space_id = var.space_id

  name        = var.test_delivery_api_key_name
  description = "key: ${var.test_delivery_api_key_name}"

  environments = [var.environment_id]
}

data "contentful_preview_api_key" "test" {
  space_id = var.space_id

  preview_api_key_id = contentful_delivery_api_key.test.preview_api_key_id
}
