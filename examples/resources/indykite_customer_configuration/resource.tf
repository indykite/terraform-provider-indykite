resource "indykite_customer_configuration" "development" {
  customer_id = "gid:customer-gid"

  default_auth_flow_id     = "gid:auth-flow-gid"
  default_email_service_id = "gid:email-gid"
}
