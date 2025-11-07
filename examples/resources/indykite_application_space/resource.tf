# Note: This file uses variables for sensitive values.
# Define var.db_password in your terraform.tfvars or pass it via environment variables.

# Original example - using customer_id (still valid)
resource "indykite_application_space" "appspace" {
  customer_id    = "CustomerGID"
  name           = "AppSpaceName"
  display_name   = "Terraform appspace"
  description    = "Application space for terraform configuration"
  region         = "us-east1"
  ikg_size       = "4GB"
  replica_region = "us-west1"
}

# Example 1: Minimal configuration
resource "indykite_application_space" "minimal" {
  customer_id = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  name        = "minimal-appspace"
  region      = "us-east1"
}

# Example 2: Full configuration
resource "indykite_application_space" "full" {
  customer_id         = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  name                = "full-appspace"
  display_name        = "Full Application Space"
  description         = "Application space with all optional fields"
  region              = "us-east1"
  ikg_size            = "4GB"
  replica_region      = "us-west1"
  deletion_protection = false
}

# Example 3: Using reference to customer resource
resource "indykite_application_space" "appspace_from_customer" {
  customer_id  = data.indykite_customer.my_customer.id
  name         = "appspace-from-customer"
  display_name = "AppSpace from Customer"
  description  = "Application space created from customer reference"
  region       = "europe-west1"
}

# Example 4: With deletion protection enabled
resource "indykite_application_space" "protected_appspace" {
  customer_id         = data.indykite_customer.my_customer.id
  name                = "protected-appspace"
  display_name        = "Protected Application Space"
  description         = "Application space with deletion protection enabled"
  region              = "us-east1"
  deletion_protection = true
}

# Example 5: With database connection configuration for sandbox
resource "indykite_application_space" "appspace_with_db" {
  customer_id  = data.indykite_customer.my_customer.id
  name         = "appspace-with-db"
  display_name = "AppSpace with DB Connection"
  description  = "Application space with database connection configuration"
  region       = "us-east1"
  db_connection {
    url      = "postgresql://db.example.com:5432/indykite"
    username = "dbuser"
    password = var.db_password
    name     = "indykite"
  }
}
