variable "space_id" {
  type = string
}

data "contentful_roles" "test" {
  space_id = var.space_id

  lifecycle {
    postcondition {
      condition     = length(self.roles) >= 2
      error_message = "Live pagination requires at least two existing Roles in the acceptance Space."
    }
  }
}
