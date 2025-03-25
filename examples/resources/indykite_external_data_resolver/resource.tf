resource "indykite_external_data_resolver" "get-resolver" {
  name         = "terraform-resolver-get"
  display_name = "Terraform external data resolver get"
  description  = "External data resolver for terraform pipeline"
  location     = "AppSpaceID"

  url    = "https://www.example.com/sourceresolver?data=xxx"
  method = "GET"
  headers {
    name   = "Authorization"
    values = ["Bearer edolkUTY"]
  }
  request_type      = "json"
  response_type     = "json"
  response_selector = ".resp"
}

resource "indykite_external_data_resolver" "post-resolver" {
  name         = "terraform-resolver-post"
  display_name = "Terraform external data resolver post"
  description  = "External data resolver for terraform pipeline"
  location     = "AppSpaceID"

  url    = "https://example.com/sourceresolver2/where-data"
  method = "POST"
  headers {
    name   = "Authorization"
    values = ["Bearer edbkLbPnb6VfcRPTkUTY"]
  }
  headers {
    name   = "Content-Type"
    values = ["application/json"]
  }
  request_type      = "json"
  request_payload   = "{\"data\": \"$resp\"}"
  response_type     = "json"
  response_selector = ".resp"
}

