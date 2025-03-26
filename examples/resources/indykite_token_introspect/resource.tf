resource "indykite_token_introspect" "token1" {
  name         = "terraform-token-introspect"
  display_name = "Terraform token introspect"
  description  = "Token introspect for terraform pipeline"
  location     = "AppSpaceID"
  jwt_matcher {
        issuer = "https://example.com"
        audience = "audience-id"
    }
    online_validation {
        cache_ttl = 600
    }
    claims_mapping = {
        "email" = "mail",
        "name" = "full_name"
    }
    ikg_node_type = "MyUser"
    perform_upsert = true
}

resource "indykite_token_introspect" "token2" {
  name         = "terraform-token-introspect"
  display_name = "Terraform token introspect"
  description  = "Token introspect for terraform pipeline"
  location     = "AppSpaceID"
  jwt_matcher {
        issuer = "https://example.com"
        audience = "audience-id"
  }
  offline_validation {
    public_jwks = [
        jsonencode({
            "kid": "abc",
            "use": "sig",
            "alg": "RS256",
            "n": "--nothing-real-just-random-xyqwerasf--",
            "kty": "RSA"
        }),
        jsonencode({
            "kid": "jkl",
            "use": "sig",
            "alg": "RS256",
            "n": "--nothing-real-just-random-435asdf43--",
            "kty": "RSA"
        })
    ]
  }
  ikg_node_type = "MyUser"
  sub_claim = "custom_sub"
}
