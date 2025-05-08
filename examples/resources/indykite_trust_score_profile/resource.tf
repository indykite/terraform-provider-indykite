resource "indykite_trust_score_profile" "trust-score" {
  name                = "terraform-trust-score"
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

resource "indykite_trust_score_profile" "trust-score2" {
  name                = "terraform-trust-score2"
  display_name        = "Terraform trust score profile2"
  description         = "Trust score profile for terraform pipeline"
  location            = "AppSpaceID"
  node_classification = "Resource"
  dimension {
    name   = "NAME_COMPLETENESS"
    weight = 0.4
  }
  dimension {
    name   = "NAME_VALIDITY"
    weight = 0.6
  }
  schedule = "UPDATE_FREQUENCY_SIX_HOURS"
}
