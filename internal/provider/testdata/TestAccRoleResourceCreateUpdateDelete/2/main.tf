resource "contentful_role" "test" {
  space_id = var.space_id

  name = "Test"

  permissions = {
    ContentDelivery    = ["all"]
    ContentModel       = ["read"]
    EnvironmentAliases = []
    Environments       = []
    Settings           = []
    Tags               = []
  }

  policies = [
    {
      effect  = "allow"
      actions = ["all"]
      constraint = jsonencode({
        and = [
          { equals = [{ doc = "sys.type" }, "Entry"] },
        ]
      })
    },
  ]
}
