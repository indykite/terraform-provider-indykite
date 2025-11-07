# Example - basic authorization policy (still valid)
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

# Example 1: Minimal policy with hardcoded location
resource "indykite_authorization_policy" "minimal_policy" {
  name     = "minimal-policy"
  location = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  status   = "active"
  json = jsonencode({
    meta = {
      policyVersion = "1.0-indykite"
    },
    subject = {
      type = "User"
    },
    actions = ["READ"],
    resource = {
      type = "Document"
    }
  })
}

# Example 2: Policy with reference to application_space
resource "indykite_authorization_policy" "policy_with_ref" {
  name         = "policy-with-reference"
  display_name = "Policy with AppSpace Reference"
  description  = "Policy using application space reference"
  location     = indykite_application_space.my_space.id
  status       = "active"
  json = jsonencode({
    meta = {
      policyVersion = "1.0-indykite"
    },
    subject = {
      type = "Person"
    },
    actions = ["EDIT"],
    resource = {
      type = "Document"
    },
    condition = {
      cypher = "MATCH (subject:Person)-[:CAN_EDIT]->(resource:Document)"
    }
  })
}

# Example 3: Policy with tags
resource "indykite_authorization_policy" "policy_with_tags" {
  name         = "policy-with-tags"
  display_name = "Policy with Tags"
  description  = "Policy demonstrating tag usage"
  location     = indykite_application_space.my_space.id
  status       = "active"
  tags         = ["production", "critical", "gdpr"]
  json = jsonencode({
    meta = {
      policyVersion = "1.0-indykite"
    },
    subject = {
      type = "Person"
    },
    actions = ["DELETE"],
    resource = {
      type = "SensitiveData"
    },
    condition = {
      cypher = "MATCH (subject:Person)-[:IS_ADMIN]->(org:Organization)"
    }
  })
}

# Example 4: Inactive policy
resource "indykite_authorization_policy" "inactive_policy" {
  name         = "inactive-policy"
  display_name = "Inactive Policy"
  description  = "Policy that is not currently active"
  location     = indykite_application_space.my_space.id
  status       = "inactive"
  json = jsonencode({
    meta = {
      policyVersion = "1.0-indykite"
    },
    subject = {
      type = "Person"
    },
    actions = ["ARCHIVE"],
    resource = {
      type = "Document"
    }
  })
}

# Example 5: CIQ policy with comprehensive configuration (original example)
resource "indykite_authorization_policy" "policy_for_ciq" {
  name         = "terraform-pipeline-policy-for-ciq"
  display_name = "Terraform policy for CIQ"
  description  = "Policy for CIQ in terraform configuration"
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

# Note: The location parameter accepts an Application Space ID.
# You can use either a hardcoded GID or a reference to an application_space resource.
# The policy will automatically populate app_space_id and customer_id as computed fields.
