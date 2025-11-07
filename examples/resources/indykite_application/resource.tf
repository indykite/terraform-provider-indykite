# Example - using app_space_id (still valid)
resource "indykite_application" "application" {
  app_space_id = "AppSpaceGID"
  name         = "terraform-application"
  display_name = "Terraform application"
  description  = "Application for terraform configuration"
}

# Example 1: Minimal configuration with app_space_id
resource "indykite_application" "minimal_app_space_id" {
  app_space_id = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  name         = "minimal-application"
}


# Example 3: Full configuration with app_space_id
resource "indykite_application" "full_app_space_id" {
  app_space_id        = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  name                = "full-application"
  display_name        = "Full Application"
  description         = "Application with all optional fields"
  deletion_protection = false
}


# Example 5: Using app_space_id with reference to application_space resource
resource "indykite_application" "app_from_space_old" {
  app_space_id = indykite_application_space.my_space.id
  name         = "application-from-space-old"
  display_name = "Application from Space (Old)"
  description  = "Application created from application space reference using app_space_id"
}


# Example 7: With deletion protection enabled
resource "indykite_application" "protected_application" {
  name                = "protected-application"
  display_name        = "Protected Application"
  description         = "Application with deletion protection enabled"
  deletion_protection = true
}
