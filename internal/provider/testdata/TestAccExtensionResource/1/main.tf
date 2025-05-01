resource "contentful_extension" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id
  extension_id   = var.test_extension_id

  extension = {
    name = var.test_extension_id

    srcdoc = "<!DOCTYPE html>"

    field_types = [
      { type = "Array", items = { type = "Link", link_type = "Entry" } },
    ]
  }
}
