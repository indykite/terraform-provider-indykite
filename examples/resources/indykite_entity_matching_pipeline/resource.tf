resource "indykite_entity_matching_pipeline" "create-pipeline" {
  name               = "terraform-entitymatching-pipeline"
  display_name       = "Terraform entitymatching pipeline"
  description        = "External entitymatching pipeline for terraform"
  location           = "AppSpaceID"
  source_node_filter = ["Person"]
  target_node_filter = ["Person"]
}
