# Prerequisites: resources referenced by the examples below.
resource "indykite_application_space" "my_space" {
  customer_id = "gid:ZXhhbXBsZS1vcmdhbml6YXRpb24"
  name        = "my-project"
  region      = "us-east1"
}

# Example - JWT with online validation
resource "indykite_token_introspect" "token1" {
  name         = "terraform-token-introspect"
  display_name = "Terraform token introspect"
  description  = "Token introspect for terraform pipeline"
  project_id   = "gid:ZXhhbXBsZS1wcm9qZWN0"
  jwt_matcher {
    issuer   = "https://example.com"
    audience = "audience-id"
  }
  online_validation {
    cache_ttl = 600
  }
  claims_mapping = {
    "email" = "mail",
    "name"  = "full_name"
  }
  ikg_node_type  = "MyUser"
  perform_upsert = true
}

# Example - JWT with offline validation
resource "indykite_token_introspect" "token2" {
  name         = "terraform-token-introspect-offline"
  display_name = "Terraform token introspect offline"
  description  = "Token introspect for terraform pipeline with offline validation"
  project_id   = "gid:ZXhhbXBsZS1wcm9qZWN0"
  jwt_matcher {
    issuer   = "https://example.com"
    audience = "audience-id"
  }
  offline_validation {
    public_jwks = [
      jsonencode({
        "kid" : "abc",
        "use" : "sig",
        "alg" : "RS256",
        "n" : "--nothing-real-just-random-xyqwerasf--",
        "kty" : "RSA"
      }),
      jsonencode({
        "kid" : "jkl",
        "use" : "sig",
        "alg" : "RS256",
        "n" : "--nothing-real-just-random-435asdf43--",
        "kty" : "RSA"
      })
    ]
  }
  ikg_node_type = "MyUser"
  sub_claim     = "custom_sub"
}

# Example 1: Minimal JWT with online validation and hardcoded project_id
resource "indykite_token_introspect" "minimal_online" {
  name       = "minimal-online-jwt"
  project_id = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
  jwt_matcher {
    issuer   = "https://auth.example.com"
    audience = "my-app"
  }
  online_validation {
    cache_ttl = 300
  }
  ikg_node_type = "User"
}

# Example 2: JWT with online validation and reference to application_space
resource "indykite_token_introspect" "online_with_ref" {
  name         = "online-jwt-with-ref"
  display_name = "Online JWT with Reference"
  description  = "JWT token introspect with online validation using app space reference"
  project_id   = indykite_application_space.my_space.id
  jwt_matcher {
    issuer   = "https://auth.example.com"
    audience = "my-application"
  }
  online_validation {
    cache_ttl = 600
  }
  claims_mapping = {
    "email"       = "user_email",
    "name"        = "user_name",
    "given_name"  = "first_name",
    "family_name" = "last_name"
  }
  ikg_node_type  = "Person"
  perform_upsert = true
}

# Example 3: JWT with offline validation and reference
resource "indykite_token_introspect" "offline_with_ref" {
  name         = "offline-jwt-with-ref"
  display_name = "Offline JWT with Reference"
  description  = "JWT token introspect with offline validation"
  project_id   = indykite_application_space.my_space.id
  jwt_matcher {
    issuer   = "https://auth.example.com"
    audience = "my-app-offline"
  }
  offline_validation {
    public_jwks = [
      jsonencode({
        "kid" : "key-1",
        "use" : "sig",
        "alg" : "RS256",
        "n" : "sXchB8A4CkaR3DjDdFUXAYVHx_dWx6gT6tEFc6aOaZ5CPAsqJ8lmwczE6xT74j_xoz1PZtKh6q6tRxX4Mk-GB4KFu0YQH1-WN6uEhEcE0fMtsymrZ1OlN3GhxUAt93Q6zpg3ZD9P4ZtXWr7g6OpkqKAv12v2YKvHtJmEznV0OZPZQj5POj1U0EJp0B98VpTmRtT1K2fKjddAlDq3t35u5xPdkL9l9yLeaMLqGw3tNhxG8amj_Mlq3zEy_QwOmR6OKO5mIF0kCshRsfKj8cMyOrUdxzqaZGuv9KYkgA3ulFjHbInCQwFZ8Z5h7bnQZEkSz94Cz3C4X8hGgSyOShLMuWw",
        "kty" : "RSA"
      })
    ]
  }
  ikg_node_type = "User"
  sub_claim     = "user_id"
}

# Example 4: JWT with offline validation and RFC 8414 discovery
# Without public_jwks, the keys are fetched from the jwks_uri of the issuer's
# .well-known/oauth-authorization-server document instead of .well-known/openid-configuration.
resource "indykite_token_introspect" "offline_rfc8414" {
  name       = "offline-jwt-rfc8414"
  project_id = indykite_application_space.my_space.id
  jwt_matcher {
    issuer   = "https://auth.example.com"
    audience = "my-app-rfc8414"
  }
  offline_validation {
    discovery = "rfc8414"
  }
  ikg_node_type = "Person"
}

# Example 5: Opaque token with OAuth 2.0 token introspection (RFC 7662)
variable "introspection_client_secret" {
  type      = string
  sensitive = true
}

resource "indykite_token_introspect" "opaque_oauth2" {
  name         = "opaque-oauth2-introspection"
  display_name = "Opaque OAuth 2.0 introspection"
  project_id   = indykite_application_space.my_space.id
  opaque_matcher {
    hint = "my.domain.com"
  }
  online_oauth2_validation {
    introspection_endpoint = "https://auth.example.com/oauth2/introspect"
    secret_basic {
      client_id     = "my-client"
      client_secret = var.introspection_client_secret
    }
    cache_ttl = 300
  }
  ikg_node_type = "Person"
}

# Example 6: JWT with custom claims mapping
resource "indykite_token_introspect" "custom_claims" {
  name         = "jwt-custom-claims"
  display_name = "JWT with Custom Claims"
  description  = "Token introspect with extensive claims mapping"
  project_id   = indykite_application_space.my_space.id
  jwt_matcher {
    issuer   = "https://auth.example.com"
    audience = "custom-app"
  }
  online_validation {
    cache_ttl = 900
  }
  claims_mapping = {
    "email"        = "email_address",
    "name"         = "full_name",
    "given_name"   = "first_name",
    "family_name"  = "last_name",
    "phone_number" = "phone",
    "address"      = "postal_address"
  }
  ikg_node_type  = "Person"
  perform_upsert = true
  sub_claim      = "user_identifier"
}

# Note: The project_id parameter accepts an Application Space ID. location is deprecated, use project_id instead.
# You must use either jwt_matcher or opaque_matcher (not both).
# You must use exactly one of offline_validation, online_validation or online_oauth2_validation.
# offline_validation works only with jwt_matcher; opaque_matcher with online_validation requires user_info_endpoint.
# client_secret in online_oauth2_validation is write-only: the API never returns it, it is kept from the Terraform state.
# The token introspect will automatically populate app_space_id and customer_id as computed fields.
