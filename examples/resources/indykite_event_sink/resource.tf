resource "indykite_event_sink" "create-event" {
  name         = "terraform-event-sink"
  display_name = "Terraform event sink"
  description  = "Event sink for terraform"
  location     = indykite_application_space.appspace.id
  providers {
     provider_name = "kafka-provider-01"
    kafka {
      brokers = ["kafka-01:9092", "kafka-02:9092"]
      topic = "events"
      username = "my-username"
      password = "some-super-secret-password"
      }
  }
  providers {
    provider_name = "kafka-provider-02"
    kafka {
      brokers = ["kafka-02-01:9092", "kafka-02-02:9092"]
      topic = "events"
      username = "other-username"
      password = "some-other-secret-password"
      }
  }
  routes {
    provider_id = "kafka-provider-01"
	  stop_processing = false
		event_type_filter = "indykite.eventsink.config.create"
	}
  routes {
    provider_id = "kafka-provider-02"
	  stop_processing = false
		context_key_value_filter {
      key = "relationshipcreated"
      value = "access-granted"
    } 
	}
}