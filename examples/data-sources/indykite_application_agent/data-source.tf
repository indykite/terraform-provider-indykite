# Example - all parameters shown (still valid)
data "indykite_application_agent" "application_agent" {
  app_space_id = "ApplicationSpaceGID"
  name         = "AppAgentName"
}

# Example 2: Read application agent by app_agent_id only
data "indykite_application_agent" "agent_by_app_agent_id" {
  app_agent_id = "gid:AAAABWluZHlraURlgAADDwAAAAA"
}

# Example 3: Read application agent by name with reference to application_space
data "indykite_application_agent" "agent_by_name_ref" {
  app_space_id = indykite_application_space.my_space.id
  name         = "my-agent"
}

# Example outputs:
#
# output "agent_app_space_id" {
#   value = data.indykite_application_agent.agent_by_id.app_space_id
# }
#
# output "agent_api_permissions" {
#   value = data.indykite_application_agent.agent_by_id.api_permissions
# }
