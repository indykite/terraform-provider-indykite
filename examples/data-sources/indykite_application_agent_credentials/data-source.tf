# List the application agent credentials of a project.
data "indykite_application_agent_credentials" "example" {
  app_space_id = "gid:ZXhhbXBsZS1wcm9qZWN0"
}

output "application_agent_credential_kids" {
  value = [for c in data.indykite_application_agent_credentials.example.app_agent_credentials : c.kid]
}
