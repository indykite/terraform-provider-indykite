// Copyright (c) 2022 IndyKite
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package indykite_test

import (
	"context"
	"fmt"
	"regexp"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/jarvis-sdk-go/config"
	configpb "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/jarvis-sdk-go/test/config/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource OAuth2 Application", func() {
	const resourceName = "indykite_oauth2_application.wonka-bars-oauth2-application"
	var (
		mockCtrl                *gomock.Controller
		mockConfigClient        *configm.MockConfigManagementAPIClient
		indykiteProviderFactory func() (*schema.Provider, error)
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)

		indykiteProviderFactory = func() (*schema.Provider, error) {
			p := indykite.Provider()
			cfgFunc := p.ConfigureContextFunc
			p.ConfigureContextFunc =
				func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
					client, _ := config.NewTestClient(ctx, mockConfigClient)
					ctx = context.WithValue(ctx, indykite.ClientContext, client)
					return cfgFunc(ctx, data)
				}
			return p, nil
		}
	})

	It("Test all CRUD", func() {
		oauth2ApplicationID := "gid:AAAAD2luZHlraURlgAAEDwAABCC"
		oauth2ProviderID := "gid:AAAAD2luZHlraURlgAAEDwABCDE"
		clientSecret := "supersecretsecret"
		tfConfigDef :=
			`resource "indykite_oauth2_application" "wonka-bars-oauth2-application" {
				oauth2_provider_id = "` + oauth2ProviderID + `"
				name = "acme"
				display_name = "%s"
				description = "%s"
				%s
			}`

		initialResp := &configpb.OAuth2Application{
			CustomerId:       customerID,
			AppSpaceId:       appSpaceID,
			Oauth2ProviderId: oauth2ProviderID,
			Id:               oauth2ApplicationID,
			Name:             "acme",
			DisplayName:      "acme",
			Description:      wrapperspb.String("Just some OAuth2Application description"),
			CreateTime:       timestamppb.Now(),
			UpdateTime:       timestamppb.Now(),
			Config: &configpb.OAuth2ApplicationConfig{
				ClientId:                "00000000-0000-0000-0000-000000000000",
				DisplayName:             "config_display_name",
				Description:             "config_description",
				RedirectUris:            []string{"https://redirectUri1"},
				Owner:                   "owner",
				PolicyUri:               "https://PolicyUri",
				AllowedCorsOrigins:      []string{"https://allowedCorsOrigin1"},
				TermsOfServiceUri:       "https://TermsOfServiceUri",
				ClientUri:               "https://ClientUri",
				LogoUri:                 "https://LogoUri",
				UserSupportEmailAddress: "UserSupport@EmailAddress.com",
				AdditionalContacts:      []string{"AdditionalContact1"},
				SubjectType:             configpb.ClientSubjectType_CLIENT_SUBJECT_TYPE_PUBLIC,
				SectorIdentifierUri:     "https://SectorIdentifierUri",
				GrantTypes: []configpb.GrantType{
					configpb.GrantType_GRANT_TYPE_AUTHORIZATION_CODE,
				},
				ResponseTypes: []configpb.ResponseType{
					configpb.ResponseType_RESPONSE_TYPE_TOKEN,
				},
				Scopes:    []string{"openid", "profile", "email", "phone"},
				Audiences: []string{"00000000-0000-0000-0000-000000000000"},
				TokenEndpointAuthMethod: configpb.
					TokenEndpointAuthMethod_TOKEN_ENDPOINT_AUTH_METHOD_CLIENT_SECRET_BASIC,
				TokenEndpointAuthSigningAlg: "ES256",
				UserinfoSignedResponseAlg:   "RS256",
			},
		}

		readAfter1stUpdateResp := &configpb.OAuth2Application{
			CustomerId:       initialResp.CustomerId,
			AppSpaceId:       initialResp.AppSpaceId,
			Oauth2ProviderId: initialResp.Oauth2ProviderId,
			Id:               initialResp.Id,
			Name:             "acme",
			DisplayName:      "acme",
			Description:      wrapperspb.String("Another OAuth2Application description"),
			CreateTime:       initialResp.CreateTime,
			UpdateTime:       timestamppb.Now(),
			Config: &configpb.OAuth2ApplicationConfig{
				ClientId:                "00000000-0000-0000-0000-000000000000",
				DisplayName:             "config_display_name",
				Description:             "config_description",
				RedirectUris:            []string{"https://redirectUri1"},
				Owner:                   "owner",
				PolicyUri:               "https://PolicyUri/updated",
				AllowedCorsOrigins:      []string{"https://allowedCorsOrigin1"},
				TermsOfServiceUri:       "https://TermsOfServiceUri",
				ClientUri:               "https://ClientUri",
				LogoUri:                 "https://LogoUri",
				UserSupportEmailAddress: "UserSupport@EmailAddress.com",
				AdditionalContacts:      []string{"AdditionalContact1"},
				SubjectType:             configpb.ClientSubjectType_CLIENT_SUBJECT_TYPE_PUBLIC,
				SectorIdentifierUri:     "https://SectorIdentifierUri",
				GrantTypes: []configpb.GrantType{
					configpb.GrantType_GRANT_TYPE_AUTHORIZATION_CODE,
				},
				ResponseTypes: []configpb.ResponseType{
					configpb.ResponseType_RESPONSE_TYPE_TOKEN,
				},
				Scopes:    []string{"openid"},
				Audiences: []string{"00000000-0000-0000-0000-000000000000"},
				TokenEndpointAuthMethod: configpb.
					TokenEndpointAuthMethod_TOKEN_ENDPOINT_AUTH_METHOD_CLIENT_SECRET_BASIC,
				TokenEndpointAuthSigningAlg: "ES256",
				UserinfoSignedResponseAlg:   "RS256",
			},
		}
		readAfter2ndUpdateResp := &configpb.OAuth2Application{
			CustomerId:       initialResp.CustomerId,
			AppSpaceId:       initialResp.AppSpaceId,
			Oauth2ProviderId: initialResp.Oauth2ProviderId,
			Id:               initialResp.Id,
			Name:             "acme",
			DisplayName:      "Some new display name",
			Description:      nil,
			CreateTime:       initialResp.CreateTime,
			UpdateTime:       timestamppb.Now(),
			Config: &configpb.OAuth2ApplicationConfig{
				ClientId:                "00000000-0000-0000-0000-000000000000",
				DisplayName:             "config_display_name_updated",
				Description:             "config_description",
				RedirectUris:            []string{"https://redirectUri1"},
				Owner:                   "owner",
				PolicyUri:               "https://PolicyUri/updated",
				AllowedCorsOrigins:      []string{"https://allowedCorsOrigin1"},
				TermsOfServiceUri:       "https://TermsOfServiceUri",
				ClientUri:               "https://ClientUri",
				LogoUri:                 "https://LogoUri",
				UserSupportEmailAddress: "UserSupport@EmailAddress.com",
				AdditionalContacts:      []string{"AdditionalContact1", "AdditionalContact2"},
				SubjectType:             configpb.ClientSubjectType_CLIENT_SUBJECT_TYPE_PUBLIC,
				SectorIdentifierUri:     "https://SectorIdentifierUri",
				GrantTypes: []configpb.GrantType{
					configpb.GrantType_GRANT_TYPE_AUTHORIZATION_CODE,
				},
				ResponseTypes: []configpb.ResponseType{
					configpb.ResponseType_RESPONSE_TYPE_TOKEN,
				},
				Scopes:    []string{"openid"},
				Audiences: []string{"00000000-0000-0000-0000-000000000000"},
				TokenEndpointAuthMethod: configpb.
					TokenEndpointAuthMethod_TOKEN_ENDPOINT_AUTH_METHOD_CLIENT_SECRET_BASIC,
				TokenEndpointAuthSigningAlg: "ES256",
				UserinfoSignedResponseAlg:   "RS256",
			},
		}

		// MOCKS
		// There are 5 test steps
		// 1. step call: Create + Read
		// 2. step call: Read, Update, Read
		// 3. step call: Read, Update, Read
		// 4. step call: Read + delete (going to fail)
		// 5. step call: Read (changes only in deletion_protection do not trigger API)
		// after steps Delete is called

		// Create
		mockConfigClient.EXPECT().
			CreateOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Oauth2ProviderId": Equal(initialResp.Oauth2ProviderId),
				"Name":             Equal(initialResp.Name),
				"DisplayName":      BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(initialResp.Description.Value),
				})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"ClientId":                    Equal(""),
					"DisplayName":                 Equal(initialResp.Config.DisplayName),
					"Description":                 Equal(initialResp.Config.Description),
					"RedirectUris":                ConsistOf(initialResp.Config.RedirectUris),
					"Owner":                       Equal(initialResp.Config.Owner),
					"PolicyUri":                   Equal(initialResp.Config.PolicyUri),
					"AllowedCorsOrigins":          ConsistOf(initialResp.Config.AllowedCorsOrigins),
					"TermsOfServiceUri":           Equal(initialResp.Config.TermsOfServiceUri),
					"ClientUri":                   Equal(initialResp.Config.ClientUri),
					"LogoUri":                     Equal(initialResp.Config.LogoUri),
					"UserSupportEmailAddress":     Equal(initialResp.Config.UserSupportEmailAddress),
					"AdditionalContacts":          ConsistOf(initialResp.Config.AdditionalContacts),
					"SubjectType":                 Equal(initialResp.Config.SubjectType),
					"SectorIdentifierUri":         Equal(initialResp.Config.SectorIdentifierUri),
					"GrantTypes":                  ConsistOf(initialResp.Config.GrantTypes),
					"ResponseTypes":               ConsistOf(initialResp.Config.ResponseTypes),
					"Scopes":                      ConsistOf(initialResp.Config.Scopes),
					"Audiences":                   ConsistOf(initialResp.Config.Audiences),
					"TokenEndpointAuthMethod":     Equal(initialResp.Config.TokenEndpointAuthMethod),
					"TokenEndpointAuthSigningAlg": Equal(initialResp.Config.TokenEndpointAuthSigningAlg),
					"UserinfoSignedResponseAlg":   Equal(initialResp.Config.UserinfoSignedResponseAlg),
				})),
			})))).
			Return(&configpb.CreateOAuth2ApplicationResponse{
				Id:           initialResp.Id,
				ClientId:     initialResp.Config.ClientId,
				ClientSecret: clientSecret,
			}, nil)

		// 2x update
		mockConfigClient.EXPECT().
			UpdateOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(initialResp.Id),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter1stUpdateResp.Description.Value),
				})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"ClientId":                    Equal(""),
					"DisplayName":                 Equal(readAfter1stUpdateResp.Config.DisplayName),
					"Description":                 Equal(readAfter1stUpdateResp.Config.Description),
					"RedirectUris":                ConsistOf(readAfter1stUpdateResp.Config.RedirectUris),
					"Owner":                       Equal(readAfter1stUpdateResp.Config.Owner),
					"PolicyUri":                   Equal(readAfter1stUpdateResp.Config.PolicyUri),
					"AllowedCorsOrigins":          ConsistOf(readAfter1stUpdateResp.Config.AllowedCorsOrigins),
					"TermsOfServiceUri":           Equal(readAfter1stUpdateResp.Config.TermsOfServiceUri),
					"ClientUri":                   Equal(readAfter1stUpdateResp.Config.ClientUri),
					"LogoUri":                     Equal(readAfter1stUpdateResp.Config.LogoUri),
					"UserSupportEmailAddress":     Equal(readAfter1stUpdateResp.Config.UserSupportEmailAddress),
					"AdditionalContacts":          ConsistOf(readAfter1stUpdateResp.Config.AdditionalContacts),
					"SubjectType":                 Equal(readAfter1stUpdateResp.Config.SubjectType),
					"SectorIdentifierUri":         Equal(readAfter1stUpdateResp.Config.SectorIdentifierUri),
					"GrantTypes":                  ConsistOf(readAfter1stUpdateResp.Config.GrantTypes),
					"ResponseTypes":               ConsistOf(readAfter1stUpdateResp.Config.ResponseTypes),
					"Scopes":                      ConsistOf(readAfter1stUpdateResp.Config.Scopes),
					"Audiences":                   ConsistOf(readAfter1stUpdateResp.Config.Audiences),
					"TokenEndpointAuthMethod":     Equal(readAfter1stUpdateResp.Config.TokenEndpointAuthMethod),
					"TokenEndpointAuthSigningAlg": Equal(readAfter1stUpdateResp.Config.TokenEndpointAuthSigningAlg),
					"UserinfoSignedResponseAlg":   Equal(readAfter1stUpdateResp.Config.UserinfoSignedResponseAlg),
				})),
			})))).
			Return(&configpb.UpdateOAuth2ApplicationResponse{Id: initialResp.Id}, nil)

		mockConfigClient.EXPECT().
			UpdateOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialResp.Id),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter2ndUpdateResp.DisplayName),
				})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal("")})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"ClientId":                    Equal(""),
					"DisplayName":                 Equal(readAfter2ndUpdateResp.Config.DisplayName),
					"Description":                 Equal(readAfter2ndUpdateResp.Config.Description),
					"RedirectUris":                ConsistOf(readAfter2ndUpdateResp.Config.RedirectUris),
					"Owner":                       Equal(readAfter2ndUpdateResp.Config.Owner),
					"PolicyUri":                   Equal(readAfter2ndUpdateResp.Config.PolicyUri),
					"AllowedCorsOrigins":          ConsistOf(readAfter2ndUpdateResp.Config.AllowedCorsOrigins),
					"TermsOfServiceUri":           Equal(readAfter2ndUpdateResp.Config.TermsOfServiceUri),
					"ClientUri":                   Equal(readAfter2ndUpdateResp.Config.ClientUri),
					"LogoUri":                     Equal(readAfter2ndUpdateResp.Config.LogoUri),
					"UserSupportEmailAddress":     Equal(readAfter2ndUpdateResp.Config.UserSupportEmailAddress),
					"AdditionalContacts":          ConsistOf(readAfter2ndUpdateResp.Config.AdditionalContacts),
					"SubjectType":                 Equal(readAfter2ndUpdateResp.Config.SubjectType),
					"SectorIdentifierUri":         Equal(readAfter2ndUpdateResp.Config.SectorIdentifierUri),
					"GrantTypes":                  ConsistOf(readAfter2ndUpdateResp.Config.GrantTypes),
					"ResponseTypes":               ConsistOf(readAfter2ndUpdateResp.Config.ResponseTypes),
					"Scopes":                      ConsistOf(readAfter2ndUpdateResp.Config.Scopes),
					"Audiences":                   ConsistOf(readAfter2ndUpdateResp.Config.Audiences),
					"TokenEndpointAuthMethod":     Equal(readAfter2ndUpdateResp.Config.TokenEndpointAuthMethod),
					"TokenEndpointAuthSigningAlg": Equal(readAfter2ndUpdateResp.Config.TokenEndpointAuthSigningAlg),
					"UserinfoSignedResponseAlg":   Equal(readAfter2ndUpdateResp.Config.UserinfoSignedResponseAlg),
				})),
			})))).
			Return(&configpb.UpdateOAuth2ApplicationResponse{Id: initialResp.Id}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(initialResp.Id),
				})))).
				Times(4).
				Return(&configpb.ReadOAuth2ApplicationResponse{Oauth2Application: initialResp}, nil),

			mockConfigClient.EXPECT().
				ReadOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(initialResp.Id),
				})))).
				Times(3).
				Return(&configpb.ReadOAuth2ApplicationResponse{Oauth2Application: readAfter1stUpdateResp}, nil),

			mockConfigClient.EXPECT().
				ReadOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(initialResp.Id),
				})))).
				Times(5).
				Return(&configpb.ReadOAuth2ApplicationResponse{Oauth2Application: readAfter2ndUpdateResp}, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialResp.Id),
			})))).
			Return(&configpb.DeleteOAuth2ApplicationResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					Config: fmt.Sprintf(tfConfigDef, "", "", `
						customer_id = "`+customerID+`"
						app_space_id = "`+appSpaceID+`"
						oauth2_application_display_name = "config_display_name"
						oauth2_application_description = "config_description"
						redirect_uris = ["https://redirectUri1"]
						owner = "owner"
						policy_uri = "https://PolicyUri"
						allowed_cors_origins = ["https://allowedCorsOrigin1"]
						terms_of_service_uri =  "https://TermsOfServiceUri"
						client_uri = "https://ClientUri"
						logo_uri = "https://LogoUri"
						user_support_email_address = "UserSupport@EmailAddress.com"
						additional_contacts = ["AdditionalContact1"]
						subject_type = "public"
						sector_identifier_uri = "https://SectorIdentifierUri"
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid", "profile", "email", "phone"]
						audiences = ["00000000-0000-0000-0000-000000000000"]
						token_endpoint_auth_method = "client_secret_basic"
						token_endpoint_auth_signing_alg = "ES256"
						userinfo_signed_response_alg = "RS256"
					`),
					ExpectError: regexp.MustCompile("Value for unconfigurable attribute"),
				},
				{
					// Checking Create and Read (initialResp)
					Config: fmt.Sprintf(tfConfigDef, "", initialResp.Description.Value, `
						oauth2_application_display_name = "config_display_name"
						oauth2_application_description = "config_description"
						redirect_uris = ["https://redirectUri1"]
						owner = "owner"
						policy_uri = "https://PolicyUri"
						allowed_cors_origins = ["https://allowedCorsOrigin1"]
						terms_of_service_uri =  "https://TermsOfServiceUri"
						client_uri = "https://ClientUri"
						logo_uri = "https://LogoUri"
						user_support_email_address = "UserSupport@EmailAddress.com"
						additional_contacts = ["AdditionalContact1"]
						subject_type = "public"
						sector_identifier_uri = "https://SectorIdentifierUri"
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid", "profile", "email", "phone"]
						audiences = ["00000000-0000-0000-0000-000000000000"]
						token_endpoint_auth_method = "client_secret_basic"
						token_endpoint_auth_signing_alg = "ES256"
						userinfo_signed_response_alg = "RS256"
					`),
					Check: resource.ComposeTestCheckFunc(
						testOAuth2ApplicationResourceDataExists(resourceName, initialResp, clientSecret),
					),
				},
				{
					// Performs 1 read (initialResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: initialResp.Id,
				},
				{
					// Checking Read (initialResp), Update and Read(readAfter1stUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, "", readAfter1stUpdateResp.Description.Value, `
						oauth2_application_display_name = "config_display_name"
						oauth2_application_description = "config_description"
						redirect_uris = ["https://redirectUri1"]
						owner = "owner"
						policy_uri = "https://PolicyUri/updated"
						allowed_cors_origins = ["https://allowedCorsOrigin1"]
						terms_of_service_uri =  "https://TermsOfServiceUri"
						client_uri = "https://ClientUri"
						logo_uri = "https://LogoUri"
						user_support_email_address = "UserSupport@EmailAddress.com"
						additional_contacts = ["AdditionalContact1"]
						subject_type = "public"
						sector_identifier_uri = "https://SectorIdentifierUri"
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid"]
						audiences = ["00000000-0000-0000-0000-000000000000"]
						token_endpoint_auth_method = "client_secret_basic"
						token_endpoint_auth_signing_alg = "ES256"
						userinfo_signed_response_alg = "RS256"
					`),
					Check: resource.ComposeTestCheckFunc(
						testOAuth2ApplicationResourceDataExists(resourceName, readAfter1stUpdateResp, clientSecret),
					),
				},
				{
					// Checking Read(readAfter1stUpdateResp), Update and Read(readAfter2ndUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.DisplayName, "", `
						oauth2_application_display_name = "config_display_name_updated"
						oauth2_application_description = "config_description"
						redirect_uris = ["https://redirectUri1"]
						owner = "owner"
						policy_uri = "https://PolicyUri/updated"
						allowed_cors_origins = ["https://allowedCorsOrigin1"]
						terms_of_service_uri =  "https://TermsOfServiceUri"
						client_uri = "https://ClientUri"
						logo_uri = "https://LogoUri"
						user_support_email_address = "UserSupport@EmailAddress.com"
						additional_contacts = ["AdditionalContact1", "AdditionalContact2"]
						subject_type = "public"
						sector_identifier_uri = "https://SectorIdentifierUri"
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid"]
						audiences = ["00000000-0000-0000-0000-000000000000"]
						token_endpoint_auth_method = "client_secret_basic"
						token_endpoint_auth_signing_alg = "ES256"
						userinfo_signed_response_alg = "RS256"
					`),
					Check: resource.ComposeTestCheckFunc(
						testOAuth2ApplicationResourceDataExists(resourceName, readAfter2ndUpdateResp, clientSecret),
					),
				},
				{
					// Checking Read(readAfter2ndUpdateResp) -> no changes but tries to destroy with error
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.DisplayName, "", `
						oauth2_application_display_name = "config_display_name_updated"
						oauth2_application_description = "config_description"
						redirect_uris = ["https://redirectUri1"]
						owner = "owner"
						policy_uri = "https://PolicyUri/updated"
						allowed_cors_origins = ["https://allowedCorsOrigin1"]
						terms_of_service_uri =  "https://TermsOfServiceUri"
						client_uri = "https://ClientUri"
						logo_uri = "https://LogoUri"
						user_support_email_address = "UserSupport@EmailAddress.com"
						additional_contacts = ["AdditionalContact1", "AdditionalContact2"]
						subject_type = "public"
						sector_identifier_uri = "https://SectorIdentifierUri"
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid"]
						audiences = ["00000000-0000-0000-0000-000000000000"]
						token_endpoint_auth_method = "client_secret_basic"
						token_endpoint_auth_signing_alg = "ES256"
						userinfo_signed_response_alg = "RS256"
					`),
					Destroy:     true,
					ExpectError: regexp.MustCompile("Cannot destroy instance"),
				},
				{
					// Checking Read(readAfter2ndUpdateResp), Update (del protection, no API call)
					// and final Read (readAfter2ndUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.DisplayName, "", `
						deletion_protection=false
						oauth2_application_display_name = "config_display_name_updated"
						oauth2_application_description = "config_description"
						redirect_uris = ["https://redirectUri1"]
						owner = "owner"
						policy_uri = "https://PolicyUri/updated"
						allowed_cors_origins = ["https://allowedCorsOrigin1"]
						terms_of_service_uri =  "https://TermsOfServiceUri"
						client_uri = "https://ClientUri"
						logo_uri = "https://LogoUri"
						user_support_email_address = "UserSupport@EmailAddress.com"
						additional_contacts = ["AdditionalContact1", "AdditionalContact2"]
						subject_type = "public"
						sector_identifier_uri = "https://SectorIdentifierUri"
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid"]
						audiences = ["00000000-0000-0000-0000-000000000000"]
						token_endpoint_auth_method = "client_secret_basic"
						token_endpoint_auth_signing_alg = "ES256"
						userinfo_signed_response_alg = "RS256"
					`),
				},
			},
		})
	})
})

func testOAuth2ApplicationResourceDataExists(n string,
	data *configpb.OAuth2Application, clientSecret string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID != data.Id {
			return fmt.Errorf("ID does not match")
		}
		if v, has := rs.Primary.Attributes["customer_id"]; !has || v != data.CustomerId {
			return fmt.Errorf("invalid customer_id: %s", v)
		}
		if v, has := rs.Primary.Attributes["app_space_id"]; !has || v != data.AppSpaceId {
			return fmt.Errorf("invalid appspaceID: %s", v)
		}
		if v, has := rs.Primary.Attributes["oauth2_provider_id"]; !has || v != data.Oauth2ProviderId {
			return fmt.Errorf("invalid oauth2_provider_id: %s", v)
		}
		if v, has := rs.Primary.Attributes["name"]; !has || v != data.Name {
			return fmt.Errorf("invalid name: %s", v)
		}
		if v, has := rs.Primary.Attributes["display_name"]; !has || v != data.DisplayName {
			//
			if data.DisplayName != data.Name {
				return fmt.Errorf("invalid display name: %s", v)
			}
		}
		if v, has := rs.Primary.Attributes["description"]; !has || v != data.Description.GetValue() {
			return fmt.Errorf("invalid description: %s", v)
		}
		if v, has := rs.Primary.Attributes["client_id"]; !has || v != data.Config.ClientId {
			return fmt.Errorf("invalid client_id: %s", v)
		}
		if v, has := rs.Primary.Attributes["client_secret"]; !has || v != clientSecret {
			return fmt.Errorf("invalid client_secret: %s", v)
		}
		if v, has := rs.Primary.Attributes["oauth2_application_display_name"]; !has || v != data.Config.DisplayName {
			return fmt.Errorf("invalid oauth2_application_display_name: %s", v)
		}
		if v, has := rs.Primary.Attributes["oauth2_application_description"]; !has || v != data.Config.Description {
			return fmt.Errorf("invalid oauth2_application_description: %s", v)
		}
		if err := testStringArray(rs.Primary.Attributes,
			data.GetConfig().RedirectUris, "redirect_uris"); err != nil {
			return err
		}
		if v, has := rs.Primary.Attributes["owner"]; !has || v != data.Config.Owner {
			return fmt.Errorf("invalid owner: %s", v)
		}
		if v, has := rs.Primary.Attributes["policy_uri"]; !has || v != data.Config.PolicyUri {
			return fmt.Errorf("invalid policy_uri: %s", v)
		}
		if err := testStringArray(rs.Primary.Attributes,
			data.GetConfig().AllowedCorsOrigins, "allowed_cors_origins"); err != nil {
			return err
		}
		if v, has := rs.Primary.Attributes["terms_of_service_uri"]; !has || v != data.Config.TermsOfServiceUri {
			return fmt.Errorf("invalid terms_of_service_uri: %s", v)
		}
		if v, has := rs.Primary.Attributes["client_uri"]; !has || v != data.Config.ClientUri {
			return fmt.Errorf("invalid client_uri: %s", v)
		}
		if v, has := rs.Primary.Attributes["logo_uri"]; !has || v != data.Config.LogoUri {
			return fmt.Errorf("invalid logo_uri: %s", v)
		}
		if v, has := rs.Primary.Attributes["user_support_email_address"]; !has ||
			v != data.Config.UserSupportEmailAddress {
			return fmt.Errorf("invalid user_support_email_address: %s", v)
		}
		if err := testStringArray(rs.Primary.Attributes,
			data.GetConfig().AdditionalContacts, "additional_contacts"); err != nil {
			return err
		}
		if err := testSubjectType(rs.Primary.Attributes,
			data.GetConfig().SubjectType); err != nil {
			return err
		}
		if v, has := rs.Primary.Attributes["sector_identifier_uri"]; !has || v != data.Config.SectorIdentifierUri {
			return fmt.Errorf("invalid sector_identifier_uri: %s", v)
		}
		if err := testGrantTypes(rs.Primary.Attributes, data.GetConfig().GrantTypes); err != nil {
			return err
		}
		if err := testResponseTypes(rs.Primary.Attributes, data.GetConfig().ResponseTypes); err != nil {
			return err
		}
		if err := testStringArray(rs.Primary.Attributes, data.GetConfig().Scopes, "scopes"); err != nil {
			return err
		}
		if err := testStringArray(rs.Primary.Attributes,
			data.GetConfig().Audiences, "audiences"); err != nil {
			return err
		}
		if err := testTokenEndpointAuthMethod(rs.Primary.Attributes,
			data.GetConfig().TokenEndpointAuthMethod); err != nil {
			return err
		}
		if v, has := rs.Primary.Attributes["token_endpoint_auth_signing_alg"]; !has ||
			v != data.Config.TokenEndpointAuthSigningAlg {
			return fmt.Errorf("invalid token_endpoint_auth_signing_alg: %s", v)
		}
		if v, has := rs.Primary.Attributes["userinfo_signed_response_alg"]; !has ||
			v != data.Config.UserinfoSignedResponseAlg {
			return fmt.Errorf("invalid userinfo_signed_response_alg: %s", v)
		}
		return nil
	}
}

func testSubjectType(attrs map[string]string, subjectType configpb.ClientSubjectType) error {
	v, has := attrs["subject_type"]
	if !has {
		return fmt.Errorf("invalid subject_type: %s", v)
	}

	s, found := indykite.OAuth2ClientSubjectTypes[v]
	if !found {
		return fmt.Errorf("invalid subject_type: %s", v)
	}

	if subjectType != s {
		return fmt.Errorf("invalid subject_type: %s", v)
	}

	return nil
}

func testTokenEndpointAuthMethod(attrs map[string]string,
	tokenEndpointAuthMethod configpb.TokenEndpointAuthMethod) error {
	v, has := attrs["token_endpoint_auth_method"]
	if !has {
		return fmt.Errorf("invalid token_endpoint_auth_method: %s", v)
	}

	m, found := indykite.OAuth2TokenEndpointAuthMethods[v]
	if !found {
		return fmt.Errorf("invalid token_endpoint_auth_method: %s", v)
	}

	if tokenEndpointAuthMethod != m {
		return fmt.Errorf("invalid token_endpoint_auth_method: %s", v)
	}

	return nil
}
