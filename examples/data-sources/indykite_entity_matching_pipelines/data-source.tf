# List the entity matching pipelines of a project with the given names.
data "indykite_entity_matching_pipelines" "example" {
  app_space_id = "gid:ZXhhbXBsZS1wcm9qZWN0"
  filter       = ["my-pipeline"]
}

# Status of the last run of each pipeline.
output "entity_matching_runs" {
  value = {
    for p in data.indykite_entity_matching_pipelines.example.entity_matching_pipelines : p.name => {
      status           = p.entity_matching_status
      matched_entities = p.matched_entities
      last_run_time    = p.last_run_time
    }
  }
}
