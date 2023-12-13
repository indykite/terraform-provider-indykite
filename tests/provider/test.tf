resource "time_static" "example" {}

locals {
  app_space_name = "terraform-pipeline-appspace-${time_static.example.unix}"
  tenant_name = "tenant-terraform-pipeline-${time_static.example.unix}"
}

data indykite_customer customer {
	name = "terraform-pipeline-customer"
}

resource indykite_application_space appspace {
	customer_id  = data.indykite_customer.customer.id
	name         = local.app_space_name
	display_name = "Terraform appspace ${time_static.example.unix}"
	description  = "Application space for terraform pipeline"
	lifecycle {
    	create_before_destroy = true
  	}
	deletion_protection = false
}

resource indykite_tenant tenant {
	name         = local.tenant_name
	display_name = "Terraform tenant ${time_static.example.unix}"
	description  = "Tenant for terraform pipeline"
	issuer_id    = indykite_application_space.appspace.issuer_id
	lifecycle {
    	create_before_destroy = true
  	}
	deletion_protection = false
}

resource indykite_application application {
	app_space_id = indykite_application_space.appspace.id
	name         = "terraform-pipeline-application-${time_static.example.unix}"
	display_name = "Terraform application ${time_static.example.unix}"
	description  = "Application for terraform pipeline"
	lifecycle {
    	create_before_destroy = true
  	}
	deletion_protection = false
}

resource indykite_application_agent agent {
	application_id = indykite_application.application.id
	name           = "terraform-pipeline-agent-${time_static.example.unix}"
	display_name   = "Terraform agent ${time_static.example.unix}"
	description    = "Agent for terraform pipeline"
	lifecycle {
    	create_before_destroy = true
  	}
	deletion_protection = false
}

resource "indykite_application_agent_credential" "with_public" {
	app_agent_id      = indykite_application_agent.agent.id
	display_name      = "Terraform credential ${time_static.example.unix}"
	expire_time       = "2040-12-31T12:34:56-01:00"
	lifecycle {
    	create_before_destroy = true
  	}
}

resource indykite_authorization_policy policy_drive_car {
	name         = "terraform-pipeline-policy-drive-car-${time_static.example.unix}"
	display_name = "Terraform policy drive car ${time_static.example.unix}"
	description  = "Policy for terraform pipeline"
	json         = jsonencode({
		meta = {
			policyVersion = "1.0-indykite"
		},
		subject = {
			type = "Person"
		},
		actions  = ["CAN_DRIVE"],
		resource = {
			type = "Car"
		},
		condition = {
			cypher = "MATCH (subject:Person)-[:OWNS]->(resource:Car)"
		}
	})
	location = indykite_application_space.appspace.id
	status   = "active"
	lifecycle {
    	create_before_destroy = true
  	}
}

resource indykite_email_notification email_conf {
	location     = indykite_application_space.appspace.id
	name         = "terraform-pipeline-email-conf-${time_static.example.unix}"
	display_name = "Terraform email conf ${time_static.example.unix}"
	description  = "Email conf for terraform pipeline"
	sendgrid {
    	api_key = "264289b5-175e-6c56-c458-049a25d1cf51"
        sandbox_mode = true
        ip_pool_name = "100.25.24.23.22"
        host = "https://api.sendgrid.com"
    }
	invitation_message {
		to {
			address = "default@example.com"
			name = "Terraform to"
		}
		subject = "Subject"
		template {
			id = "286959f5-132e-6c74-c951-049c36d1bf12"
		}
	}
	lifecycle {
    	create_before_destroy = true
  	}
}

resource indykite_oauth2_provider oauth2_provider {
	app_space_id              = indykite_application_space.appspace.id
	name                      = "terraform-pipeline-oauth2-provider-${time_static.example.unix}"
	display_name              = "Terraform oauth2 provider ${time_static.example.unix}"
	description               = "Oauth2 provider for terraform pipeline"
	front_channel_consent_uri = {
		"en" = "https://example.com/consent"
	}
	front_channel_login_uri = {
		"en" = "https://example.com/login"
	}
	grant_types                     = ["authorization_code", "refresh_token"]
	response_types                  = ["code", "token"]
	scopes                          = ["openid", "profile", "email", "offline_access"]
	token_endpoint_auth_method      = ["client_secret_basic"]
	token_endpoint_auth_signing_alg = ["RS256"]
	request_object_signing_alg      = "RS256"
	request_uris                    = []
	lifecycle {
    	create_before_destroy = true
  	}
	deletion_protection = false
}

resource indykite_oauth2_application oauth2_app {
	name                            = "terraform-pipeline-oauth2-app-${time_static.example.unix}"
	display_name                    = "Terraform oauth2 app ${time_static.example.unix}"
	description                     = "Oauth2 app for terraform pipeline"
	client_uri                      = "https://example.com"
	logo_uri                        = "https://example.com/logo.png"
	oauth2_application_display_name = "Terraform oauth2 app"
	oauth2_provider_id              = indykite_oauth2_provider.oauth2_provider.id
	owner                           = "Terraform oauth2 app owner"
	policy_uri                      = "https://example.com/policy"
	redirect_uris                   = ["https://example.com/callback"]
	scopes                          = ["openid", "profile", "email", "offline_access"]
	subject_type                    = "public"
	terms_of_service_uri            = "https://example.com/terms"
	user_support_email_address      = "contact@example.com"
	additional_contacts             = ["support@example.com"]
	allowed_cors_origins            = ["https://example.com"]
	grant_types                     = ["authorization_code", "refresh_token"]
	oauth2_application_description  = "Terraform oauth2 app description"
	response_types                  = ["code", "token"]
	sector_identifier_uri           = "https://example.com/sector"
	userinfo_signed_response_alg    = "RS256"
	token_endpoint_auth_method      = "client_secret_basic"
	token_endpoint_auth_signing_alg = "RS256"
	lifecycle {
    	create_before_destroy = true
  	}
	deletion_protection = false
}

resource indykite_oauth2_client oauth2_client {
	name          = "terraform-pipeline-oauth2-client-${time_static.example.unix}"
	display_name  = "Terraform oauth2 client ${time_static.example.unix}"
	description   = "Oauth2 client for terraform pipeline"
	location      = indykite_application_space.appspace.id
	auth_style    = "auto_detect"
	client_id     = "terraform-pipeline-oauth2-client"
	provider_type = "google.com"

	allow_signup           = true
	allowed_scopes         = ["openid", "profile", "email", "offline_access"]
	authorization_endpoint = "https://example.com/authorize"
	client_secret          = "secretsecret"
	default_scopes         = ["openid", "profile", "email", "offline_access"]
	discovery_url          = "https://example.com/.well-known/openid-configuration"
	hosted_domain          = "example.com"
	image_url              = "https://example.com/logo.png"
	issuer                 = "https://example.com"
	jwks_uri               = "https://example.com/jwks"
	redirect_uri           = ["https://example.com/callback"]
	team_id                = "team-id"
	tenant                 = "tenant"
	token_endpoint         = "https://example.com/token"
	userinfo_endpoint      = "https://example.com/userinfo"
	lifecycle {
    	create_before_destroy = true
  	}
}

resource "indykite_customer_configuration" "customer_config" {
	customer_id = data.indykite_customer.customer.id
	lifecycle {
    	create_before_destroy = true
  	}
}

resource "indykite_application_space_configuration" "appspace_config" {
	app_space_id = indykite_application_space.appspace.id

	username_policy {
		allowed_username_formats  = ["email"]
		valid_email               = true
		verify_email              = true
		verify_email_grace_period = "600s"
		allowed_email_domains     = ["example.com", "outlook.com"]
		exclusive_email_domains   = ["indykite.com"]
	}
	unique_property_constraints = {
		"property1" : jsonencode({ "tenantUnique" : false })
		"super_property" : jsonencode({ "tenantUnique" : true, "canonicalization" : ["unicode"] })
	}
	lifecycle {
    	create_before_destroy = true
  	}
}

resource "indykite_tenant_configuration" "tenant_config" {
	tenant_id = indykite_tenant.tenant.id

	username_policy {
		allowed_username_formats  = ["email"]
		valid_email               = true
		verify_email              = true
		verify_email_grace_period = "600s"
		allowed_email_domains     = ["example.com", "outlook.com"]
		exclusive_email_domains   = ["indykite.com"]
	}

}
