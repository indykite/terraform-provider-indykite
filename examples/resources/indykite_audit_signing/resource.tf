# Example 1: Platform managed signing key
resource "indykite_audit_signing" "platform_managed" {
  name         = "terraform-audit-signing"
  project_id   = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  key_provider = "PLATFORM_MANAGED"
}

# Example 2: Customer managed key in Google Cloud KMS
# Customer managed keys are stored but not yet used: records are signed with the platform key for now.
resource "indykite_audit_signing" "gcp_kms" {
  name         = "terraform-audit-signing-gcp"
  display_name = "Audit signing with Cloud KMS"
  description  = "Audit records will be signed with a key hosted in the customer's Cloud KMS"
  project_id   = indykite_application_space.my_space.id
  key_provider = "CUSTOMER_GCP_KMS"
  # the full crypto key VERSION name; the key version must use EC_SIGN_P256_SHA256
  key_resource = "projects/my-project/locations/europe-west1/keyRings/audit/cryptoKeys/signing/cryptoKeyVersions/1"
  kid          = "audit-signing-2026"
  # each auth_params key and value is limited to 256 characters: pass identifiers and
  # short secrets, never a key file or a PEM private key
  auth_params = {
    client_email = "audit-signer@my-project.iam.gserviceaccount.com"
  }
}

# Example 3: Customer managed key in AWS KMS
# Customer managed keys are stored but not yet used: records are signed with the platform key for now.
resource "indykite_audit_signing" "aws_kms" {
  name         = "terraform-audit-signing-aws"
  project_id   = indykite_application_space.my_space.id
  key_provider = "CUSTOMER_AWS_KMS"
  # the key ARN; its region segment selects the KMS region. ECC_NIST_P256, ECDSA_SHA_256
  key_resource = "arn:aws:kms:eu-west-1:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab"
  kid          = "audit-signing-aws"
  # an IAM role with an external ID, as AWS recommends for third-party access,
  # instead of long-term access keys
  auth_params = {
    role_arn    = "arn:aws:iam::123456789012:role/indykite-audit-signer"
    external_id = var.indykite_external_id
  }
}

# Note: The project_id parameter accepts an Application Space ID. location is deprecated, use project_id instead.
# key_provider is always required; key_resource, kid and auth_params are only
# needed for CUSTOMER_* providers. CUSTOMER_AZURE_KEY_VAULT is accepted but has no signing support yet.
# auth_params values are write-only: the API never returns them, so Terraform keeps
# the configured values in state and only reconciles the set of keys.
