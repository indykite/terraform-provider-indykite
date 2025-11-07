# Example - basic entity matching pipeline (still valid)
resource "indykite_entity_matching_pipeline" "create-pipeline" {
  name               = "terraform-entitymatching-pipeline"
  display_name       = "Terraform entitymatching pipeline"
  description        = "External entitymatching pipeline for terraform"
  location           = "AppSpaceID"
  source_node_filter = ["Person"]
  target_node_filter = ["Person"]
}

# Example 1: Minimal configuration with hardcoded location
resource "indykite_entity_matching_pipeline" "minimal_pipeline" {
  name               = "minimal-pipeline"
  location           = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  source_node_filter = ["User"]
  target_node_filter = ["User"]
}

# Example 2: Pipeline with reference to application_space
resource "indykite_entity_matching_pipeline" "pipeline_with_ref" {
  name               = "pipeline-with-reference"
  display_name       = "Pipeline with AppSpace Reference"
  description        = "Entity matching pipeline using application space reference"
  location           = indykite_application_space.my_space.id
  source_node_filter = ["Person"]
  target_node_filter = ["Person"]
}

# Example 3: Pipeline matching different node types
resource "indykite_entity_matching_pipeline" "cross_type_pipeline" {
  name               = "cross-type-pipeline"
  display_name       = "Cross-Type Matching Pipeline"
  description        = "Pipeline matching Person to Organization"
  location           = indykite_application_space.my_space.id
  source_node_filter = ["Person"]
  target_node_filter = ["Organization"]
}

# Example 4: Pipeline with multiple source and target types
resource "indykite_entity_matching_pipeline" "multi_type_pipeline" {
  name               = "multi-type-pipeline"
  display_name       = "Multi-Type Matching Pipeline"
  description        = "Pipeline matching multiple entity types"
  location           = indykite_application_space.my_space.id
  source_node_filter = ["Person", "User", "Employee"]
  target_node_filter = ["Person", "User", "Employee"]
}

# Example 5: Pipeline for resource matching
resource "indykite_entity_matching_pipeline" "resource_pipeline" {
  name               = "resource-matching-pipeline"
  display_name       = "Resource Matching Pipeline"
  description        = "Pipeline for matching resource entities"
  location           = indykite_application_space.my_space.id
  source_node_filter = ["Asset", "Resource"]
  target_node_filter = ["Asset", "Resource"]
}

# Note: The location parameter accepts an Application Space ID.
# source_node_filter and target_node_filter are required and cannot be changed after creation (ForceNew).
# The pipeline will automatically populate app_space_id and customer_id as computed fields.
