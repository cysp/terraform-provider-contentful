# This snippet uses contentful_delivery_api_key.this from the Delivery API Key example.
data "contentful_preview_api_key" "this" {
  space_id = contentful_delivery_api_key.this.space_id

  preview_api_key_id = contentful_delivery_api_key.this.preview_api_key_id
}
