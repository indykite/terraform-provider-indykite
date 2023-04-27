# indykite provider integrates IndyKite platform with Terraform scripting.
# Provider for now does not support any parameters and all is set within service account credential file.
terraform {
  required_providers {
    indykite = {
      source  = "indykite/indykite"
      version = "0.1.2"
    }
  }
}


provider "indykite" {
}

resource "indykite_authorization_policy" "policy" {
  location    = "gid:AAAAAvqK6QGpxUg9lyCGC7GxtMs"
  name        = "indykite-car-authorization-policy-config"
  description = "Allowed person to drive a car."
  status      = "active"
  tags        = ["household", "driver"]

  json = <<-EOT
    {
      "meta":{
        "policyVersion":"1.0-indykite"
      },
      "subject":{
        "type":"Person"
      },
      "actions":[
        "CAN_DRIVE",
        "CAN_PERFORM_SERVICE"
      ],
      "resource":{
        "type":"Car"
      },
      "condition":{
        "cypher":"MATCH (subject:Person)-[:PART_OF]->(:Household)-[:DISPOSES]->(resource:Car)"
      }
    }
  EOT
}
