import {
  identity = {
    space_id       = var.contentful_space_id
    environment_id = "staging-yyyy-mm-dd"
  }
  to = contentful_environment.staging
}
