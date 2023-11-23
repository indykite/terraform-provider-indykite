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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"
	"github.com/pborman/uuid"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("DataSource OAuth2Provider", func() {
	const resourceName = "data.indykite_oauth2_provider.wonka-bars-oauth2-provider-2"
	var (
		mockCtrl         *gomock.Controller
		mockConfigClient *configm.MockConfigManagementAPIClient
		provider         *schema.Provider
		mockedBookmark   string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)
		// Bookmark must be longer than 40 chars
		mockedBookmark = "for-oauth2-provider-reads" + uuid.NewRandom().String()
		bmOnce := &sync.Once{}

		provider = indykite.Provider()
		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc =
			func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
				client, _ := config.NewTestClient(ctx, mockConfigClient)
				ctx = indykite.WithClient(ctx, client)
				i, d := cfgFunc(ctx, data)
				bmOnce.Do(func() {
					i.(*indykite.ClientContext).AddBookmarks(mockedBookmark)
				})
				return i, d
			}
	})

	It("Test load by ID", func() {
		oauth2ProviderID := "gid:AAAAD2luZHlraURlgAAEDwABCDE"
		resp := &configpb.OAuth2Provider{
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
				FrontChannelLoginUri:        map[string]string{"front_channel_1": "https://www.google.com"},
				FrontChannelConsentUri:      map[string]string{"front_channel_1": "https://www.google.com"},
			},
		}

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadOAuth2Provider(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(resp.Id),
					"Bookmarks": ConsistOf(mockedBookmark),
				})))).
				Times(5).
				Return(&configpb.ReadOAuth2ProviderResponse{Oauth2Provider: resp}, nil),
		)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					Config: `data "indykite_oauth2_provider" "wonka-bars-oauth2-provider-2" {
						display_name = "anything"
					}`,
					ExpectError: regexp.MustCompile("\"oauth2_provider_id\" is required"),
				},
				// Success test cases
				{
					Check: resource.ComposeTestCheckFunc(testOAuth2ProviderDataExists(resourceName, resp)),
					Config: `data "indykite_oauth2_provider" "wonka-bars-oauth2-provider-2" {
						oauth2_provider_id = "` + oauth2ProviderID + `"
					}`,
				},
			},
		})
	})
})

func testOAuth2ProviderDataExists(n string, data *configpb.OAuth2Provider) resource.TestCheckFunc {
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
