resource "indykite_application_space" "appspace" {
  customer_id  = "CustomerGID"
  name         = "AppSpaceName"
  display_name = "Terraform appspace"
  description  = "Application space for terraform configuration"
  region       = "europe-west1"
}
