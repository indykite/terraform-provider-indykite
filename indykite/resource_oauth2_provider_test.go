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
	"strconv"

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

var _ = Describe("Resource OAuth2 Provider", func() {
	const resourceName = "indykite_oauth2_provider.wonka-bars-oauth2-provider"
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
			})))).
			Return(&configpb.CreateOAuth2ProviderResponse{Id: initialResp.Id}, nil)

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
			})))).
			Return(&configpb.UpdateOAuth2ProviderResponse{Id: initialResp.Id}, nil)

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
			})))).
			Return(&configpb.UpdateOAuth2ProviderResponse{Id: initialResp.Id}, nil)

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
				"Id": Equal(initialResp.Id),
			})))).
			Return(&configpb.DeleteOAuth2ProviderResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
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
			return fmt.Errorf("ID does not match")
		}
		if v, has := rs.Primary.Attributes["customer_id"]; !has || v != data.CustomerId {
			return fmt.Errorf("invalid customer_id: %s", v)
		}
		if v, has := rs.Primary.Attributes["app_space_id"]; !has || v != data.AppSpaceId {
			return fmt.Errorf("invalid appspaceID: %s", v)
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

		if err := testGrantTypes(rs.Primary.Attributes, data.GetConfig().GrantTypes); err != nil {
			return err
		}

		if err := testResponseTypes(rs.Primary.Attributes, data.GetConfig().ResponseTypes); err != nil {
			return err
		}

		if err := testStringArray(rs.Primary.Attributes, data.GetConfig().Scopes, "scopes"); err != nil {
			return err
		}

		if err := testTokenEndpointAuthMethods(rs.Primary.Attributes,
			data.GetConfig().TokenEndpointAuthMethod); err != nil {
			return err
		}
		if err := testStringArray(rs.Primary.Attributes,
			data.GetConfig().TokenEndpointAuthSigningAlg, "token_endpoint_auth_signing_alg"); err != nil {
			return err
		}

		if err := testStringArray(rs.Primary.Attributes,
			data.GetConfig().RequestUris, "request_uris"); err != nil {
			return err
		}

		if v, has := rs.Primary.Attributes["request_object_signing_alg"]; !has ||
			v != data.Config.RequestObjectSigningAlg {
			return fmt.Errorf("invalid request_object_signing_alg: %s", v)
		}

		if err := testStringMap(rs.Primary.Attributes,
			data.GetConfig().FrontChannelLoginUri, "front_channel_login_uri"); err != nil {
			return err
		}

		if err := testStringMap(rs.Primary.Attributes,
			data.GetConfig().FrontChannelConsentUri, "front_channel_consent_uri"); err != nil {
			return err
		}

		return nil
	}
}

func testGrantTypes(attrs map[string]string, grantTypes []configpb.GrantType) error {
	if grantTypes == nil {
		return nil
	}
	key := "grant_types"
	cnt, _ := strconv.Atoi(attrs[key+".#"])
	if cnt != len(grantTypes) {
		return fmt.Errorf("expected %d grant_types, got %d, ", cnt, len(grantTypes))
	}

	for i := 0; i < cnt; i++ {
		curKey := fmt.Sprintf("%s.%d", key, i)
		v, has := attrs[curKey]
		if !has {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
		grantType, found := indykite.OAuth2GrantTypes[v]
		if !found {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
		if !containsGrantType(grantTypes, grantType) {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
	}
	return nil
}

func testResponseTypes(attrs map[string]string, responseTypes []configpb.ResponseType) error {
	if responseTypes == nil {
		return nil
	}
	key := "response_types"
	cnt, _ := strconv.Atoi(attrs[key+".#"])
	if cnt != len(responseTypes) {
		return fmt.Errorf("expected %d response_types, got %d, ", cnt, len(responseTypes))
	}

	for i := 0; i < cnt; i++ {
		curKey := fmt.Sprintf("%s.%d", key, i)
		v, has := attrs[curKey]
		if !has {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
		responseType, found := indykite.OAuth2ResponseTypes[v]
		if !found {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
		if !containsResponseType(responseTypes, responseType) {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
	}
	return nil
}

func testTokenEndpointAuthMethods(attrs map[string]string,
	endpointAuthMethods []configpb.TokenEndpointAuthMethod) error {
	if endpointAuthMethods == nil {
		return nil
	}
	key := "token_endpoint_auth_method"
	cnt, _ := strconv.Atoi(attrs[key+".#"])
	if cnt != len(endpointAuthMethods) {
		return fmt.Errorf("expected %d grant_types, got %d, ", cnt, len(endpointAuthMethods))
	}

	for i := 0; i < cnt; i++ {
		curKey := fmt.Sprintf("%s.%d", key, i)
		v, has := attrs[curKey]
		if !has {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
		tokenEndpointAuthMethod, found := indykite.OAuth2TokenEndpointAuthMethods[v]
		if !found {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
		if !containsEndpointAuthMethod(endpointAuthMethods, tokenEndpointAuthMethod) {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
	}
	return nil
}

func testStringArray(attrs map[string]string, array []string, key string) error {
	if array == nil {
		return nil
	}
	cnt, _ := strconv.Atoi(attrs[key+".#"])
	if cnt != len(array) {
		return fmt.Errorf("expected %d %s, got %d, ", cnt, key, len(array))
	}

	for i := 0; i < cnt; i++ {
		curKey := fmt.Sprintf("%s.%d", key, i)
		if v, has := attrs[curKey]; !has || !contains(array, v) {
			return fmt.Errorf("invalid %s: %s", curKey, v)
		}
	}
	return nil
}

func testStringMap(attrs map[string]string, dataMap map[string]string, key string) error {
	if dataMap == nil {
		return nil
	}
	cnt, _ := strconv.Atoi(attrs[key+".%"])

	if cnt != len(dataMap) {
		return fmt.Errorf("expected %d %s, got %d, ", cnt, key, len(dataMap))
	}

	for k, val := range dataMap {
		if v, has := attrs[key+"."+k]; !has || v != val {
			return fmt.Errorf("invalid key: %s", v)
		}
	}
	return nil
}

func containsGrantType(s []configpb.GrantType, grantType configpb.GrantType) bool {
	for _, v := range s {
		if v == grantType {
			return true
		}
	}
	return false
}

func containsResponseType(s []configpb.ResponseType, responseType configpb.ResponseType) bool {
	for _, v := range s {
		if v == responseType {
			return true
		}
	}
	return false
}

func containsEndpointAuthMethod(s []configpb.TokenEndpointAuthMethod,
	endpointAuthMethod configpb.TokenEndpointAuthMethod) bool {
	for _, v := range s {
		if v == endpointAuthMethod {
			return true
		}
	}
	return false
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
