# List the knowledge queries of a project with the given names.
data "indykite_knowledge_queries" "example" {
  app_space_id = "gid:ZXhhbXBsZS1wcm9qZWN0"
  filter       = ["my-knowledge-query"]
}

output "knowledge_query_ids" {
  value = [for q in data.indykite_knowledge_queries.example.knowledge_queries : q.id]
}
