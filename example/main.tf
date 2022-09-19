// indykite provider integrates IndyKite platform with Terraform scripting.
provider "indykite" {}

// Read customer information
data "indykite_customer" "wonka" {
  name = "wonka"
}

# ##########################
# ### Application Spaces ###
# ##########################
data "indykite_application_space" "wonka-factory" {
  customer_id = data.indykite_customer.wonka.id
  name        = "factory"
}

# List multiple application spaces, non existing names are silently ignored
data "indykite_application_spaces" "more-of-them" {
  customer_id = data.indykite_customer.wonka.id
  filter      = ["factory", "oompa-loompas", "non-existing-one"]
}

# Handle 2 more resources
resource "indykite_application_space" "development" {
  customer_id = data.indykite_customer.wonka.id
  name        = "development-space-1"
  # deletion_protection=false
}
resource "indykite_application_space" "factory2" {
  customer_id         = indykite_application_space.development.customer_id
  name                = "two-${data.indykite_application_space.wonka-factory.name}"
  display_name        = "two-${data.indykite_application_space.wonka-factory.display_name}"
  deletion_protection = false
}

# ###############
# ### Tenants ###
# ###############

data "indykite_tenant" "wonka-1" {
  app_space_id = data.indykite_application_space.wonka-factory.id
  name         = "wonka-1"
}

data "indykite_tenants" "all-tenants" {
  # This is how it should be done properly, no count or other iteration
  # See https://blog.gruntwork.io/terraform-tips-tricks-loops-if-statements-and-gotchas-f739bbae55f9
  for_each = {
    for appSpace in data.indykite_application_spaces.more-of-them.app_spaces : appSpace.id => appSpace
  }
  app_space_id = each.key
  filter       = ["wonka-1", "wonka-2", "cocoa-beans-1", "non-existing-one"]
}

output "tenants" {
  value       = flatten(values(data.indykite_tenants.all-tenants)[*].tenants)
  description = "Contains all tenants under multiple app spaces combined into single array"
}

resource "indykite_tenant" "development" {
  # This is how it should be done properly, no count or other iteration
  # See https://blog.gruntwork.io/terraform-tips-tricks-loops-if-statements-and-gotchas-f739bbae55f9
  for_each = {
    for appSpace in data.indykite_application_spaces.more-of-them.app_spaces : appSpace.id => appSpace
  }
  issuer_id    = each.value.issuer_id
  name         = "tenant-for-${each.value.name}"
  display_name = "Hola! New Tenant: ${each.value.name}"
  # deletion_protection=false
}

# ####################
# ### Applications ###
# ####################

data "indykite_application" "wonka-bars" {
  app_space_id = data.indykite_application_space.wonka-factory.id
  name         = "wonka-bars"
}

data "indykite_applications" "all-wonkas" {
  # This is how it should be done properly, no count or other iteration
  # See https://blog.gruntwork.io/terraform-tips-tricks-loops-if-statements-and-gotchas-f739bbae55f9
  for_each = {
    for appSpace in data.indykite_application_spaces.more-of-them.app_spaces : appSpace.id => appSpace
  }
  app_space_id = each.key
  filter       = ["wonka-bars", "non-existing-one", "loompaland"]
}

output "apps" {
  value       = flatten(values(data.indykite_applications.all-wonkas)[*].applications)
  description = "Contains all applications under multiple app spaces combined into single array"
}

resource "indykite_application" "development" {
  # This is how it should be done properly, no count or other iteration
  # See https://blog.gruntwork.io/terraform-tips-tricks-loops-if-statements-and-gotchas-f739bbae55f9
  for_each = {
    for appSpace in data.indykite_application_spaces.more-of-them.app_spaces : appSpace.id => appSpace
  }
  app_space_id = each.key
  name         = "app-for-${each.value.name}"
  # deletion_protection=false
}

# ##########################
# ### Application Agents ###
# ##########################

data "indykite_application_agent" "opa" {
  app_space_id = data.indykite_application_space.wonka-factory.id
  name         = "wonka-app-agent"
}

data "indykite_application_agents" "all-wonkas" {
  # This is how it should be done properly, no count or other iteration
  # See https://blog.gruntwork.io/terraform-tips-tricks-loops-if-statements-and-gotchas-f739bbae55f9
  for_each = {
    for appSpace in data.indykite_application_spaces.more-of-them.app_spaces : appSpace.id => appSpace
  }
  app_space_id = each.key
  filter       = ["wonka-app-agent", "non-existing-one", "loompaland-app-agent"]
}

output "app_agents" {
  value       = flatten(values(data.indykite_application_agents.all-wonkas)[*].app_agents)
  description = "Contains all application agents under multiple app spaces combined into single array"
}

resource "indykite_application_agent" "development" {
  application_id = data.indykite_application.wonka-bars.id
  name           = "app-agent-for-${data.indykite_application_space.wonka-factory.name}"
  # deletion_protection=false
}

# ######################################
# ### Application Agents Credentials ###
# ######################################

resource "indykite_application_agent_credential" "development" {
  app_agent_id = data.indykite_application_agent.opa.id
}

resource "indykite_application_agent_credential" "with-public" {
  app_agent_id      = data.indykite_application_agent.opa.id
  display_name      = "Key with custom private-public key pair"
  expire_time       = "2040-12-31T12:34:56-01:00"
  default_tenant_id = data.indykite_tenant.wonka-1.id
  public_key_jwk    = <<-EOT
  {
    "kty": "EC",
    "use": "sig",
    "crv": "P-256",
    "kid": "sig-1631087814",
    "x": "WZQ9LMC08l-L05oxRa-9ObmaPQlTuWHX2GaAmMgAuSE",
    "y": "V5VuhYvyQ2ACiVznB_aqzfVWmwfhvPFD6Dc4X32WUE8",
    "alg": "ES256"
  }
  EOT
}

//// Creates a new OAuth2 authenticator
//resource "indykite_oauth2_client" "facebook" {
//  tenant = indykite_tenant
//
//  name = "oauth2-facebook"
//  description = "Facebook Authenticator"
//
//  client {
//    provider = "facebook.com"
//    client_id = "test"
//    client_secret = "secret"
//    redirect_uri = [
//      "https://indykite-dev.indykite.id/handler/oauth2/callback"]
//  }
//
//  depends_on = [
//    indykite_service_authentication
//  ]
//}
//
//


resource "indykite_auth_flow" "flow" {
  # This is how it should be done properly, no count or other iteration
  # See https://blog.gruntwork.io/terraform-tips-tricks-loops-if-statements-and-gotchas-f739bbae55f9
  for_each = {
    for appSpace in data.indykite_application_spaces.more-of-them.app_spaces : appSpace.id => appSpace
  }
  location = each.key

  name        = "default-flow-for-${each.value.name}"
  description = "Sample Authentication Flow for _${each.value.display_name}_"

  yaml = <<-EOT
    ---
    configuration:
        activities:
            '00000000':
                '@type': input
            000000F0:
                '@type': success
        sequences:
            - sourceRef: '00000000'
              targetRef: 000000F0
    EOT
}

resource "indykite_email_notification" "wonka" {
  # This is how it should be done properly, no count or other iteration
  # See https://blog.gruntwork.io/terraform-tips-tricks-loops-if-statements-and-gotchas-f739bbae55f9
  for_each = {
    for tenant in flatten(values(data.indykite_tenants.all-tenants)[*].tenants) : tenant.id => tenant
  }
  location = each.key

  name         = "wonka-email-service-for-${each.value.name}"
  display_name = "Email Service TADAAAAA: ${each.value.display_name}"
  # description = "A brief description of the display name"

  default_from_address {
    address = "wonka@chocolate-factory.com"
  }

  amazon_ses {
    access_key_id          = "amazon_access_key_id_for_wonka"
    secret_access_key      = "SuperDuperSecretKeyFromAmazonForWonka"
    region                 = "eu-west-1"
    configuration_set_name = "config-name"
    default_from_address {
      address = "wonka-ses@chocolate-factory.com"
    }
    feedback_forwarding_email_address = "oompa-boss@chocolate-factory.com"
    reply_to_addresses = [
      "secretary@chocolate-factory.com",
      "stock@chocolate-factory.com"
    ]
  }

  # sendgrid {
  #   api_key = "sendgrid_access_key_id_for_wonka"
  #   sandbox_mode = true
  #   ip_pool_name = "oompa_pool"
  #   host = "https://wonka.sengrid.com"
  # }

  invitation_message {
    from {
      address = "wonka@chocolate-factory.com"
    }
    reply_to {
      address = "oompa@chocolate-factory.com"
    }

    to {
      address = "customer@example.com"
      name    = "John Doe"
    }
    to {
      address = "customer-2@example.com"
    }

    cc {
      address = "customer-3@example.com"
      name    = "John Doe"
    }
    cc {
      address = "another@example.com"
    }
    bcc {
      address = "hidden@example.com"
      name    = "Secret Doe"
    }
    bcc {
      address = "customer-hidden@example.com"
    }
    subject = "Subject of the message"

    template {
      id      = "MTID-2"
      version = "v3"

      headers = {
        SomeHeader               = "a"
        "X-Mailgun-Variables"    = "{\"user-id\": \"Mailgun accept JSON in headers as variables\"}"
        "X Something With Space" = "Just testing spaces in key"
      }

      custom_arguments = {
        a = "a"
        b = "132"
      }
      template_dynamic_values = "{\"value\": 123}"
      categories = [
        "a",
        "b",
        "c"
      ]
      event_payload = "abc_def"
      ses_arn       = "SES_ARN_number"
    }
  }
}

# ########################
# ### OAuth2 Providers ###
# ########################
resource "indykite_oauth2_provider" "wonka-bars-oauth2-provider" {
  app_space_id = data.indykite_application_space.wonka-factory.id
  name         = "oauth2-provider-for-${data.indykite_application_space.wonka-factory.name}"

  grant_types                     = [local.oauth2_grant_types.authorization_code]
  response_types                  = [local.oauth2_response_types.code]
  scopes                          = ["openid", "profile", "email", "phone"]
  token_endpoint_auth_method      = [local.oauth2_token_endpoint_auth_methods.client_secret_basic]
  token_endpoint_auth_signing_alg = [local.supported_auth_signing_algs.ES256]
  request_uris                    = ["https://request_uri"]
  request_object_signing_alg      = local.supported_auth_signing_algs.ES256
  front_channel_login_uri         = { "front_channel_1" = "https://front_channel_login_uri.com" }
  front_channel_consent_uri       = { "front_channel_1" = "https://front_channel_consent_uri.com" }
  #     deletion_protection=false
}

data "indykite_oauth2_provider" "wonka-bars-oauth2-provider2" {
  oauth2_provider_id = indykite_oauth2_provider.wonka-bars-oauth2-provider.id
}

# ########################
# ### OAuth2 Applications ###
# ########################
resource "indykite_oauth2_application" "wonka-bars-oauth2-application" {
  oauth2_provider_id = indykite_oauth2_provider.wonka-bars-oauth2-provider.id
  name               = "oauth2-application-for-${data.indykite_application_space.wonka-factory.name}"

  oauth2_application_display_name = "oauth2_application_display_name"
  oauth2_application_description  = "oauth2_application_description"
  redirect_uris                   = ["https://redirect_uris"]
  owner                           = "owner"
  policy_uri                      = "https://policy_uri"
  allowed_cors_origins            = ["https://allowed_cors_origins"]
  terms_of_service_uri            = "https://terms_of_service_uri"
  client_uri                      = "https://client_uri"
  logo_uri                        = "https://logo_uri"
  user_support_email_address      = "user_support_email@address.com"
  additional_contacts             = ["additional_contacts"]
  subject_type                    = local.oauth2_client_subject_types.public
  sector_identifier_uri           = "https://sector_identifier_uri"
  grant_types                     = [local.oauth2_grant_types.authorization_code]
  response_types                  = [local.oauth2_response_types.code]
  scopes                          = ["openid", "profile", "email", "phone"]
  audiences                       = ["00000000-0000-0000-0000-000000000000"]
  token_endpoint_auth_method      = local.oauth2_token_endpoint_auth_methods.client_secret_basic
  token_endpoint_auth_signing_alg = local.supported_auth_signing_algs.ES256
  userinfo_signed_response_alg    = local.supported_auth_signing_algs.RS256
  deletion_protection             = false
}

data "indykite_oauth2_application" "wonka-bars-oauth2-application-2" {
  oauth2_application_id = indykite_oauth2_application.wonka-bars-oauth2-application.id
}
