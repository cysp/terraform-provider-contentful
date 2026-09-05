data "contentful_environment_status_ready" "test" {
  space_id       = var.space_id
  environment_id = var.environment_id
}
