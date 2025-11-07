# Example - basic application agent (still valid)
resource "indykite_application_agent" "agent" {
  application_id = "ApplicationGID"
  name           = "terraform-agent"
  display_name   = "Terraform agent"
  description    = "Agent for terraform configuration"
}

# Example 1: Minimal configuration with hardcoded application_id
resource "indykite_application_agent" "minimal_agent" {
  application_id = "gid:AAAABGluZHlraURlgAACDwAAAAA"
  name           = "minimal-agent"
}

# Example 2: Full configuration with all optional fields
resource "indykite_application_agent" "full_agent" {
  application_id      = "gid:AAAABGluZHlraURlgAACDwAAAAA"
  name                = "full-agent"
  display_name        = "Full Agent"
  description         = "Agent with all optional fields"
  api_permissions     = ["Authorization", "Capture"]
  deletion_protection = false
}

# Example 3: Agent with reference to application resource
resource "indykite_application_agent" "agent_from_app" {
  application_id  = indykite_application.my_application.id
  name            = "agent-from-app"
  display_name    = "Agent from Application"
  description     = "Agent created from application reference"
  api_permissions = ["Authorization", "Capture", "EntityMatching"]
}

# Example 4: Agent with multiple API permissions
resource "indykite_application_agent" "agent_multi_permissions" {
  application_id  = indykite_application.my_application.id
  name            = "agent-multi-permissions"
  display_name    = "Agent with Multiple Permissions"
  description     = "Agent with all available API permissions"
  api_permissions = ["Authorization", "Capture", "EntityMatching", "IKGRead"]
}

# Example 5: Agent with deletion protection enabled
resource "indykite_application_agent" "protected_agent" {
  application_id      = indykite_application.my_application.id
  name                = "protected-agent"
  display_name        = "Protected Agent"
  description         = "Agent with deletion protection enabled"
  api_permissions     = ["Authorization"]
  deletion_protection = true
}


# These are automatically set when the agent is created, regardless of whether the
# Both fields can be referenced in outputs:
#
# }
#
# output "agent_app_space_id" {
#   value = indykite_application_agent.agent.app_space_id
# }
#
# output "agent_customer_id" {
#   value = indykite_application_agent.agent.customer_id
# }
