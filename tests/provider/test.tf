resource "time_static" "example" {}

locals {
  app_space_name = "terraform-pipeline-appspace-${time_static.example.unix}"
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
