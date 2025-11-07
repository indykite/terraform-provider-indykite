# Note: This file uses variables for sensitive values like passwords and access keys.
# Define these variables in your terraform.tfvars or pass them via environment variables:
# - var.kafka_password_01
# - var.kafka_password_02
# - var.azure_event_grid_access_key
# - var.azure_service_bus_connection_string
# - var.kafka_password
# - var.kafka_prod_password
# - var.azure_grid_access_key
# - var.azure_bus_connection_string
# - var.kafka_audit_password
# - var.kafka_capture_password

# Original example - comprehensive event sink with all provider types (still valid)
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
      password              = var.kafka_password_01
      provider_display_name = "provider-display-name"
    }
  }
  providers {
    provider_name = "kafka-provider-02"
    kafka {
      brokers  = ["kafka-02-01:9092", "kafka-02-02:9092"]
      topic    = "events"
      username = "other-username"
      password = var.kafka_password_02
    }
  }
  providers {
    provider_name = "azuregrid"
    azure_event_grid {
      topic_endpoint = "https://ik-test.eventgrid.azure.net/api/events"
      access_key     = var.azure_event_grid_access_key
    }
  }
  providers {
    provider_name = "azurebus"
    azure_service_bus {
      connection_string   = var.azure_service_bus_connection_string
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
        key   = "captureLabel"
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
        key   = "captureLabel"
        value = "access-granted"
      }
      event_type = "indykite.audit.capture.*"
    }
  }
  lifecycle {
    create_before_destroy = true
  }
}

# Example 1: Minimal Kafka event sink with hardcoded location
resource "indykite_event_sink" "minimal_kafka" {
  name     = "minimal-kafka-sink"
  location = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  providers {
    provider_name = "kafka-main"
    kafka {
      brokers  = ["kafka.example.com:9092"]
      topic    = "indykite-events"
      username = "kafka-user"
      password = var.kafka_password
    }
  }
  routes {
    provider_id = "kafka-main"
    keys_values_filter {
      event_type = "indykite.audit.*"
    }
  }
}

# Example 2: Kafka event sink with reference to application_space
resource "indykite_event_sink" "kafka_with_ref" {
  name         = "kafka-sink-with-ref"
  display_name = "Kafka Sink with Reference"
  description  = "Event sink using application space reference"
  location     = indykite_application_space.my_space.id
  providers {
    provider_name = "kafka-prod"
    kafka {
      brokers               = ["kafka-01.prod.example.com:9092", "kafka-02.prod.example.com:9092"]
      topic                 = "production-events"
      username              = "prod-user"
      password              = var.kafka_prod_password
      provider_display_name = "Production Kafka"
    }
  }
  routes {
    provider_id        = "kafka-prod"
    stop_processing    = true
    route_display_name = "Production Events Route"
    route_id           = "prod-route-01"
    keys_values_filter {
      event_type = "indykite.audit.*"
    }
  }
}

# Example 3: Azure Event Grid sink
resource "indykite_event_sink" "azure_grid" {
  name         = "azure-event-grid-sink"
  display_name = "Azure Event Grid Sink"
  description  = "Event sink for Azure Event Grid"
  location     = indykite_application_space.my_space.id
  providers {
    provider_name = "azure-grid-prod"
    azure_event_grid {
      topic_endpoint = "https://my-topic.eventgrid.azure.net/api/events"
      access_key     = var.azure_grid_access_key
    }
  }
  routes {
    provider_id = "azure-grid-prod"
    keys_values_filter {
      event_type = "indykite.audit.config.*"
    }
  }
}

# Example 4: Azure Service Bus sink
resource "indykite_event_sink" "azure_bus" {
  name         = "azure-service-bus-sink"
  display_name = "Azure Service Bus Sink"
  description  = "Event sink for Azure Service Bus"
  location     = indykite_application_space.my_space.id
  providers {
    provider_name = "azure-bus-prod"
    azure_service_bus {
      connection_string   = var.azure_bus_connection_string
      queue_or_topic_name = "indykite-events"
    }
  }
  routes {
    provider_id = "azure-bus-prod"
    keys_values_filter {
      event_type = "indykite.audit.capture.*"
    }
  }
}

# Example 5: Event sink with multiple routes and filters
resource "indykite_event_sink" "multi_route" {
  name         = "multi-route-sink"
  display_name = "Multi-Route Event Sink"
  description  = "Event sink with multiple routes and complex filtering"
  location     = indykite_application_space.my_space.id
  providers {
    provider_name = "kafka-audit"
    kafka {
      brokers  = ["kafka.example.com:9092"]
      topic    = "audit-events"
      username = "audit-user"
      password = var.kafka_audit_password
    }
  }
  providers {
    provider_name = "kafka-capture"
    kafka {
      brokers  = ["kafka.example.com:9092"]
      topic    = "capture-events"
      username = "capture-user"
      password = var.kafka_capture_password
    }
  }
  routes {
    provider_id        = "kafka-audit"
    stop_processing    = false
    route_display_name = "Audit Events"
    route_id           = "audit-route"
    keys_values_filter {
      event_type = "indykite.audit.config.*"
    }
  }
  routes {
    provider_id        = "kafka-capture"
    stop_processing    = true
    route_display_name = "Capture Events with Label"
    route_id           = "capture-route"
    keys_values_filter {
      key_value_pairs {
        key   = "captureLabel"
        value = "important"
      }
      event_type = "indykite.audit.capture.*"
    }
  }
}

# Note: The location parameter accepts an Application Space ID.
# You must define at least one provider (kafka, azure_event_grid, or azure_service_bus).
# You must define at least one route that references a provider_id.
# The event sink will automatically populate app_space_id and customer_id as computed fields.
# Use lifecycle.create_before_destroy = true to avoid downtime during updates.
