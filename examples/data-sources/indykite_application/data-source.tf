# Prerequisites: resources referenced by the examples below.
resource "indykite_application_space" "my_space" {
  customer_id = "gid:ZXhhbXBsZS1vcmdhbml6YXRpb24"
  name        = "my-project"
  region      = "us-east1"
}

# Example - all parameters shown (still valid)
data "indykite_application" "application" {
  app_space_id = "gid:ZXhhbXBsZS1wcm9qZWN0"
  name         = "my-application"
}

# Example 2: Read application by application_id only
data "indykite_application" "application_by_app_id" {
  application_id = "gid:AAAABGluZHlraURlgAACDwAAAAA"
}

# Example 3: Read application by name with reference to application_space
data "indykite_application" "application_by_name_ref" {
  app_space_id = indykite_application_space.my_space.id
  name         = "my-application"
}

# Note: You can output the app_space_id if needed:
#
# output "app_space_id" {
#   value = data.indykite_application.application_by_id.app_space_id
# }
