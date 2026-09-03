data "indykite_audit_signings" "example" {
  app_space_id = indykite_application_space.my_space.id
  filter       = ["terraform-audit-signing"]
}

output "audit_signing_provider" {
  value = data.indykite_audit_signings.example.audit_signings[0].key_provider
}
