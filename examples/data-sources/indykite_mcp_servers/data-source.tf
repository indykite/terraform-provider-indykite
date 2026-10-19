# List the MCP servers of a project with the given names.
data "indykite_mcp_servers" "example" {
  app_space_id = "gid:ZXhhbXBsZS1wcm9qZWN0"
  filter       = ["my-mcp-server"]
}

output "mcp_server_ids" {
  value = [for s in data.indykite_mcp_servers.example.mcp_servers : s.id]
}
