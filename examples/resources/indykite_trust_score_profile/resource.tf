# Example - trust score for Person
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

# Example - trust score for Resource
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

# Example 1: Minimal trust score with hardcoded location
resource "indykite_trust_score_profile" "minimal_trust_score" {
  name                = "minimal-trust-score"
  location            = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  node_classification = "User"
  dimension {
    name   = "NAME_VERIFICATION"
    weight = 1.0
  }
  schedule = "UPDATE_FREQUENCY_DAILY"
}

# Example 2: Trust score with reference to application_space
resource "indykite_trust_score_profile" "trust_score_with_ref" {
  name                = "trust-score-with-ref"
  display_name        = "Trust Score with Reference"
  description         = "Trust score profile using application space reference"
  location            = indykite_application_space.my_space.id
  node_classification = "Person"
  dimension {
    name   = "NAME_VERIFICATION"
    weight = 0.6
  }
  dimension {
    name   = "NAME_ORIGIN"
    weight = 0.4
  }
  schedule = "UPDATE_FREQUENCY_DAILY"
}

# Example 3: Trust score with multiple dimensions
resource "indykite_trust_score_profile" "multi_dimension_trust_score" {
  name                = "multi-dimension-trust-score"
  display_name        = "Multi-Dimension Trust Score"
  description         = "Trust score with multiple weighted dimensions"
  location            = indykite_application_space.my_space.id
  node_classification = "Person"
  dimension {
    name   = "NAME_VERIFICATION"
    weight = 0.3
  }
  dimension {
    name   = "NAME_ORIGIN"
    weight = 0.2
  }
  dimension {
    name   = "NAME_COMPLETENESS"
    weight = 0.25
  }
  dimension {
    name   = "NAME_VALIDITY"
    weight = 0.25
  }
  schedule = "UPDATE_FREQUENCY_DAILY"
}

# Example 4: Trust score with hourly updates
resource "indykite_trust_score_profile" "hourly_trust_score" {
  name                = "hourly-trust-score"
  display_name        = "Hourly Trust Score"
  description         = "Trust score updated every hour"
  location            = indykite_application_space.my_space.id
  node_classification = "Asset"
  dimension {
    name   = "NAME_VERIFICATION"
    weight = 0.5
  }
  dimension {
    name   = "NAME_VALIDITY"
    weight = 0.5
  }
  schedule = "UPDATE_FREQUENCY_HOURLY"
}

# Example 5: Trust score with six-hour updates
resource "indykite_trust_score_profile" "six_hour_trust_score" {
  name                = "six-hour-trust-score"
  display_name        = "Six-Hour Trust Score"
  description         = "Trust score updated every six hours"
  location            = indykite_application_space.my_space.id
  node_classification = "Organization"
  dimension {
    name   = "NAME_COMPLETENESS"
    weight = 0.7
  }
  dimension {
    name   = "NAME_VALIDITY"
    weight = 0.3
  }
  schedule = "UPDATE_FREQUENCY_SIX_HOURS"
}

# Example 6: Trust score for Document classification
resource "indykite_trust_score_profile" "document_trust_score" {
  name                = "document-trust-score"
  display_name        = "Document Trust Score"
  description         = "Trust score for document entities"
  location            = indykite_application_space.my_space.id
  node_classification = "Document"
  dimension {
    name   = "NAME_VERIFICATION"
    weight = 0.4
  }
  dimension {
    name   = "NAME_ORIGIN"
    weight = 0.3
  }
  dimension {
    name   = "NAME_COMPLETENESS"
    weight = 0.3
  }
  schedule = "UPDATE_FREQUENCY_DAILY"
}

# Note: The location parameter accepts an Application Space ID.
# node_classification specifies the type of nodes this profile applies to.
# dimension weights must sum to 1.0 across all dimensions.
# schedule options: UPDATE_FREQUENCY_HOURLY, UPDATE_FREQUENCY_SIX_HOURS, UPDATE_FREQUENCY_DAILY
# The profile will automatically populate app_space_id and customer_id as computed fields.
