resource "indykite_application_agent_credential" "with_public" {
  app_agent_id = indykite_application_agent.agent.id
  display_name = "Key with custom private-public key pair"
  expire_time  = "2026-12-31T12:34:56-01:00"
}

