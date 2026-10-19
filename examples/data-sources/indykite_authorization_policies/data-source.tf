# List the KBAC authorization policies of a project with the given names.
data "indykite_authorization_policies" "kbac" {
  app_space_id = "gid:ZXhhbXBsZS1wcm9qZWN0"
  filter       = ["policy-drive-car", "policy-delegated-drive"]
  type         = "kbac" # optional: kbac or ciq, all types when omitted
}

output "kbac_policy_statuses" {
  value = { for p in data.indykite_authorization_policies.kbac.authorization_policies : p.name => p.status }
}
