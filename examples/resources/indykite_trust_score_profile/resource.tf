resource "indykite_trust_score_profile" "trust-score" {
  name                = "terraform-resolver-get"
  display_name        = "Terraform trust score profile"
  description         = "Trust score profile for terraform pipeline"
  location            = "AppSpaceID"
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
}
