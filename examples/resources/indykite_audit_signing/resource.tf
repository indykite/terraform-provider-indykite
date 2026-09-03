# Example 1: Platform managed signing key (default provider)
resource "indykite_audit_signing" "platform_managed" {
  name     = "terraform-audit-signing"
  location = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
}

# Example 2: Customer managed key in Google Cloud KMS
resource "indykite_audit_signing" "gcp_kms" {
  name         = "terraform-audit-signing-gcp"
  display_name = "Audit signing with Cloud KMS"
  description  = "Audit records are signed with a key hosted in the customer's Cloud KMS"
  location     = indykite_application_space.my_space.id
  key_provider = "CUSTOMER_GCP_KMS"
  key_resource = "projects/my-project/locations/europe-west1/keyRings/audit/cryptoKeys/signing/cryptoKeyVersions/1"
  kid          = "audit-signing-2026"
  auth_params = {
    credentials_json = file("${path.module}/kms-service-account.json")
  }
}

# Example 3: Customer managed key in AWS KMS
resource "indykite_audit_signing" "aws_kms" {
  name         = "terraform-audit-signing-aws"
  location     = indykite_application_space.my_space.id
  key_provider = "CUSTOMER_AWS_KMS"
  key_resource = "arn:aws:kms:eu-west-1:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab"
  kid          = "audit-signing-aws"
  auth_params = {
    access_key_id     = var.aws_access_key_id
    secret_access_key = var.aws_secret_access_key
  }
}

# Note: The location parameter accepts an Application Space ID.
# key_resource, kid and auth_params are only needed for CUSTOMER_* providers.
# auth_params values are write-only: the API never returns them, so Terraform keeps
# the configured values in state and only reconciles the set of keys.
