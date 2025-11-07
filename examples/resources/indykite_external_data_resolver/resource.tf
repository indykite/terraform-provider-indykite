# Example - GET resolver (still valid)
resource "indykite_external_data_resolver" "get-resolver" {
  name         = "terraform-resolver-get"
  display_name = "Terraform external data resolver get"
  description  = "External data resolver for terraform pipeline"
  location     = "AppSpaceID"

  url    = "https://www.example.com/sourceresolver?data=xxx"
  method = "GET"
  headers {
    name   = "Authorization"
    values = ["Bearer <your-token-here>"]
  }
  request_type      = "json"
  response_type     = "json"
  response_selector = ".resp"
}

# Example - POST resolver (still valid)
resource "indykite_external_data_resolver" "post-resolver" {
  name         = "terraform-resolver-post"
  display_name = "Terraform external data resolver post"
  description  = "External data resolver for terraform pipeline"
  location     = "AppSpaceID"

  url    = "https://example.com/sourceresolver2/where-data"
  method = "POST"
  headers {
    name   = "Authorization"
    values = ["Bearer <your-token-here>"]
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

# Example 1: Minimal GET resolver with hardcoded location
resource "indykite_external_data_resolver" "minimal_get" {
  name              = "minimal-get-resolver"
  location          = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  url               = "https://api.example.com/data"
  method            = "GET"
  request_type      = "json"
  response_type     = "json"
  response_selector = ".data"
}

# Example 2: GET resolver with reference to application_space
resource "indykite_external_data_resolver" "get_with_ref" {
  name              = "get-resolver-with-ref"
  display_name      = "GET Resolver with Reference"
  description       = "External data resolver using application space reference"
  location          = indykite_application_space.my_space.id
  url               = "https://api.example.com/users"
  method            = "GET"
  request_type      = "json"
  response_type     = "json"
  response_selector = ".users"
}

# Example 3: GET resolver with multiple headers
resource "indykite_external_data_resolver" "get_multi_headers" {
  name         = "get-resolver-multi-headers"
  display_name = "GET Resolver with Multiple Headers"
  description  = "Resolver with multiple authentication and custom headers"
  location     = indykite_application_space.my_space.id
  url          = "https://api.example.com/secure-data"
  method       = "GET"
  headers {
    name   = "Authorization"
    values = ["Bearer <your-token-here>"]
  }
  headers {
    name   = "X-API-Key"
    values = ["<your-api-key-here>"]
  }
  headers {
    name   = "X-Custom-Header"
    values = ["custom-value-1", "custom-value-2"]
  }
  request_type      = "json"
  response_type     = "json"
  response_selector = ".result"
}

# Example 4: POST resolver with reference
resource "indykite_external_data_resolver" "post_with_ref" {
  name         = "post-resolver-with-ref"
  display_name = "POST Resolver with Reference"
  description  = "POST resolver using application space reference"
  location     = indykite_application_space.my_space.id
  url          = "https://api.example.com/query"
  method       = "POST"
  headers {
    name   = "Authorization"
    values = ["Bearer <your-token-here>"]
  }
  headers {
    name   = "Content-Type"
    values = ["application/json"]
  }
  request_type      = "json"
  request_payload   = "{\"query\": \"$query\", \"limit\": 100}"
  response_type     = "json"
  response_selector = ".results"
}

# Example 5: POST resolver with complex payload
resource "indykite_external_data_resolver" "post_complex" {
  name         = "post-resolver-complex"
  display_name = "POST Resolver with Complex Payload"
  description  = "Resolver with complex request payload"
  location     = indykite_application_space.my_space.id
  url          = "https://api.example.com/advanced-query"
  method       = "POST"
  headers {
    name   = "Authorization"
    values = ["Bearer <your-token-here>"]
  }
  headers {
    name   = "Content-Type"
    values = ["application/json"]
  }
  request_type      = "json"
  request_payload   = "{\"filters\": {\"type\": \"$type\", \"status\": \"active\"}, \"sort\": \"created_at\", \"order\": \"desc\"}"
  response_type     = "json"
  response_selector = ".data.items"
}

# Note: The location parameter accepts an Application Space ID.
# method can be either "GET" or "POST".
# request_type and response_type currently only support "json".
# The resolver will automatically populate app_space_id and customer_id as computed fields.
