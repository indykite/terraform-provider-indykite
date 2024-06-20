resource "indykite_application_agent_credential" "with-public" {
  app_agent_id   = data.indykite_application_agent.opa.id
  display_name   = "Key with custom private-public key pair"
  expire_time    = "2040-12-31T12:34:56-01:00"
  public_key_jwk = <<-EOT
  {
    "kty": "EC",
    "use": "sig",
    "crv": "P-256",
    "kid": "sig-1631087814",
    "x": "WZQ9LMC08l-L05oxRa-9ObmaPQlTuWHX2GaAmMgAuSE",
    "y": "V5VuhYvyQ2ACiVznB_aqzfVWmwfhvPFD6Dc4X32WUE8",
    "alg": "ES256"
  }
  EOT
}
