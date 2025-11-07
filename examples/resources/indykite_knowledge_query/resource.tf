# Example - authorization policy for CIQ
resource "indykite_authorization_policy" "policy_for_ciq" {
  name         = "terraform-pipeline-policy-for-ciq"
  display_name = "Terraform policy for CIQ"
  description  = "Policy for CIQ in terraform pipeline"
  json = jsonencode({
    "meta" : { "policy_version" : "1.0-ciq" },
    "subject" : { "type" : "Person" },
    "condition" : {
      "cypher" : "MATCH (subject:Person)-[r1:ACCEPTED]->(contract:Contract)-[r2:COVERS]->(vehicle:Vehicle)-[r3:HAS]->(ln:LicenseNumber)",
      "filter" : [{ "app" : "app1", "attribute" : "subject.property.username", "operator" : "=", "value" : "$username" }]
    },
    "allowed_reads" : {
      "nodes" : ["ln.property.value", "ln.property.transferrable", "ln.external_id"],
      "relationships" : ["r1"]
    }
  })
  location = "AppSpaceID"
  status   = "active"
}

# Example - knowledge query
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

# Example 1: Minimal knowledge query with hardcoded location
resource "indykite_knowledge_query" "minimal_query" {
  name      = "minimal-query"
  location  = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  status    = "active"
  policy_id = indykite_authorization_policy.policy_for_ciq.id
  query = jsonencode({
    "nodes" : ["Person.property.email"]
  })
}

# Example 2: Knowledge query with reference to application_space
resource "indykite_knowledge_query" "query_with_ref" {
  name         = "query-with-reference"
  display_name = "Query with AppSpace Reference"
  description  = "Knowledge query using application space reference"
  location     = indykite_application_space.my_space.id
  status       = "active"
  policy_id    = indykite_authorization_policy.policy_for_ciq.id
  query = jsonencode({
    "nodes" : ["User.property.name", "User.property.email"],
    "relationships" : ["BELONGS_TO"]
  })
}

# Example 3: Knowledge query with policy reference
resource "indykite_knowledge_query" "query_with_policy" {
  name         = "query-with-policy"
  display_name = "Query with Policy"
  description  = "Knowledge query with authorization policy"
  location     = indykite_application_space.my_space.id
  status       = "active"
  policy_id    = indykite_authorization_policy.my_policy.id
  query = jsonencode({
    "nodes" : ["Document.property.title", "Document.property.content"],
    "relationships" : ["OWNS", "CAN_READ"],
    "filter" : { "attribute" : "Document.property.status", "operator" : "=", "value" : "published" }
  })
}

# Example 4: Complex knowledge query with filters
resource "indykite_knowledge_query" "complex_query" {
  name         = "complex-query"
  display_name = "Complex Knowledge Query"
  description  = "Query with complex filtering and relationships"
  location     = indykite_application_space.my_space.id
  status       = "active"
  policy_id    = indykite_authorization_policy.my_policy.id
  query = jsonencode({
    "nodes" : [
      "Person.property.name",
      "Person.property.email",
      "Organization.property.name"
    ],
    "relationships" : ["WORKS_FOR", "MANAGES"],
    "filter" : {
      "attribute" : "Person.property.status",
      "operator" : "=",
      "value" : "$status"
    }
  })
}

# Example 5: Inactive knowledge query
resource "indykite_knowledge_query" "inactive_query" {
  name         = "inactive-query"
  display_name = "Inactive Query"
  description  = "Knowledge query that is not currently active"
  location     = indykite_application_space.my_space.id
  status       = "inactive"
  policy_id    = indykite_authorization_policy.policy_for_ciq.id
  query = jsonencode({
    "nodes" : ["Asset.property.value"]
  })
}

# Note: The location parameter accepts an Application Space ID.
# status can be either "active" or "inactive".
# policy_id is optional and references an authorization policy.
# The query will automatically populate app_space_id and customer_id as computed fields.
