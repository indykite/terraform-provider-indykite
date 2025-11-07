# Example 1: List application agents using app_space_id
data "indykite_application_agents" "agents_by_app_space_id" {
  app_space_id = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  filter       = ["my-agent", "another-agent"]
}

# Example 2: List application agents with reference to application_space
data "indykite_application_agents" "agents_by_ref" {
  app_space_id = indykite_application_space.my_space.id
  filter       = ["my-agent"]
}

# Example usage:
#
# output "application_agents_list" {
#   value = {
#     app_space_id = data.indykite_application_agents.agents_by_app_space_id.app_space_id
#     count        = length(data.indykite_application_agents.agents_by_app_space_id.app_agents)
#     agents = [
#       for agent in data.indykite_application_agents.agents_by_app_space_id.app_agents : {
#         id              = agent.id
#         name            = agent.name
#         app_space_id    = agent.app_space_id
#         application_id  = agent.application_id
#         api_permissions = agent.api_permissions
#       }
#     ]
#   }
# }
