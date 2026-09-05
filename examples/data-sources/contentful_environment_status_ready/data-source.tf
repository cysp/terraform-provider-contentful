data "contentful_environment_status_ready" "example" {
  space_id       = "your-space-id"
  environment_id = "staging-yyyy-mm-dd"

  timeouts = {
    read = "15m"
  }
}
