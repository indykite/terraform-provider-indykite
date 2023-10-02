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

var _ = Describe("DataSource OAuth2Application", func() {
	const resourceName = "data.indykite_oauth2_application.wonka-bars-oauth2-application-2"
	var (
		mockCtrl         *gomock.Controller
		mockConfigClient *configm.MockConfigManagementAPIClient
		provider         *schema.Provider
		mockedBookmark   string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)
		mockedBookmark = "for-oauth2-app-reads" + uuid.NewRandom().String() // Bookmark must be longer than 40 chars
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
					"Id":        Equal(resp.Id),
					"Bookmarks": ConsistOf(mockedBookmark),
				})))).
				Times(5).
				Return(&configpb.ReadOAuth2ApplicationResponse{Oauth2Application: resp}, nil),
		)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
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

			"oauth2_application_id":           Equal(data.Id),
			"oauth2_provider_id":              Equal(data.Oauth2ProviderId),
			"oauth2_application_display_name": Equal(data.Config.DisplayName),
			"oauth2_application_description":  Equal(data.Config.Description),

			"client_id":                       Equal(data.Config.ClientId),
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
