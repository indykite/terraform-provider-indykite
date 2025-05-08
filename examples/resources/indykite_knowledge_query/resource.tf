resource "indykite_authorization_policy" "policy_for_ciq" {
  name         = "terraform-pipeline-policy-for-ciq"
  display_name = "Terraform policy for CIQ"
  description  = "Policy for CIQ in terraform pipeline"
  json = jsonencode({
    "meta" : { "policy_version" : "1.0-ciq" },
    "subject" : { "type" : "Person" },
    "condition" : {
      "cypher" : "MATCH (person:Person)-[r1:ACCEPTED]->(contract:Contract)-[r2:COVERS]->(vehicle:Vehicle)-[r3:HAS]->(ln:LicenseNumber)",
      "filter" : [{ "app" : "app1", "attribute" : "person.property.username", "operator" : "=", "value" : "$username" }]
    },
    "allowed_reads" : {
      "nodes" : ["ln.property.value", "ln.property.transferrable", "ln.external_id"],
      "relationships" : ["r1"]
    }
  })
  location = "AppSpaceID"
  status   = "active"

}

resource "indykite_knowledge_query" "create-query" {
  name         = "terraform-knowledge-query"
  display_name = "Terraform knowledge-query"
  description  = "Knowledge query for terraform"
  location     = "AppSpaceID"
  query = jsonencode({
    "nodes" : ["ln.property.value"],
    "relationships" : [], 
    "filter" : { "attribute" : "ln.property.value", "operator" : "=", "value" : "$lnValue" }
  })

  status    = "active"
  policy_id = indykite_authorization_policy.policy_for_ciq.id
}