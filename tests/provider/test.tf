resource "time_static" "example" {}

variable "LOCATION_ID" {
  type        = string
  description = "AppSpace for entitymatching"
}

locals {
  app_space_name = "terraform-pipeline-appspace-${time_static.example.unix}"
  location_id    = var.LOCATION_ID
}

data "indykite_customer" "customer" {
  name = "terraform-pipeline"
}

resource "indykite_application_space" "appspace" {
  customer_id    = data.indykite_customer.customer.id
  name           = local.app_space_name
  display_name   = "Terraform appspace ${time_static.example.unix}"
  description    = "Application space for terraform pipeline"
  region         = "us-east1"
  ikg_size       = "4GB"
  replica_region = "us-west1"
  lifecycle {
    create_before_destroy = true
  }
  deletion_protection = false
}

# resource "indykite_application_space" "appspaceb" {
#   customer_id    = data.indykite_customer.customer.id
#   name           = "terraform-pipeline-appspaceb-${time_static.example.unix}"
#   display_name   = "Terraform appspace ${time_static.example.unix}"
#   description    = "Application space for terraform pipeline"
#   region         = "us-east1"
#   ikg_size       = "4GB"
#   replica_region = "us-west1"
#   db_connection {
#     url      = "neo4j+s://xxxxxxxx.databases.neo4j.io"
#     username = "testuser"
#     password = "testpass"
#     name     = "testdb"
#   }
#   lifecycle {
#     create_before_destroy = true
#   }
#   deletion_protection = false
# }

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
  application_id  = indykite_application.application.id
  name            = "terraform-pipeline-agent-${time_static.example.unix}"
  display_name    = "Terraform agent ${time_static.example.unix}"
  description     = "Agent for terraform pipeline"
  api_permissions = ["Authorization", "Capture"]
  lifecycle {
    create_before_destroy = true
  }
  deletion_protection = false
}

resource "indykite_application_agent_credential" "with_public" {
  app_agent_id = indykite_application_agent.agent.id
  display_name = "Terraform credential ${time_static.example.unix}"
  expire_time  = "2026-12-31T12:34:56-01:00"
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

resource "indykite_authorization_policy" "policy_for_ciq" {
  name         = "terraform-pipeline-policy-for-ciq-${time_static.example.unix}"
  display_name = "Terraform policy for CIQ ${time_static.example.unix}"
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
  location = indykite_application_space.appspace.id
  status   = "active"
  lifecycle {
    create_before_destroy = true
  }
}


resource "indykite_knowledge_query" "create-query" {
  name         = "terraform-knowledge-query-${time_static.example.unix}"
  display_name = "Terraform knowledge-query  ${time_static.example.unix}"
  description  = "Knowledge query for terraform"
  location     = indykite_application_space.appspace.id
  query = jsonencode({
    "nodes" : ["ln.property.value"],
    "relationships" : [],
    "filter" : { "attribute" : "ln.property.value", "operator" : "=", "value" : "$lnValue" }
  })

  status    = "active"
  policy_id = indykite_authorization_policy.policy_for_ciq.id
  lifecycle {
    create_before_destroy = true
  }
}


resource "indykite_event_sink" "create-event" {
  name         = "terraform-event-sink-${time_static.example.unix}"
  display_name = "Terraform event sink  ${time_static.example.unix}"
  description  = "Event sink for terraform"
  location     = indykite_application_space.appspace.id
  providers {
    provider_name = "kafka-provider-01"
    kafka {
      brokers               = ["kafka-01:9092", "kafka-02:9092"]
      topic                 = "events"
      username              = "my-username"
      password              = "some-super-secret-password" # checkov:skip=CKV_SECRET_6:acceptance test
      provider_display_name = "provider-display-name"
    }
  }
  providers {
    provider_name = "kafka-provider-02"
    kafka {
      brokers  = ["kafka-02-01:9092", "kafka-02-02:9092"]
      topic    = "events"
      username = "other-username"
      password = "some-other-secret-password" # checkov:skip=CKV_SECRET_6:acceptance test
    }
  }
  providers {
    provider_name = "azuregrid"
    azure_event_grid {
      topic_endpoint = "https://ik-test.eventgrid.azure.net/api/events"
      access_key     = "secret-access-key"
    }
  }
  providers {
    provider_name = "azurebus"
    azure_service_bus {
      connection_string   = "personal-connection-info"
      queue_or_topic_name = "your-queue"
    }
  }
  routes {
    provider_id     = "kafka-provider-01"
    stop_processing = false
    keys_values_filter {
      event_type = "indykite.audit.config.create"
    }
    route_display_name = "route-display-name"
    route_id           = "route-id"
  }
  routes {
    provider_id     = "kafka-provider-02"
    stop_processing = false
    keys_values_filter {
      key_value_pairs {
        key   = "relationshipcreated"
        value = "access-granted"
      }
      event_type = "indykite.audit.capture.*"
    }
  }
  routes {
    provider_id     = "azuregrid"
    stop_processing = false
    keys_values_filter {
      event_type = "indykite.audit.config.create"
    }
  }
  routes {
    provider_id     = "azurebus"
    stop_processing = false
    keys_values_filter {
      key_value_pairs {
        key   = "relationshipcreated"
        value = "access-granted"
      }
      event_type = "indykite.audit.capture.*"
    }
  }
  lifecycle {
    create_before_destroy = true
  }
}

resource "indykite_trust_score_profile" "create-score" {
  name                = "terraform-trust-score-profile-${time_static.example.unix}"
  display_name        = "Terraform trust score profile  ${time_static.example.unix}"
  description         = "Trust score profile for terraform"
  location            = local.location_id
  node_classification = "Person"
  dimension {
    name   = "NAME_VERIFICATION"
    weight = 0.5
  }
  dimension {
    name   = "NAME_ORIGIN"
    weight = 0.5
  }
  schedule = "UPDATE_FREQUENCY_DAILY"
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

# ====================================================
# Additional tests for different configuration options
# ====================================================

# -----------------------------------------------------------------------------
# Test: Minimal configuration (only required fields)
# -----------------------------------------------------------------------------

# Authorization policy with simple config for ciq
resource "indykite_authorization_policy" "policy_minimal" {
  name = "terraform-pipeline-policy-minimal-${time_static.example.unix}"
  json = jsonencode({
    "meta" : {
      "policy_version" : "1.0-ciq"
    },
    "subject" : {
      "type" : "Person"
    },
    "condition" : {
      "cypher" : "MATCH (company:Company)-[:OFFERS]->(contract:Contract)<-[:ACCEPTED]-(subject:Person)-[:HAS]->(payment:PaymentMethod), (contract)-[:COVERS]->(vehicle:Vehicle)-[:HAS]->(ln:LicenseNumber)",
      "filter" : [
        {
          "operator" : "AND",
          "operands" : [
            {
              "attribute" : "subject.external_id",
              "operator" : "=",
              "value" : "$subject_external_id"
            },
            {
              "attribute" : "$token.sub",
              "operator" : "=",
              "value" : "$token_sub"
            }
          ]
        }
      ]
    },
    "allowed_reads" : {
      "nodes" : [
        "company.*",
        "subject.*",
        "payment.*"
      ]
    },
    "allowed_upserts" : {
      "relationships" : {
        "relationship_types" : [
          {
            "type" : "GRANTED",
            "source_node_label" : "Company",
            "target_node_label" : "PaymentMethod"
          }
        ]
      }
    }
  })
  location = indykite_application_space.appspace.id
  status   = "active"
  lifecycle {
    create_before_destroy = true
  }
}

# External data resolver with minimal config (GET without headers)
resource "indykite_external_data_resolver" "resolver_minimal" {
  name     = "terraform-resolver-minimal-${time_static.example.unix}"
  location = indykite_application_space.appspace.id

  url               = "https://api.example.com/data"
  method            = "GET"
  request_type      = "json"
  response_type     = "json"
  response_selector = "."
  lifecycle {
    create_before_destroy = true
  }
}

# Knowledge query with minimal config
resource "indykite_knowledge_query" "query_minimal" {
  name     = "terraform-knowledge-query-minimal-${time_static.example.unix}"
  location = indykite_application_space.appspace.id
  query = jsonencode({
    "nodes" : [
      "company.external_id",
      "subject.external_id",
      "payment.external_id"
    ],
    "upsert_relationships" : [
      {
        "name" : "newRel",
        "source" : "company",
        "target" : "payment",
        "type" : "GRANTED"
      }
    ]
  })
  status    = "active"
  policy_id = indykite_authorization_policy.policy_minimal.id
  lifecycle {
    create_before_destroy = true
  }
}

# Trust score profile with minimal dimensions
resource "indykite_trust_score_profile" "score_minimal" {
  name                = "terraform-trust-score-minimal-${time_static.example.unix}"
  location            = local.location_id
  node_classification = "Organization"
  dimension {
    name   = "NAME_FRESHNESS"
    weight = 1.0
  }
  schedule = "UPDATE_FREQUENCY_SIX_HOURS"
  lifecycle {
    create_before_destroy = true
  }
}

# Entity matching pipeline with minimal config
resource "indykite_entity_matching_pipeline" "pipeline_minimal" {
  name     = "terraform-entitymatching-minimal-${time_static.example.unix}"
  location = local.location_id

  source_node_filter = ["Organization"]
  target_node_filter = ["Organization"]
  lifecycle {
    create_before_destroy = true
  }
}

# -----------------------------------------------------------------------------
# Test: Different trust score dimensions and schedules
# -----------------------------------------------------------------------------

# Trust score profile with all dimensions
resource "indykite_trust_score_profile" "score_all_dimensions" {
  name                = "terraform-trust-score-all-dims-${time_static.example.unix}"
  display_name        = "All dimensions profile"
  description         = "Profile testing all available dimensions"
  location            = local.location_id
  node_classification = "Asset"
  dimension {
    name   = "NAME_ORIGIN"
    weight = 0.2
  }
  dimension {
    name   = "NAME_VALIDITY"
    weight = 0.2
  }
  dimension {
    name   = "NAME_COMPLETENESS"
    weight = 0.2
  }
  dimension {
    name   = "NAME_FRESHNESS"
    weight = 0.2
  }
  dimension {
    name   = "NAME_VERIFICATION"
    weight = 0.2
  }
  schedule = "UPDATE_FREQUENCY_SIX_HOURS"
  lifecycle {
    create_before_destroy = true
  }
}

# -----------------------------------------------------------------------------
# Test: External data resolver with different methods and configurations
# -----------------------------------------------------------------------------

# External data resolver with multiple headers
resource "indykite_external_data_resolver" "resolver_multi_headers" {
  name         = "terraform-resolver-multi-headers-${time_static.example.unix}"
  display_name = "Resolver with multiple headers"
  description  = "Testing multiple header configurations"
  location     = indykite_application_space.appspace.id

  url    = "https://api.example.com/enrichment"
  method = "GET"
  headers {
    name   = "Authorization"
    values = ["Bearer token123"]
  }
  headers {
    name   = "X-API-Key"
    values = ["api-key-value"]
  }
  headers {
    name   = "Accept"
    values = ["application/json", "application/xml"]
  }
  request_type      = "json"
  response_type     = "json"
  response_selector = ".data"
  lifecycle {
    create_before_destroy = true
  }
}

# External data resolver with POST and payload
resource "indykite_external_data_resolver" "resolver_post_payload" {
  name     = "terraform-resolver-post-payload-${time_static.example.unix}"
  location = indykite_application_space.appspace.id

  url               = "https://api.example.com/query"
  method            = "POST"
  request_type      = "json"
  request_payload   = jsonencode({ query = "SELECT * FROM users WHERE id = $1", params = ["$userId"] })
  response_type     = "json"
  response_selector = ".results[0]"
  lifecycle {
    create_before_destroy = true
  }
}

# -----------------------------------------------------------------------------
# Test: Authorization policies with different policy versions
# -----------------------------------------------------------------------------

# Authorization policy with draft status
resource "indykite_authorization_policy" "policy_draft" {
  name         = "terraform-pipeline-policy-draft-${time_static.example.unix}"
  display_name = "Draft policy for testing"
  json = jsonencode({
    meta = {
      policyVersion = "1.0-indykite"
    },
    subject = {
      type = "User"
    },
    actions = ["CAN_EDIT"],
    resource = {
      type = "File"
    },
    condition = {
      cypher = "MATCH (subject:User)-[:EDITOR_OF]->(resource:File)"
    }
  })
  location = indykite_application_space.appspace.id
  status   = "draft"
  lifecycle {
    create_before_destroy = true
  }
}

# -----------------------------------------------------------------------------
# Test: Data source lookups for existing resources
# -----------------------------------------------------------------------------

# Look up the application space we created
data "indykite_application_space" "lookup_appspace" {
  app_space_id = indykite_application_space.appspace.id
  depends_on   = [indykite_application_space.appspace]
}

# Look up the application we created
data "indykite_application" "lookup_application" {
  application_id = indykite_application.application.id
  depends_on     = [indykite_application.application]
}

# Look up the application agent we created
data "indykite_application_agent" "lookup_agent" {
  app_agent_id    = indykite_application_agent.agent.id
  api_permissions = ["Authorization", "Capture"]
  depends_on      = [indykite_application_agent.agent]
}

# -----------------------------------------------------------------------------
# Test: Entity matching pipeline with multiple node filters
# -----------------------------------------------------------------------------

# Entity matching pipeline with multiple source and target filters
resource "indykite_entity_matching_pipeline" "pipeline_multi_filters" {
  name         = "terraform-entitymatching-multi-${time_static.example.unix}"
  display_name = "Multi-filter pipeline"
  description  = "Pipeline with multiple source and target filters"
  location     = local.location_id

  source_node_filter = ["Person", "Organization", "Device"]
  target_node_filter = ["Person", "Organization"]
  lifecycle {
    create_before_destroy = true
  }
}
