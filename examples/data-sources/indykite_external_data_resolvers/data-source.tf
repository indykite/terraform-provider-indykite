# List the external data resolvers of a project with the given names.
data "indykite_external_data_resolvers" "example" {
  app_space_id = "gid:ZXhhbXBsZS1wcm9qZWN0"
  filter       = ["my-resolver"]
}

output "external_data_resolver_ids" {
  value = [for r in data.indykite_external_data_resolvers.example.external_data_resolvers : r.id]
}
