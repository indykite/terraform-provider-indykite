# List the service accounts of an organization with the given names.
data "indykite_service_accounts" "example" {
  customer_id = "gid:ZXhhbXBsZS1vcmdhbml6YXRpb24"
  filter      = ["my-service-account"]
}

output "service_account_ids" {
  value = [for a in data.indykite_service_accounts.example.service_accounts : a.id]
}
