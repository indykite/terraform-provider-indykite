resource "indykite_application_agent" "agent" {
  application_id  = "ApplicationGID"
  name            = "terraform-agent"
  display_name    = "Terraform agent"
  description     = "Agent for terraform configuration"
  api_permissions = ["Authorization", "Capture"]
}
