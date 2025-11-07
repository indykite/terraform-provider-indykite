# Example 1: List application spaces using customer_id
data "indykite_application_spaces" "spaces_by_customer_id" {
  customer_id = "gid:AAAAAWluZHlraURlgAAAAAAAAA8"
  filter      = ["my-space", "another-space"]
}

# Example 2: List application spaces with reference to customer
data "indykite_application_spaces" "spaces_by_ref" {
  customer_id = data.indykite_customer.my_customer.id
  filter      = ["my-space"]
}

# Note: The data source returns customer_id field at the top level
# and in each application space item.
#
# Example usage:
#
# output "application_spaces_list" {
#   value = {
#     customer_id = data.indykite_application_spaces.spaces_by_customer_id.customer_id
#     count       = length(data.indykite_application_spaces.spaces_by_customer_id.app_spaces)
#     spaces      = [
#       for space in data.indykite_application_spaces.spaces_by_customer_id.app_spaces : {
#         id             = space.id
#         name           = space.name
#         customer_id    = space.customer_id
#         region         = space.region
#         ikg_size       = space.ikg_size
#         replica_region = space.replica_region
#       }
#     ]
#   }
# }
