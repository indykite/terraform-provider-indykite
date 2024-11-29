resource "time_static" "example" {}

variable "LOCATION_ID" {
    type = string
    description = "AppSpace for entitymatching"
}

locals {
  app_space_name = "terraform-pipeline-appspace-${time_static.example.unix}"
  location_id =  var.LOCATION_ID
}

data "indykite_customer" "customer" {
  name = "terraform-pipeline"
}

resource "indykite_application_space" "appspace" {
  customer_id  = data.indykite_customer.customer.id
  name         = local.app_space_name
  display_name = "Terraform appspace ${time_static.example.unix}"
  description  = "Application space for terraform pipeline"
  region       = "europe-west1"
  lifecycle {
    create_before_destroy = true
  }
  deletion_protection = false
}

resource "indykite_application" "application" {
  app_space_id = indykite_application_space.appspace.id
  name         = "terraform-pipeline-application-${time_static.example.unix}"
  display_name = "Terraform application ${time_static.example.unix}"
  description  = "Application for terraform pipeline"
  lifecycle {
    create_before_destroy = true
  }
  deletion_protection = false
}

resource "indykite_application_agent" "agent" {
  application_id = indykite_application.application.id
  name           = "terraform-pipeline-agent-${time_static.example.unix}"
  display_name   = "Terraform agent ${time_static.example.unix}"
  description    = "Agent for terraform pipeline"
  lifecycle {
    create_before_destroy = true
  }
  deletion_protection = false
}

resource "indykite_application_agent_credential" "with_public" {
  app_agent_id = indykite_application_agent.agent.id
  display_name = "Terraform credential ${time_static.example.unix}"
  expire_time  = "2040-12-31T12:34:56-01:00"
  lifecycle {
    create_before_destroy = true
  }
}

resource "indykite_authorization_policy" "policy_drive_car" {
  name         = "terraform-pipeline-policy-drive-car-${time_static.example.unix}"
  display_name = "Terraform policy drive car ${time_static.example.unix}"
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
  location = indykite_application_space.appspace.id
  status   = "active"
  lifecycle {
    create_before_destroy = true
  }
}

resource "indykite_consent" "basic-user-data" {
  location    = indykite_application_space.appspace.id
  name        = "location-name-sharing"
  description = "This consent will allow third parties to access the location and name of the user"

  purpose          = "To send you your order you need to share your location and name with the delivery service"
  application_id   = indykite_application.application.id
  validity_period  = 96400
  revoke_after_use = false
  data_points = [
    "{\"returns\": [{\"properties\": [\"name\", \"location\"]}]}"
  ]
}

resource "indykite_consent" "advance-user-data" {
  location    = indykite_application_space.appspace.id
  name        = "advance-sharing"
  description = "Allow servicing company to access car model and manufacturer name"

  purpose          = "Share you car model and manufacturer name with the car service"
  application_id   = indykite_application.application.id
  validity_period  = 96400
  revoke_after_use = false
  data_points = [jsonencode(
    {
      "query" : "->[:BELONGS]-(c:CAR)-[:MADEBY]->(o:MANUFACTURER)",
      "returns" : [
        {
          "variable" : "c",
          "properties" : [
            "Model"
          ]
        },
        {
          "variable" : "o",
          "properties" : [
            "Name"
          ]
        }
      ]
    }
  )]
}

resource "indykite_external_data_resolver" "get-resolver" {
  name         = "terraform-resolver-get-${time_static.example.unix}"
  display_name = "Terraform external data resolver get ${time_static.example.unix}"
  description  = "External data resolver for terraform pipeline"
  location     = indykite_application_space.appspace.id

  url    = "https://www.example.com/sourceresolver"
  method = "GET"
  headers {
    name   = "Authorization"
    values = ["Bearer edolkUTY"]
  }
  request_type      = "json"
  response_type     = "json"
  response_selector = "."
  lifecycle {
    create_before_destroy = true
  }
}

resource "indykite_external_data_resolver" "post-resolver" {
  name         = "terraform-resolver-post-${time_static.example.unix}"
  display_name = "Terraform external data resolver post ${time_static.example.unix}"
  description  = "External data resolver for terraform pipeline"
  location     = indykite_application_space.appspace.id

  url    = "https://example.com/sourceresolver2"
  method = "POST"
  headers {
    name   = "Authorization"
    values = ["Bearer edokLoPnb6VfcRRTkUTY"]
  }
  headers {
    name   = "Content-Type"
    values = ["application/json"]
  }
  request_type      = "json"
  request_payload   = "{\"key\": \"value\"}"
  response_type     = "json"
  response_selector = "."
  lifecycle {
    create_before_destroy = true
  }
}

resource "indykite_entity_matching_pipeline" "create-pipeline" {
  name         = "terraform-entitymatching-pipeline-${time_static.example.unix}"
  display_name = "Terraform entitymatching pipeline  ${time_static.example.unix}"
  description  = "External entitymatching pipeline for terraform"
  location     = local.location_id

  source_node_filter = ["Person"]
	target_node_filter = ["Person"]
  lifecycle {
    create_before_destroy = true
  }
}