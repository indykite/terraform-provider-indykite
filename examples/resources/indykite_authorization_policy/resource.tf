resource "indykite_authorization_policy" "policy_drive_car" {
  name         = "terraform-pipeline-policy-drive-car"
  display_name = "Terraform policy drive car"
  description  = "Policy for terraform pipeline"
  json = jsonencode({
    meta = {
      policyVersion = "1.0-indykite"
    },
    subject = {
      type = "Person"
    },
    actions = ["CAN_DRIVE"],
    resource = {
      type = "Car"
    },
    condition = {
      cypher = "MATCH (subject:Person)-[:OWNS]->(resource:Car)"
    }
  })
  location = "AppSpaceID"
  status   = "active"
}
