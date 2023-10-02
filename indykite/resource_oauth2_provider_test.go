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

var _ = Describe("Resource OAuth2 Provider", func() {
	const resourceName = "indykite_oauth2_provider.wonka-bars-oauth2-provider"
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
		mockedBookmark = "for-oauth2-provider" + uuid.NewRandom().String()
		bmOnce := &sync.Once{}

		provider = indykite.Provider()
		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc =
			func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
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
		oauth2ProviderID := "gid:AAAAD2luZHlraURlgAAEDwABCDE"
		// Terraform create config must be in sync with returned data in expectedAppSpace and expectedUpdatedAppSpace
		// otherwise "After applying this test step, the plan was not empty" error is thrown
		tfConfigDef :=
			`resource "indykite_oauth2_provider" "wonka-bars-oauth2-provider" {
				app_space_id = "` + appSpaceID + `"
				name = "acme"
				display_name = "%s"
				description = "%s"
				%s
			}`

		initialResp := &configpb.OAuth2Provider{
			CustomerId:  customerID,
			AppSpaceId:  appSpaceID,
			Id:          oauth2ProviderID,
			Name:        "acme",
			DisplayName: "acme",
			Description: wrapperspb.String("Just some OAuth2Provider description"),
			CreateTime:  timestamppb.Now(),
			UpdateTime:  timestamppb.Now(),
			Config: &configpb.OAuth2ProviderConfig{
				GrantTypes: []configpb.GrantType{
					configpb.GrantType_GRANT_TYPE_AUTHORIZATION_CODE,
				},
				ResponseTypes: []configpb.ResponseType{
					configpb.ResponseType_RESPONSE_TYPE_TOKEN,
				},
				Scopes: []string{"openid", "profile", "email", "phone"},
				TokenEndpointAuthMethod: []configpb.TokenEndpointAuthMethod{
					configpb.TokenEndpointAuthMethod_TOKEN_ENDPOINT_AUTH_METHOD_CLIENT_SECRET_BASIC},
				TokenEndpointAuthSigningAlg: []string{"ES256"},
				RequestUris:                 []string{"https://request_uri"},
				RequestObjectSigningAlg:     "ES256",
				FrontChannelLoginUri: map[string]string{
					"front_channel_1": "https://www.google.com",
					"front_channel_2": "https://www.google.com",
				},
				FrontChannelConsentUri: map[string]string{"front_channel_1": "https://www.google.com"},
			},
		}

		readAfter1stUpdateResp := &configpb.OAuth2Provider{
			CustomerId:  customerID,
			Id:          initialResp.Id,
			AppSpaceId:  initialResp.AppSpaceId,
			Name:        "acme",
			DisplayName: "acme",
			Description: wrapperspb.String("Another OAuth2Provider description"),
			CreateTime:  initialResp.CreateTime,
			UpdateTime:  timestamppb.Now(),
			Config: &configpb.OAuth2ProviderConfig{
				GrantTypes: []configpb.GrantType{
					configpb.GrantType_GRANT_TYPE_AUTHORIZATION_CODE,
				},
				ResponseTypes: []configpb.ResponseType{
					configpb.ResponseType_RESPONSE_TYPE_TOKEN,
				},
				Scopes: []string{"openid"},
				TokenEndpointAuthMethod: []configpb.TokenEndpointAuthMethod{
					configpb.TokenEndpointAuthMethod_TOKEN_ENDPOINT_AUTH_METHOD_CLIENT_SECRET_BASIC},
				TokenEndpointAuthSigningAlg: []string{"ES256"},
				RequestUris:                 []string{"https://request_uri"},
				RequestObjectSigningAlg:     "ES256",
				FrontChannelLoginUri:        map[string]string{"front_channel_1": "https://www.google.com/updated"},
				FrontChannelConsentUri:      map[string]string{"front_channel_1": "https://www.google.com"},
			},
		}
		readAfter2ndUpdateResp := &configpb.OAuth2Provider{
			CustomerId:  customerID,
			Id:          initialResp.Id,
			AppSpaceId:  initialResp.AppSpaceId,
			Name:        "acme",
			DisplayName: "Some new display name",
			Description: nil,
			CreateTime:  initialResp.CreateTime,
			UpdateTime:  timestamppb.Now(),
			Config: &configpb.OAuth2ProviderConfig{
				GrantTypes: []configpb.GrantType{
					configpb.GrantType_GRANT_TYPE_AUTHORIZATION_CODE,
				},
				ResponseTypes: []configpb.ResponseType{
					configpb.ResponseType_RESPONSE_TYPE_TOKEN,
				},
				Scopes: []string{"openid"},
				TokenEndpointAuthMethod: []configpb.TokenEndpointAuthMethod{
					configpb.TokenEndpointAuthMethod_TOKEN_ENDPOINT_AUTH_METHOD_CLIENT_SECRET_BASIC},
				TokenEndpointAuthSigningAlg: []string{"ES256", "ES512"},
				RequestUris:                 []string{"https://request_uri"},
				RequestObjectSigningAlg:     "ES512",
				FrontChannelLoginUri:        map[string]string{"front_channel_1": "https://www.google.com/updated"},
				FrontChannelConsentUri:      map[string]string{"front_channel_1": "https://www.google.com"},
			},
		}

		createBM := "created-oauth2-provider" + uuid.NewRandom().String()
		updateBM := "updated-oauth2-provider" + uuid.NewRandom().String()
		updateBM2 := "updated-oauth2-provider-2" + uuid.NewRandom().String()
		deleteBM := "deleted-oauth2-provider" + uuid.NewRandom().String()

		// Create
		mockConfigClient.EXPECT().
			CreateOAuth2Provider(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"AppSpaceId":  Equal(appSpaceID),
				"Name":        Equal(initialResp.Name),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(initialResp.Description.Value),
				})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"GrantTypes":                  ConsistOf(initialResp.Config.GrantTypes),
					"ResponseTypes":               ConsistOf(initialResp.Config.ResponseTypes),
					"Scopes":                      ConsistOf(initialResp.Config.Scopes),
					"TokenEndpointAuthMethod":     ConsistOf(initialResp.Config.TokenEndpointAuthMethod),
					"TokenEndpointAuthSigningAlg": ConsistOf(initialResp.Config.TokenEndpointAuthSigningAlg),
					"RequestUris":                 ConsistOf(initialResp.Config.RequestUris),
					"RequestObjectSigningAlg":     Equal(initialResp.Config.RequestObjectSigningAlg),
					"FrontChannelLoginUri": HaveKeyWithValue(
						"front_channel_1", initialResp.Config.FrontChannelLoginUri["front_channel_1"],
					),
					"FrontChannelConsentUri": HaveKeyWithValue(
						"front_channel_1", initialResp.Config.FrontChannelConsentUri["front_channel_1"],
					),
				})),
				"Bookmarks": ConsistOf(mockedBookmark),
			})))).
			Return(&configpb.CreateOAuth2ProviderResponse{Id: initialResp.Id, Bookmark: createBM}, nil)

		// 2x update
		mockConfigClient.EXPECT().
			UpdateOAuth2Provider(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(initialResp.Id),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter1stUpdateResp.Description.Value),
				})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"GrantTypes":                  ConsistOf(readAfter1stUpdateResp.Config.GrantTypes),
					"ResponseTypes":               ConsistOf(readAfter1stUpdateResp.Config.ResponseTypes),
					"Scopes":                      ConsistOf(readAfter1stUpdateResp.Config.Scopes),
					"TokenEndpointAuthMethod":     ConsistOf(readAfter1stUpdateResp.Config.TokenEndpointAuthMethod),
					"TokenEndpointAuthSigningAlg": ConsistOf(readAfter1stUpdateResp.Config.TokenEndpointAuthSigningAlg),
					"RequestUris":                 ConsistOf(readAfter1stUpdateResp.Config.RequestUris),
					"RequestObjectSigningAlg":     Equal(readAfter1stUpdateResp.Config.RequestObjectSigningAlg),
					"FrontChannelLoginUri": HaveKeyWithValue(
						"front_channel_1", readAfter1stUpdateResp.Config.FrontChannelLoginUri["front_channel_1"],
					),
					"FrontChannelConsentUri": HaveKeyWithValue(
						"front_channel_1", readAfter1stUpdateResp.Config.FrontChannelConsentUri["front_channel_1"],
					),
				})),
				"Bookmarks": ConsistOf(mockedBookmark, createBM),
			})))).
			Return(&configpb.UpdateOAuth2ProviderResponse{Id: initialResp.Id, Bookmark: updateBM}, nil)

		mockConfigClient.EXPECT().
			UpdateOAuth2Provider(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialResp.Id),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter2ndUpdateResp.DisplayName),
				})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal("")})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"GrantTypes":                  ConsistOf(readAfter2ndUpdateResp.Config.GrantTypes),
					"ResponseTypes":               ConsistOf(readAfter2ndUpdateResp.Config.ResponseTypes),
					"Scopes":                      ConsistOf(readAfter2ndUpdateResp.Config.Scopes),
					"TokenEndpointAuthMethod":     ConsistOf(readAfter2ndUpdateResp.Config.TokenEndpointAuthMethod),
					"TokenEndpointAuthSigningAlg": ConsistOf(readAfter2ndUpdateResp.Config.TokenEndpointAuthSigningAlg),
					"RequestUris":                 ConsistOf(readAfter2ndUpdateResp.Config.RequestUris),
					"RequestObjectSigningAlg":     Equal(readAfter2ndUpdateResp.Config.RequestObjectSigningAlg),
					"FrontChannelLoginUri": HaveKeyWithValue(
						"front_channel_1", readAfter2ndUpdateResp.Config.FrontChannelLoginUri["front_channel_1"],
					),
					"FrontChannelConsentUri": HaveKeyWithValue(
						"front_channel_1", readAfter2ndUpdateResp.Config.FrontChannelConsentUri["front_channel_1"],
					),
				})),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM),
			})))).
			Return(&configpb.UpdateOAuth2ProviderResponse{Id: initialResp.Id, Bookmark: updateBM2}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadOAuth2Provider(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(initialResp.Id),
				})))).
				Times(4).
				Return(&configpb.ReadOAuth2ProviderResponse{Oauth2Provider: initialResp}, nil),

			mockConfigClient.EXPECT().
				ReadOAuth2Provider(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(initialResp.Id),
				})))).
				Times(3).
				Return(&configpb.ReadOAuth2ProviderResponse{Oauth2Provider: readAfter1stUpdateResp}, nil),

			mockConfigClient.EXPECT().
				ReadOAuth2Provider(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(initialResp.Id),
				})))).
				Times(5).
				Return(&configpb.ReadOAuth2ProviderResponse{Oauth2Provider: readAfter2ndUpdateResp}, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteOAuth2Provider(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":        Equal(initialResp.Id),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM, updateBM2),
			})))).
			Return(&configpb.DeleteOAuth2ProviderResponse{Bookmark: deleteBM}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					Config: fmt.Sprintf(tfConfigDef, "", "", `
						customer_id = "`+customerID+`"
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid", "profile", "email", "phone"]
						token_endpoint_auth_method = ["client_secret_basic"]
						token_endpoint_auth_signing_alg = ["ES256"]
						request_uris = ["https://request_uri"]
						request_object_signing_alg = "ES256"
						front_channel_login_uri = { "front_channel_1" = "https://www.google.com" }
						front_channel_consent_uri = { "front_channel_1" = "https://www.google.com" }
					`),
					ExpectError: regexp.MustCompile("Value for unconfigurable attribute"),
				},
				{
					// Checking Create and Read (initialResp)
					Config: fmt.Sprintf(tfConfigDef, "", initialResp.Description.Value, `
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid", "profile", "email", "phone"]
						token_endpoint_auth_method = ["client_secret_basic"]
						token_endpoint_auth_signing_alg = ["ES256"]
						request_uris = ["https://request_uri"]
						request_object_signing_alg = "ES256"
						front_channel_login_uri = {
							"front_channel_1" = "https://www.google.com",
							"front_channel_2" = "https://www.google.com",
						}
						front_channel_consent_uri = { "front_channel_1" = "https://www.google.com" }
					`),
					Check: resource.ComposeTestCheckFunc(
						testOAuth2ProviderResourceDataExists(resourceName, initialResp),
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
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid"]
						token_endpoint_auth_method = ["client_secret_basic"]
						token_endpoint_auth_signing_alg = ["ES256"]
						request_uris = ["https://request_uri"]
						request_object_signing_alg = "ES256"
						front_channel_login_uri = { "front_channel_1" = "https://www.google.com/updated" }
						front_channel_consent_uri = { "front_channel_1" = "https://www.google.com" }
					`),
					Check: resource.ComposeTestCheckFunc(
						testOAuth2ProviderResourceDataExists(resourceName, readAfter1stUpdateResp),
					),
				},
				{
					// Checking Read(readAfter1stUpdateResp), Update and Read(readAfter2ndUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.DisplayName, "", `
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid"]
						token_endpoint_auth_method = ["client_secret_basic"]
						token_endpoint_auth_signing_alg = ["ES256", "ES512"]
						request_uris = ["https://request_uri"]
						request_object_signing_alg = "ES512"
						front_channel_login_uri = { "front_channel_1" = "https://www.google.com/updated" }
						front_channel_consent_uri = { "front_channel_1" = "https://www.google.com" }
					`),
					Check: resource.ComposeTestCheckFunc(
						testOAuth2ProviderResourceDataExists(resourceName, readAfter2ndUpdateResp),
					),
				},
				{
					// Checking Read(readAfter2ndUpdateResp) -> no changes but tries to destroy with error
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.DisplayName, "", `
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid"]
						token_endpoint_auth_method = ["client_secret_basic"]
						token_endpoint_auth_signing_alg = ["ES256", "ES512"]
						request_uris = ["https://request_uri"]
						request_object_signing_alg = "ES512"
						front_channel_login_uri = { "front_channel_1" = "https://www.google.com/updated" }
						front_channel_consent_uri = { "front_channel_1" = "https://www.google.com" }
					`),
					Destroy:     true,
					ExpectError: regexp.MustCompile("Cannot destroy instance"),
				},
				{
					// Checking Read(readAfter2ndUpdateResp), Update (del protection, no API call)
					// and final Read (readAfter2ndUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.DisplayName, "", `
						deletion_protection=false
						grant_types = ["authorization_code"]
						response_types = ["token"]
						scopes = ["openid"]
						token_endpoint_auth_method = ["client_secret_basic"]
						token_endpoint_auth_signing_alg = ["ES256", "ES512"]
						request_uris = ["https://request_uri"]
						request_object_signing_alg = "ES512"
						front_channel_login_uri = { "front_channel_1" = "https://www.google.com/updated" }
						front_channel_consent_uri = { "front_channel_1" = "https://www.google.com" }
					`),
				},
			},
		})
	})
})

func testOAuth2ProviderResourceDataExists(n string, data *configpb.OAuth2Provider) resource.TestCheckFunc {
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

			"customer_id":  Equal(data.CustomerId),
			"app_space_id": Equal(data.AppSpaceId),
			"name":         Equal(data.Name),
			"display_name": Equal(data.DisplayName),
			"description":  Equal(data.Description.GetValue()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),

			"request_object_signing_alg": Equal(data.Config.RequestObjectSigningAlg),
		}
		addStringMapMatcherToKeys(keys, "front_channel_login_uri", data.GetConfig().FrontChannelLoginUri)
		addStringMapMatcherToKeys(keys, "front_channel_consent_uri", data.GetConfig().FrontChannelConsentUri)

		addStringArrayToKeys(keys, "scopes", data.GetConfig().Scopes)
		addStringArrayToKeys(keys, "token_endpoint_auth_signing_alg", data.GetConfig().TokenEndpointAuthSigningAlg)
		addStringArrayToKeys(keys, "request_uris", data.GetConfig().RequestUris)

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

		strTokenEndpointAuthMethod := []string{}
		for _, v := range data.Config.TokenEndpointAuthMethod {
			strTokenEndpointAuthMethod = append(
				strTokenEndpointAuthMethod,
				indykite.OAuth2TokenEndpointAuthMethodsReverse[v])
		}
		addStringArrayToKeys(keys, "token_endpoint_auth_method", strTokenEndpointAuthMethod)

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), rs.Primary.Attributes)
	}
}
