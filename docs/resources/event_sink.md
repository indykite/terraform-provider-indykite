---
# generated by https://github.com/hashicorp/terraform-plugin-docs with custom templates
page_title: "indykite_event_sink Resource - IndyKite"
subcategory: ""
description: |-
  Event Sink configuration is used to configure outbound events.

  	There can be only one configuration per AppSpace (Project).

  	Outbound events are designed to notify external systems about important changes within
  	the IndyKite Knowledge Graph (IKG).

  	These external systems may require real-time synchronization or need to react to
  	changes occurring in the platform.

  Supported filters
  | **Method** | **Event Type** | **Key** | **Value (example)** |
  | --- | --- | --- | --- |
  |  | **Ingest Events** |  |  |
  | **BatchUpsertNodes** | indykite.audit.capture.upsert.node | captureLabel | Car |
  |  |  | captureLabel | Green |
  | **BatchUpsertRelationships** | indykite.audit.capture.upsert.relationship | captureLabel | RENT |
  | **BatchDeleteNodes** | indykite.audit.capture.delete.node | captureLabel | Car |
  |  |  | captureLabel | Green |
  | **BatchDeleteRelationships** | indykite.audit.capture.delete.relationship | captureLabel | RENT |
  | **BatchDeleteNodeProperties** | indykite.audit.capture.delete.node.property |  |  |
  | **BatchDeleteRelationshipProperties** | indykite.audit.capture.delete.relationship.property |  |  |
  | **BatchDeleteNodeTags** | indykite.audit.capture.delete.node.tag | captureLabel | Car |
  |  |  | captureLabel | Green |
  |  | **Configuration Events** |  |  |
  | Config | indykite.audit.config.create |  |  |
  |  | indykite.audit.config.read |  |  |
  |  | indykite.audit.config.update |  |  |
  |  | indykite.audit.config.delete |  |  |
  |  | indykite.audit.config.permission.assign |  |  |
  |  | indykite.audit.config.permission.revoke |  |  |
  |  | **Token Events** |  |  |
  | TokenIntrospect | indykite.audit.credentials.token.introspected |  |  |
  |  | **Authorization Events** |  |  |
  | Authorization | indykite.audit.authorization.isauthorized |  |  |
  |  | indykite.audit.authorization.whatauthorized |  |  |
  |  | indykite.audit.authorization.whoauthorized |  |  |
  |  | **Ciq Events** |  |  |
  | Ciq | indykite.audit.ciq.execute |  |  |
  |  |  |  |  |
---

# indykite_event_sink (Resource)

Event Sink configuration is used to configure outbound events.

		There can be only one configuration per AppSpace (Project).

		Outbound events are designed to notify external systems about important changes within
		the IndyKite Knowledge Graph (IKG).

		These external systems may require real-time synchronization or need to react to
		changes occurring in the platform.


## Supported filters

| **Method** | **Event Type** | **Key** | **Value (example)** |
| --- | --- | --- | --- |
|  | **Ingest Events** |  |  |
| **BatchUpsertNodes** | indykite.audit.capture.upsert.node | captureLabel | Car |
|  |  | captureLabel | Green |
| **BatchUpsertRelationships** | indykite.audit.capture.upsert.relationship | captureLabel | RENT |
| **BatchDeleteNodes** | indykite.audit.capture.delete.node | captureLabel | Car |
|  |  | captureLabel | Green |
| **BatchDeleteRelationships** | indykite.audit.capture.delete.relationship | captureLabel | RENT |
| **BatchDeleteNodeProperties** | indykite.audit.capture.delete.node.property |  |  |
| **BatchDeleteRelationshipProperties** | indykite.audit.capture.delete.relationship.property |  |  |
| **BatchDeleteNodeTags** | indykite.audit.capture.delete.node.tag | captureLabel | Car |
|  |  | captureLabel | Green |
|  | **Configuration Events** |  |  |
| Config | indykite.audit.config.create |  |  |
|  | indykite.audit.config.read |  |  |
|  | indykite.audit.config.update |  |  |
|  | indykite.audit.config.delete |  |  |
|  | indykite.audit.config.permission.assign |  |  |
|  | indykite.audit.config.permission.revoke |  |  |
|  | **Token Events** |  |  |
| TokenIntrospect | indykite.audit.credentials.token.introspected |  |  |
|  | **Authorization Events** |  |  |
| Authorization | indykite.audit.authorization.isauthorized |  |  |
|  | indykite.audit.authorization.whatauthorized |  |  |
|  | indykite.audit.authorization.whoauthorized |  |  |
|  | **Ciq Events** |  |  |
| Ciq | indykite.audit.ciq.execute |  |  |
|  |  |  |  |

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `location` (String) Identifier of Location, where to create resource
- `name` (String) Unique client assigned immutable identifier. Can not be updated without creating a new resource.
- `providers` (Block List, Min: 1) (see [below for nested schema](#nestedblock--providers))
- `routes` (Block List, Min: 1) (see [below for nested schema](#nestedblock--routes))

### Optional

- `description` (String) Your own description of the resource. Must be less than or equal to 256 UTF-8 bytes.
- `display_name` (String) The display name for the instance. Can be updated without creating a new resource.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `app_space_id` (String) Identifier of Application Space
- `create_time` (String) Timestamp when the Resource was created. Assigned by the server. A timestamp in RFC3339 UTC "Zulu" format, accurate to nanoseconds. Example: "2014-10-02T15:01:23.045123456Z".
- `customer_id` (String) Identifier of Customer
- `id` (String) The ID of this resource.
- `update_time` (String) Timestamp when the Resource was last updated. Assigned by the server. A timestamp in RFC3339 UTC "Zulu" format, accurate to nanoseconds. Example: "2014-10-02T15:01:23.045123456Z".

<a id="nestedblock--providers"></a>
### Nested Schema for `providers`

Required:

- `provider_name` (String)

Optional:

- `azure_event_grid` (Block List, Max: 1) AzureEventGridSinkConfig (see [below for nested schema](#nestedblock--providers--azure_event_grid))
- `azure_service_bus` (Block List, Max: 1) AzureServiceBusSinkConfig (see [below for nested schema](#nestedblock--providers--azure_service_bus))
- `kafka` (Block List, Max: 1) KafkaSinkConfig (see [below for nested schema](#nestedblock--providers--kafka))

<a id="nestedblock--providers--azure_event_grid"></a>
### Nested Schema for `providers.azure_event_grid`

Required:

- `access_key` (String, Sensitive)
- `topic_endpoint` (String)

Optional:

- `provider_display_name` (String)


<a id="nestedblock--providers--azure_service_bus"></a>
### Nested Schema for `providers.azure_service_bus`

Required:

- `connection_string` (String, Sensitive)
- `queue_or_topic_name` (String)

Optional:

- `provider_display_name` (String)


<a id="nestedblock--providers--kafka"></a>
### Nested Schema for `providers.kafka`

Required:

- `brokers` (List of String) Brokers specify Kafka destinations to connect to.
- `password` (String, Sensitive)
- `topic` (String)
- `username` (String)

Optional:

- `disable_tls` (Boolean) Disable TLS for communication. Highly NOT RECOMMENDED.
- `provider_display_name` (String)
- `tls_skip_verify` (Boolean) Skip TLS certificate verification. NOT RECOMMENDED.



<a id="nestedblock--routes"></a>
### Nested Schema for `routes`

Required:

- `provider_id` (String)

Optional:

- `keys_values_filter` (Block List, Max: 1) (see [below for nested schema](#nestedblock--routes--keys_values_filter))
- `route_display_name` (String)
- `route_id` (String)
- `stop_processing` (Boolean)

<a id="nestedblock--routes--keys_values_filter"></a>
### Nested Schema for `routes.keys_values_filter`

Required:

- `event_type` (String)

Optional:

- `key_value_pairs` (Block List) List of key/value pairs for the ingest event types. (see [below for nested schema](#nestedblock--routes--keys_values_filter--key_value_pairs))

<a id="nestedblock--routes--keys_values_filter--key_value_pairs"></a>
### Nested Schema for `routes.keys_values_filter.key_value_pairs`

Required:

- `key` (String) Key for the ingest eventType
- `value` (String) Value for the ingest eventType




<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `default` (String)
- `delete` (String)
- `read` (String)
- `update` (String)
