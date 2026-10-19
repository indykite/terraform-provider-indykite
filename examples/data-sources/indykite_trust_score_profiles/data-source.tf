# List the trust score profiles of a project with the given names.
data "indykite_trust_score_profiles" "example" {
  app_space_id = "gid:ZXhhbXBsZS1wcm9qZWN0"
  filter       = ["my-trust-score"]
}

# Last calculation run of each profile. schedule holds the API value (for example DAILY);
# the indykite_trust_score_profile resource takes it prefixed with UPDATE_FREQUENCY_.
output "trust_score_runs" {
  value = {
    for p in data.indykite_trust_score_profiles.example.trust_score_profiles : p.name => {
      schedule      = p.schedule
      last_run_id   = p.last_run_id
      last_run_ends = p.last_run_end_time
    }
  }
}
