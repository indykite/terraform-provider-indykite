# List the event sinks of a project with the given names.
data "indykite_event_sinks" "example" {
  app_space_id = "gid:ZXhhbXBsZS1wcm9qZWN0"
  filter       = ["my-event-sink"]
}

# Last delivery error of each provider; credentials are never exposed by this data source.
output "event_sink_provider_errors" {
  value = {
    for p in flatten(data.indykite_event_sinks.example.event_sinks[*].providers) : p.provider_name => p.last_error
  }
}
