# Example 1: List applications using app_space_id
data "indykite_applications" "apps_by_app_space_id" {
  app_space_id = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  filter       = ["my-app", "another-app"]
}

# Example 2: List applications with reference to application_space
data "indykite_applications" "apps_by_ref" {
  app_space_id = indykite_application_space.my_space.id
  filter       = ["my-app"]
}

# Example usage:
#
# output "applications_list" {
#   value = {
#     app_space_id = data.indykite_applications.apps_by_app_space_id.app_space_id
#     count        = length(data.indykite_applications.apps_by_app_space_id.applications)
#     apps = [
#       for app in data.indykite_applications.apps_by_app_space_id.applications : {
#         id           = app.id
#         name         = app.name
#         app_space_id = app.app_space_id
#       }
#     ]
#   }
# }
