# Prerequisites: resources referenced by the examples below.
resource "indykite_application_space" "my_space" {
  customer_id = "gid:ZXhhbXBsZS1vcmdhbml6YXRpb24"
  name        = "my-project"
  region      = "us-east1"
}

resource "indykite_application" "my_application" {
  app_space_id = indykite_application_space.my_space.id
  name         = "my-application"
}

resource "indykite_application_agent" "my_agent" {
  application_id  = indykite_application.my_application.id
  name            = "my-agent"
  api_permissions = ["Authorization"]
}

resource "indykite_token_introspect" "my_introspect" {
  project_id    = indykite_application_space.my_space.id
  name          = "my-token-introspect"
  ikg_node_type = "Person"
  jwt_matcher {
    issuer   = "https://issuer.example.com"
    audience = "my-audience"
  }
  online_validation {
    cache_ttl = 600
  }
}

# Example 1: Minimal MCP Server with hardcoded IDs
resource "indykite_mcp_server" "minimal" {
  name                = "terraform-mcp-server"
  project_id          = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
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
  project_id          = indykite_application_space.my_space.id
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
  project_id          = indykite_application_space.my_space.id
  app_agent_id        = indykite_application_agent.my_agent.id
  token_introspect_id = indykite_token_introspect.my_introspect.id
  scopes_supported    = ["read", "write", "admin"]
  enabled             = false
}

# Note: The project_id parameter accepts an Application Space ID. location is deprecated, use project_id instead.
# app_agent_id and token_introspect_id must point at existing resources in the same project.
# scopes_supported must contain at least one OAuth scope.
# The MCP server will automatically populate app_space_id and customer_id as computed fields.
