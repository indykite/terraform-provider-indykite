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
      password              = "some-super-secret-password"
      provider_display_name = "provider-display-name"
    }
  }
  providers {
    provider_name = "kafka-provider-02"
    kafka {
      brokers  = ["kafka-02-01:9092", "kafka-02-02:9092"]
      topic    = "events"
      username = "other-username"
      password = "some-other-secret-password"
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
      event_type = "indykite.audit.ingest.*"
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
      event_type = "indykite.audit.ingest.*"
    }
  }
  lifecycle {
    create_before_destroy = true
  }
}
