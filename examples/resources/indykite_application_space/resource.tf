resource "indykite_application_space" "appspace" {
  customer_id    = "CustomerGID"
  name           = "AppSpaceName"
  display_name   = "Terraform appspace"
  description    = "Application space for terraform configuration"
  region         = "us-east1"
  ikg_size       = "4GB"
  replica_region = "us-west1"
}
