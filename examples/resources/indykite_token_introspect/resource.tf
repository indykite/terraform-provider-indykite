# Example - JWT with online validation
resource "indykite_token_introspect" "token1" {
  name         = "terraform-token-introspect"
  display_name = "Terraform token introspect"
  description  = "Token introspect for terraform pipeline"
  location     = "AppSpaceID"
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
  location     = "AppSpaceID"
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

# Example 1: Minimal JWT with online validation and hardcoded location
resource "indykite_token_introspect" "minimal_online" {
  name     = "minimal-online-jwt"
  location = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
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
  location     = indykite_application_space.my_space.id
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
  location     = indykite_application_space.my_space.id
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
        "n" : "--public-key-modulus-here--",
        "kty" : "RSA"
      })
    ]
  }
  ikg_node_type = "User"
  sub_claim     = "user_id"
}

# Example 4: JWT with custom claims mapping
resource "indykite_token_introspect" "custom_claims" {
  name         = "jwt-custom-claims"
  display_name = "JWT with Custom Claims"
  description  = "Token introspect with extensive claims mapping"
  location     = indykite_application_space.my_space.id
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

# Note: The location parameter accepts an Application Space ID.
# You must use either jwt_matcher or opaque_matcher (not both).
# You must use either online_validation or offline_validation (not both).
# The token introspect will automatically populate app_space_id and customer_id as computed fields.
