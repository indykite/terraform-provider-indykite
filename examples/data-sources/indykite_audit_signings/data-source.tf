# Prerequisites: resources referenced by the examples below.
resource "indykite_application_space" "my_space" {
  customer_id = "gid:ZXhhbXBsZS1vcmdhbml6YXRpb24"
  name        = "my-project"
  region      = "us-east1"
}

resource "indykite_audit_signing" "platform_managed" {
  project_id   = indykite_application_space.my_space.id
  name         = "platform-managed"
  key_provider = "PLATFORM_MANAGED"
}

data "indykite_audit_signings" "example" {
  app_space_id = indykite_application_space.my_space.id
  filter       = [indykite_audit_signing.platform_managed.name]
  depends_on   = [indykite_audit_signing.platform_managed]
}

output "audit_signing_provider" {
  value = data.indykite_audit_signings.example.audit_signings[0].key_provider
}
