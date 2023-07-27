resource "indykite_application_space_configuration" "development" {
  app_space_id = "gid:app-space-gid"

  default_auth_flow_id     = "gid:auth-flow-gid"
  default_email_service_id = "gid:email-gid"
  default_tenant_id        = "gid:tenant-gid"

  username_policy {
    allowed_username_formats  = ["email"] # Possible also username and mobile
    valid_email               = true
    verify_email              = true
    verify_email_grace_period = "600s" # Use Golang duration syntax https://pkg.go.dev/time#ParseDuration
    allowed_email_domains     = ["gmail.com", "outlook.com"]
    exclusive_email_domains   = ["indykite.com"]
  }

  # unique_property_constraints is map, so collon (:) and equal sign (=) is supported after key.
  # Here we used collon to explicitly say, this is a map.
  unique_property_constraints = {
    # Need to specify as JSON, currently only tenantUnique and canonicalization is supported.
    # canonicalization can contain 'unicode' or 'case-insensitive' or both.
    "property1" : jsonencode({ "tenantUnique" : false })
    "super_property" : jsonencode({ "tenantUnique" : true, "canonicalization" : ["unicode"] })
  }
}
