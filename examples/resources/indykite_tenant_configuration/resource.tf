resource "indykite_tenant_configuration" "development" {
  tenant_id = "gid:tenant-gid"

  default_auth_flow_id     = "gid:auth-flow-gid"
  default_email_service_id = "gid:email-gid"

  username_policy {
    allowed_username_formats  = ["email"] # Possible also username and mobile
    valid_email               = true
    verify_email              = true
    verify_email_grace_period = "600s" # Use Golang duration syntax https://pkg.go.dev/time#ParseDuration
    allowed_email_domains     = ["gmail.com", "outlook.com"]
    exclusive_email_domains   = ["indykite.com"]
  }
}
