# Example - basic entity matching pipeline (still valid)
resource "indykite_entity_matching_pipeline" "create-pipeline" {
  name               = "terraform-entitymatching-pipeline"
  display_name       = "Terraform entitymatching pipeline"
  description        = "External entitymatching pipeline for terraform"
  project_id         = "AppSpaceID"
  source_node_filter = ["Person"]
  target_node_filter = ["Person"]
}

# Example 1: Minimal configuration with hardcoded project_id
resource "indykite_entity_matching_pipeline" "minimal_pipeline" {
  name               = "minimal-pipeline"
  project_id         = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  source_node_filter = ["User"]
  target_node_filter = ["User"]
}

# Example 2: Pipeline with reference to application_space
resource "indykite_entity_matching_pipeline" "pipeline_with_ref" {
  name               = "pipeline-with-reference"
  display_name       = "Pipeline with AppSpace Reference"
  description        = "Entity matching pipeline using application space reference"
  project_id         = indykite_application_space.my_space.id
  source_node_filter = ["Person"]
  target_node_filter = ["Person"]
}

# Example 3: Pipeline matching different node types
resource "indykite_entity_matching_pipeline" "cross_type_pipeline" {
  name               = "cross-type-pipeline"
  display_name       = "Cross-Type Matching Pipeline"
  description        = "Pipeline matching Person to Organization"
  project_id         = indykite_application_space.my_space.id
  source_node_filter = ["Person"]
  target_node_filter = ["Organization"]
}

# Example 4: Pipeline with multiple source and target types
resource "indykite_entity_matching_pipeline" "multi_type_pipeline" {
  name               = "multi-type-pipeline"
  display_name       = "Multi-Type Matching Pipeline"
  description        = "Pipeline matching multiple entity types"
  project_id         = indykite_application_space.my_space.id
  source_node_filter = ["Person", "User", "Employee"]
  target_node_filter = ["Person", "User", "Employee"]
}

# Example 5: Pipeline for resource matching
resource "indykite_entity_matching_pipeline" "resource_pipeline" {
  name               = "resource-matching-pipeline"
  display_name       = "Resource Matching Pipeline"
  description        = "Pipeline for matching resource entities"
  project_id         = indykite_application_space.my_space.id
  source_node_filter = ["Asset", "Resource"]
  target_node_filter = ["Asset", "Resource"]
}

# Note: The project_id parameter accepts an Application Space ID. location is deprecated, use project_id instead.
# source_node_filter and target_node_filter are required and cannot be changed after creation (ForceNew).
# The pipeline will automatically populate app_space_id and customer_id as computed fields.
