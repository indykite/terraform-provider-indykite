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

var _ = Describe("DataSource OAuth2Provider", func() {
	const resourceName = "data.indykite_oauth2_provider.wonka-bars-oauth2-provider-2"
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
					"Id": Equal(resp.Id),
				})))).
				Times(5).
				Return(&configpb.ReadOAuth2ProviderResponse{Oauth2Provider: resp}, nil),
		)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
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
			return fmt.Errorf("ID does not match")
		}
		if v, has := rs.Primary.Attributes["customer_id"]; !has || v != data.CustomerId {
			return fmt.Errorf("invalid customer_id: %s", v)
		}
		if v, has := rs.Primary.Attributes["name"]; !has || v != data.Name {
			return fmt.Errorf("invalid name: %s", v)
		}
		if v, has := rs.Primary.Attributes["app_space_id"]; !has || v != data.AppSpaceId {
			return fmt.Errorf("invalid issuer_id: %s", v)
		}
		if v, has := rs.Primary.Attributes["display_name"]; !has || v != data.DisplayName {
			return fmt.Errorf("invalid display name: %s", v)
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
