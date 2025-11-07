# Example - basic credential (still valid)
resource "indykite_application_agent_credential" "with_public" {
  app_agent_id = "ApplicationAgentGID"
  display_name = "Key with custom private-public key pair"
  expire_time  = "2026-12-31T12:34:56-01:00"
}

# Example 1: Minimal configuration with hardcoded app_agent_id
resource "indykite_application_agent_credential" "minimal_credential" {
  app_agent_id = "gid:AAAABWluZHlraURlgAADDwAAAAA"
  display_name = "Minimal Credential"
}

# Example 2: Credential with reference to application agent
resource "indykite_application_agent_credential" "credential_from_agent" {
  app_agent_id = indykite_application_agent.my_agent.id
  display_name = "Credential from Agent Reference"
}

# Example 3: Credential with custom expiration time
resource "indykite_application_agent_credential" "credential_with_expiry" {
  app_agent_id = indykite_application_agent.my_agent.id
  display_name = "Credential with Custom Expiry"
  expire_time  = "2027-12-31T23:59:59Z"
}

# Example 4: Credential with far future expiration
resource "indykite_application_agent_credential" "long_lived_credential" {
  app_agent_id = indykite_application_agent.my_agent.id
  display_name = "Long-lived Credential"
  expire_time  = "2030-01-01T00:00:00Z"
}

# Example 5: Credential for agent created with app_space_id
resource "indykite_application_agent_credential" "credential_for_new_agent" {
  app_agent_id = indykite_application_agent.agent_from_new_app.id
  display_name = "Credential for New Agent"
  expire_time  = "2028-06-30T12:00:00Z"
}

# Note: The credential's kid (key ID) and other details are automatically
# generated and can be referenced in outputs:
#
# output "credential_kid" {
#   value = indykite_application_agent_credential.with_public.kid
# }
#
# output "credential_customer_id" {
#   value = indykite_application_agent_credential.with_public.customer_id
# }
