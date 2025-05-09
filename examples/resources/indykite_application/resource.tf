resource "indykite_application" "application" {
  app_space_id = "AppSpaceGID"
  name         = "terraform-application"
  display_name = "Terraform application"
  description  = "Application for terraform configuration"
}
