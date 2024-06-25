resource "contentful_role" "editor" {
  space_id = local.contentful_space_id

  name        = "Editor"
  description = null

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
    {
      effect  = "allow"
      actions = ["all"]
      constraint = jsonencode({
        and = [
          { equals = [{ doc = "sys.type" }, "Asset"] },
        ]
      })
    },
  ]
}
