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

resource "indykite_authorization_policy" "policy_for_ciq" {
  name         = "terraform-pipeline-policy-for-ciq"
  display_name = "Terraform policy for CIQ"
  description  = "Policy for CIQ in terraform configuration"
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
    "allowed_upserts" : { # omitted if empty
      "nodes" : {         # omitted if empty
        "existing_nodes" : [
          "<string>"
        ],
        "node_labels" : [
          "<string>"
        ]
      },
      "relationships" : { # omitted if empty
        "existing_relationships" : [
          "<string>"
        ],
        "relationship_types" : [ # omitted if empty
          {
            "type" : "<string>",
            "source_node_label" : "<string>",
            "target_node_label" : "<string>"
          }
        ]
      }
    },
    "allowed_deletes" : { # omitted if empty
      "nodes" : ["<string>"],
      "relationships" : ["<string>"]
    },
  })
  location = indykite_application_space.appspace.id
  status   = "active"
}
