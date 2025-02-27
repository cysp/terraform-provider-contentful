resource "contentful_content_type" "test" {
  space_id        = "0p38pssr0fi3"
  environment_id  = "nonexistent"
  content_type_id = "nonexistent"

  name        = ""
  description = ""

  display_field = ""

  fields = []
}
