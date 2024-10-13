resource "contentful_delivery_api_key" "this" {
  space_id = local.contentful_space_id

  name = "Content Delivery API Key"
}
