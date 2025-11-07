# Example 1: Minimal service account credential with hardcoded service_account_id
resource "indykite_service_account_credential" "minimal_credential" {
  service_account_id = "gid:AAAABGluZHlraURlgAACDwAAAAA"
}

# Example 2: Service account credential with reference to service account
resource "indykite_service_account_credential" "credential_with_ref" {
  service_account_id = indykite_service_account.my_service_account.id
}

# Example 3: Service account credential with display name
resource "indykite_service_account_credential" "credential_with_display_name" {
  service_account_id = indykite_service_account.my_service_account.id
  display_name       = "Production API Credential"
}

# Example 4: Service account credential with custom expiration time
resource "indykite_service_account_credential" "credential_with_expiration" {
  service_account_id = indykite_service_account.my_service_account.id
  display_name       = "Temporary Credential"
  expire_time        = "2025-12-31T23:59:59Z"
}

# Example 5: Service account credential with 90-day expiration
resource "indykite_service_account_credential" "credential_90_days" {
  service_account_id = indykite_service_account.my_service_account.id
  display_name       = "90-Day Credential"
  expire_time        = timeadd(timestamp(), "2160h") # 90 days = 2160 hours
}

# Example 6: Service account credential with 1-year expiration
resource "indykite_service_account_credential" "credential_1_year" {
  service_account_id = indykite_service_account.my_service_account.id
  display_name       = "Annual Credential"
  expire_time        = timeadd(timestamp(), "8760h") # 365 days = 8760 hours
}

# Example 7: Multiple credentials for the same service account
resource "indykite_service_account_credential" "dev_credential" {
  service_account_id = indykite_service_account.my_service_account.id
  display_name       = "Development Environment"
}

resource "indykite_service_account_credential" "staging_credential" {
  service_account_id = indykite_service_account.my_service_account.id
  display_name       = "Staging Environment"
}

resource "indykite_service_account_credential" "prod_credential" {
  service_account_id = indykite_service_account.my_service_account.id
  display_name       = "Production Environment"
}

# Example 8: Credential for service account with reference
resource "indykite_service_account_credential" "credential_with_service_account_ref" {
  service_account_id = indykite_service_account.minimal.id
  display_name       = "Credential for Service Account Reference"
}

# Example 10: Saving credential to a file (use with caution!)
resource "indykite_service_account_credential" "credential_to_file" {
  service_account_id = indykite_service_account.my_service_account.id
  display_name       = "Credential Saved to File"
}

# Save the credential configuration to a file
# WARNING: This file will contain sensitive credentials!
# Make sure to add it to .gitignore and handle it securely
resource "local_file" "service_account_config" {
  content         = indykite_service_account_credential.credential_to_file.service_account_config
  filename        = "${path.module}/service-account-config.json"
  file_permission = "0600"
}

# Example 11: Using credential output in other resources
output "service_account_credential_kid" {
  description = "Key identifier of the service account credential"
  value       = indykite_service_account_credential.credential_with_ref.kid
}

output "service_account_credential_id" {
  description = "ID of the service account credential"
  value       = indykite_service_account_credential.credential_with_ref.id
}

# Note: The service_account_config is only available after creation and is sensitive.
# It contains the complete JSON configuration needed to authenticate with IndyKite APIs.
# customer_id is a computed field automatically populated from the service account.
# expire_time must be in RFC3339 format (e.g., "2025-12-31T23:59:59Z").
# The credential cannot be updated after creation - all fields are ForceNew.
