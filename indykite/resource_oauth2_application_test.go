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
	"errors"
	"fmt"
	"regexp"
	"sync"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"
	"github.com/pborman/uuid"
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
		mockCtrl         *gomock.Controller
		mockConfigClient *configm.MockConfigManagementAPIClient
		provider         *schema.Provider
		mockedBookmark   string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)

		// Bookmark must be longer than 40 chars - have just 1 added before the first write to test all cases
		mockedBookmark = "for-oauth2-app" + uuid.NewRandom().String()
		bmOnce := &sync.Once{}

		provider = indykite.Provider()
		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc =
			func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
				client, _ := config.NewTestClient(ctx, mockConfigClient)
				ctx = indykite.WithClient(ctx, client)
				i, d := cfgFunc(ctx, data)
				// ConfigureContextFunc is called repeatedly, add initial bookmark just once
				bmOnce.Do(func() {
					i.(*indykite.ClientContext).AddBookmarks(mockedBookmark)
				})
				return i, d
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

		createBM := "created-oauth2-app" + uuid.NewRandom().String()
		updateBM := "updated-oauth2-app" + uuid.NewRandom().String()
		updateBM2 := "updated-oauth2-app-2" + uuid.NewRandom().String()
		deleteBM := "deleted-oauth2-app" + uuid.NewRandom().String()

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
				"Bookmarks": ConsistOf(mockedBookmark),
			})))).
			Return(&configpb.CreateOAuth2ApplicationResponse{
				Id:           initialResp.Id,
				ClientId:     initialResp.Config.ClientId,
				ClientSecret: clientSecret,
				Bookmark:     createBM,
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
				"Bookmarks": ConsistOf(mockedBookmark, createBM),
			})))).
			Return(&configpb.UpdateOAuth2ApplicationResponse{
				Id:       initialResp.Id,
				Bookmark: updateBM,
			}, nil)

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
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM),
			})))).
			Return(&configpb.UpdateOAuth2ApplicationResponse{
				Id:       initialResp.Id,
				Bookmark: updateBM2,
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(initialResp.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM),
				})))).
				Times(4).
				Return(&configpb.ReadOAuth2ApplicationResponse{Oauth2Application: initialResp}, nil),

			mockConfigClient.EXPECT().
				ReadOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(initialResp.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM),
				})))).
				Times(3).
				Return(&configpb.ReadOAuth2ApplicationResponse{Oauth2Application: readAfter1stUpdateResp}, nil),

			mockConfigClient.EXPECT().
				ReadOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(initialResp.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM, updateBM2),
				})))).
				Times(5).
				Return(&configpb.ReadOAuth2ApplicationResponse{Oauth2Application: readAfter2ndUpdateResp}, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":        Equal(initialResp.Id),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM, updateBM2),
			})))).
			Return(&configpb.DeleteOAuth2ApplicationResponse{
				Bookmark: deleteBM,
			}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
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
			return errors.New("ID does not match")
		}

		keys := Keys{
			"id": Equal(data.Id),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"customer_id":         Equal(data.CustomerId),
			"app_space_id":        Equal(data.AppSpaceId),
			"oauth2_provider_id":  Equal(data.Oauth2ProviderId),
			"name":                Equal(data.Name),
			"display_name":        Equal(data.DisplayName),
			"description":         Equal(data.Description.GetValue()),
			"create_time":         Not(BeEmpty()),
			"update_time":         Not(BeEmpty()),
			"deletion_protection": Not(BeEmpty()),

			"client_id":                       Equal(data.Config.ClientId),
			"client_secret":                   Equal(clientSecret),
			"oauth2_application_display_name": Equal(data.Config.DisplayName),
			"oauth2_application_description":  Equal(data.Config.Description),

			"owner":                           Equal(data.Config.Owner),
			"policy_uri":                      Equal(data.Config.PolicyUri),
			"terms_of_service_uri":            Equal(data.Config.TermsOfServiceUri),
			"client_uri":                      Equal(data.Config.ClientUri),
			"logo_uri":                        Equal(data.Config.LogoUri),
			"user_support_email_address":      Equal(data.Config.UserSupportEmailAddress),
			"sector_identifier_uri":           Equal(data.Config.SectorIdentifierUri),
			"token_endpoint_auth_signing_alg": Equal(data.Config.TokenEndpointAuthSigningAlg),
			"userinfo_signed_response_alg":    Equal(data.Config.UserinfoSignedResponseAlg),
			"subject_type":                    Equal(indykite.OAuth2ClientSubjectTypesReverse[data.Config.SubjectType]),
			"token_endpoint_auth_method": Equal(
				indykite.OAuth2TokenEndpointAuthMethodsReverse[data.Config.TokenEndpointAuthMethod],
			),
		}
		addStringArrayToKeys(keys, "redirect_uris", data.GetConfig().RedirectUris)
		addStringArrayToKeys(keys, "allowed_cors_origins", data.GetConfig().AllowedCorsOrigins)
		addStringArrayToKeys(keys, "additional_contacts", data.GetConfig().AdditionalContacts)
		addStringArrayToKeys(keys, "scopes", data.GetConfig().Scopes)
		addStringArrayToKeys(keys, "audiences", data.GetConfig().Audiences)

		strGrantTypes := []string{}
		oauth2GrantTypesReverse := indykite.ReverseProtoEnumMap(indykite.OAuth2GrantTypes)
		for _, v := range data.Config.GrantTypes {
			strGrantTypes = append(strGrantTypes, oauth2GrantTypesReverse[v])
		}
		addStringArrayToKeys(keys, "grant_types", strGrantTypes)

		strResponseTypes := []string{}
		oauth2ResponseTypesReverse := indykite.ReverseProtoEnumMap(indykite.OAuth2ResponseTypes)
		for _, v := range data.Config.ResponseTypes {
			strResponseTypes = append(strResponseTypes, oauth2ResponseTypesReverse[v])
		}
		addStringArrayToKeys(keys, "response_types", strResponseTypes)

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}
