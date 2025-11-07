# Example 1: Minimal service account
resource "indykite_service_account" "minimal" {
  customer_id = "gid:AAAAAmluZHlraURlgAAAAAAAAA"
  name        = "minimal-service-account"
  role        = "all_editor"
}

# Example 2: Full configuration
resource "indykite_service_account" "full" {
  customer_id  = "gid:AAAAAmluZHlraURlgAAAAAAAAA"
  name         = "full-service-account"
  display_name = "Full Service Account"
  description  = "Service account with all optional fields"
  role         = "all_editor"
}

# Example 3: Service account with deletion protection
resource "indykite_service_account" "protected_service_account" {
  customer_id         = "gid:AAAAAmluZHlraURlgAAAAAAAAA"
  name                = "protected-service-account"
  display_name        = "Protected Service Account"
  description         = "Service account with deletion protection enabled"
  role                = "all_editor"
  deletion_protection = true
}

# Example 4: Editor role service account
resource "indykite_service_account" "editor_service_account" {
  customer_id  = "gid:AAAAAmluZHlraURlgAAAAAAAAA"
  name         = "editor-service-account"
  display_name = "Editor Service Account"
  description  = "Service account with all_editor role"
  role         = "all_editor"
}

# Example 5: Viewer role service account
resource "indykite_service_account" "viewer_service_account" {
  customer_id  = "gid:AAAAAmluZHlraURlgAAAAAAAAA"
  name         = "viewer-service-account"
  display_name = "Viewer Service Account"
  description  = "Service account with all_viewer role (read-only)"
  role         = "all_viewer"
}

# Note: role is required and must be either "all_editor" or "all_viewer".
# role cannot be changed after creation (ForceNew).
# The service account will automatically populate create_time and update_time as computed fields.
