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

var _ = Describe("DataSource OAuth2Application", func() {
	const resourceName = "data.indykite_oauth2_application.wonka-bars-oauth2-application-2"
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
		oauth2ApplicationID := "gid:AAAAD2luZHlraURlgAAEDwAABCC"
		resp := &configpb.OAuth2Application{
			CustomerId:       customerID,
			AppSpaceId:       appSpaceID,
			Oauth2ProviderId: "gid:AAAAD2luZHlraURlgAAEDwABCDE",
			Id:               oauth2ApplicationID,
			Name:             "acme",
			DisplayName:      "acme",
			Description:      wrapperspb.String("Just some OAuth2Application description"),
			CreateTime:       timestamppb.Now(),
			UpdateTime:       timestamppb.Now(),
			Config: &configpb.OAuth2ApplicationConfig{
				ClientId:                "gid:AAAAD2luZHlraURlgAAEDwEDCBA",
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
				Audiences: []string{"Audience1"},
				TokenEndpointAuthMethod: configpb.
					TokenEndpointAuthMethod_TOKEN_ENDPOINT_AUTH_METHOD_CLIENT_SECRET_BASIC,
				TokenEndpointAuthSigningAlg: "ES256",
				UserinfoSignedResponseAlg:   "ES256",
			},
		}

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadOAuth2Application(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(resp.Id),
				})))).
				Times(5).
				Return(&configpb.ReadOAuth2ApplicationResponse{Oauth2Application: resp}, nil),
		)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					Config: `data "indykite_oauth2_application" "wonka-bars-oauth2-application-2" {
						display_name = "anything"
					}`,
					ExpectError: regexp.MustCompile("\"oauth2_application_id\" is required"),
				},
				// Success test cases
				{
					Check: resource.ComposeTestCheckFunc(testOAuth2ApplicationDataExists(resourceName, resp)),
					Config: `data "indykite_oauth2_application" "wonka-bars-oauth2-application-2" {
						oauth2_application_id = "` + oauth2ApplicationID + `"
					}`,
				},
			},
		})
	})
})

func testOAuth2ApplicationDataExists(n string, data *configpb.OAuth2Application) resource.TestCheckFunc {
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
		if v, has := rs.Primary.Attributes["oauth2_provider_id"]; !has || v != data.Oauth2ProviderId {
			return fmt.Errorf("invalid oauth2_provider_id: %s", v)
		}
		if v, has := rs.Primary.Attributes["display_name"]; !has || v != data.DisplayName {
			return fmt.Errorf("invalid display name: %s", v)
		}
		if v, has := rs.Primary.Attributes["description"]; !has || v != data.Description.GetValue() {
			return fmt.Errorf("invalid description: %s", v)
		}
		if v, has := rs.Primary.Attributes["client_id"]; !has || v != data.Config.ClientId {
			return fmt.Errorf("invalid client_id: %s", v)
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
