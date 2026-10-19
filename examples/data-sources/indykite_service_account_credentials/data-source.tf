# List the service account credentials of an organization.
data "indykite_service_account_credentials" "example" {
  customer_id = "gid:ZXhhbXBsZS1vcmdhbml6YXRpb24"
}

output "service_account_credential_kids" {
  value = [for c in data.indykite_service_account_credentials.example.service_account_credentials : c.kid]
}
