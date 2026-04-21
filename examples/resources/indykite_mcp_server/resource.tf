# Example 1: Minimal MCP Server with hardcoded IDs
resource "indykite_mcp_server" "minimal" {
  name                = "terraform-mcp-server"
  location            = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  app_agent_id        = "gid:AAAABWluZHlraURlgAAFDwAAAAA"
  token_introspect_id = "gid:AAAABnRva2VuSW50cm9zcGVjdAAA"
  scopes_supported    = ["read"]
  enabled             = true
}

# Example 2: MCP Server with references and full metadata
resource "indykite_mcp_server" "with_refs" {
  name                = "terraform-mcp-server-full"
  display_name        = "Terraform MCP Server"
  description         = "MCP Server exposing knowledge graph access to AI agents"
  location            = indykite_application_space.my_space.id
  app_agent_id        = indykite_application_agent.my_agent.id
  token_introspect_id = indykite_token_introspect.my_introspect.id
  scopes_supported    = ["read", "write"]
  enabled             = true
}

# Example 3: Disabled MCP Server with broader scope list
resource "indykite_mcp_server" "disabled" {
  name                = "terraform-mcp-server-disabled"
  display_name        = "Disabled MCP Server"
  description         = "Temporarily disabled while the underlying agent is being rotated"
  location            = indykite_application_space.my_space.id
  app_agent_id        = indykite_application_agent.my_agent.id
  token_introspect_id = indykite_token_introspect.my_introspect.id
  scopes_supported    = ["read", "write", "admin"]
  enabled             = false
}

# Note: The location parameter accepts an Application Space ID.
# app_agent_id and token_introspect_id must point at existing resources in the same project.
# scopes_supported must contain at least one OAuth scope.
# The MCP server will automatically populate app_space_id and customer_id as computed fields.
