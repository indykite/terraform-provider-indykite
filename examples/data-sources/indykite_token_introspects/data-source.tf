# List the token introspect configurations of a project with the given names.
data "indykite_token_introspects" "example" {
  app_space_id = "gid:ZXhhbXBsZS1wcm9qZWN0"
  filter       = ["my-token-introspect"]
}

output "token_introspect_ids" {
  value = [for t in data.indykite_token_introspects.example.token_introspects : t.id]
}
