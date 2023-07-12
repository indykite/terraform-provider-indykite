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
	"strconv"
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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource OAuth2 Client", func() {
	const resourceName = "indykite_oauth2_client.wonka"
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
		mockedBookmark = "for-oauth2-client" + uuid.NewRandom().String()
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
		oauth2ClientMinimalConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				TenantId:    tenantID,
				Id:          sampleID,
				Name:        "wonka-oauth2-client",
				DisplayName: "Wonka-Slugworth client",
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_Oauth2ClientConfig{
					Oauth2ClientConfig: &configpb.OAuth2ClientConfig{
						ProviderType: configpb.ProviderType_PROVIDER_TYPE_GOOGLE_COM,
						ClientId:     "my-own-client-id",
						ClientSecret: "client-secret-for-google",
						AuthStyle:    configpb.AuthStyle_AUTH_STYLE_AUTO_DETECT,
					},
				},
			},
		}

		// Those are just to verify invalid types are reported as error to user
		invalidProviderTypeConfigResp := proto.Clone(oauth2ClientMinimalConfigResp).(*configpb.ReadConfigNodeResponse)
		invalidProviderTypeConfigResp.ConfigNode.
			GetOauth2ClientConfig().ProviderType = configpb.ProviderType_PROVIDER_TYPE_INVALID

		invalidAuthStyleConfigResp := proto.Clone(oauth2ClientMinimalConfigResp).(*configpb.ReadConfigNodeResponse)
		invalidAuthStyleConfigResp.ConfigNode.GetOauth2ClientConfig().AuthStyle = configpb.AuthStyle_AUTH_STYLE_INVALID

		oauth2ClientFullConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				TenantId:    tenantID,
				Id:          sampleID,
				Name:        "wonka-oauth2-client",
				Description: wrapperspb.String("Description of the best OAuth2Client by Wonka inc."),
				CreateTime:  oauth2ClientMinimalConfigResp.ConfigNode.CreateTime,
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_Oauth2ClientConfig{
					Oauth2ClientConfig: &configpb.OAuth2ClientConfig{
						ProviderType: configpb.ProviderType_PROVIDER_TYPE_APPLE_COM,
						ClientId:     "my-own-client-id",

						RedirectUri:           []string{"http://unsecure.indykite.com", "https://indykite.com"},
						DefaultScopes:         []string{"openid", "profile"},
						AllowedScopes:         []string{"openid", "profile", "email", "phone"},
						AllowSignup:           true,
						Issuer:                "https://example.com",
						AuthorizationEndpoint: "https://example.com/auth",
						TokenEndpoint:         "https://example.com/token",
						DiscoveryUrl:          "https://example.com/discovery",
						UserinfoEndpoint:      "https://example.com/user-info",
						JwksUri:               "https://example.com/json-web-key",
						ImageUrl:              "https://example.com/oauth2.png",
						AuthStyle:             configpb.AuthStyle_AUTH_STYLE_IN_PARAMS,
						Tenant:                "some-tenant",
						HostedDomain:          "indykite.com",
						PrivateKeyId:          "my-own-kid",
						TeamId:                "some-team",
					},
				},
			},
		}

		createBM := "created-oauth2-client" + uuid.NewRandom().String()
		updateBM := "updated-oauth2-client" + uuid.NewRandom().String()
		deleteBM := "deleted-oauth2-client" + uuid.NewRandom().String()

		// Create
		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(oauth2ClientMinimalConfigResp.ConfigNode.Name),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					oauth2ClientMinimalConfigResp.ConfigNode.DisplayName,
				)})),
				"Description": BeNil(),
				"Location":    Equal(tenantID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"Oauth2ClientConfig": test.EqualProto(&configpb.OAuth2ClientConfig{
						ProviderType: configpb.ProviderType_PROVIDER_TYPE_GOOGLE_COM,
						ClientId:     "my-own-client-id",
						ClientSecret: "client-secret-for-google",
						AuthStyle:    configpb.AuthStyle_AUTH_STYLE_AUTO_DETECT,
					}),
				})),
				"Bookmarks": ConsistOf(mockedBookmark),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         oauth2ClientMinimalConfigResp.ConfigNode.Id,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
				Bookmark:   createBM,
			}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(oauth2ClientMinimalConfigResp.ConfigNode.Id),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					oauth2ClientFullConfigResp.ConfigNode.Description.GetValue(),
				)})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"Oauth2ClientConfig": test.EqualProto(&configpb.OAuth2ClientConfig{
						ProviderType:          configpb.ProviderType_PROVIDER_TYPE_APPLE_COM,
						ClientId:              "my-own-client-id",
						RedirectUri:           []string{"http://unsecure.indykite.com", "https://indykite.com"},
						DefaultScopes:         []string{"openid", "profile"},
						AllowedScopes:         []string{"openid", "profile", "email", "phone"},
						AllowSignup:           true,
						Issuer:                "https://example.com",
						AuthorizationEndpoint: "https://example.com/auth",
						TokenEndpoint:         "https://example.com/token",
						DiscoveryUrl:          "https://example.com/discovery",
						UserinfoEndpoint:      "https://example.com/user-info",
						JwksUri:               "https://example.com/json-web-key",
						ImageUrl:              "https://example.com/oauth2.png",
						Tenant:                "some-tenant",
						HostedDomain:          "indykite.com",
						AuthStyle:             configpb.AuthStyle_AUTH_STYLE_IN_PARAMS,
						PrivateKeyPem: []byte("-----BEGIN PRIVATE KEY-----\n" +
							"MEECAQAwEwYHKoZIzj0CAQYIKoZIzj0DAQcEJzAlAgEBBCCwZbKaPfdXYNDxejZD\n" +
							"kQFncFsMtK1qo1e/Ol5gtxk1vw==\n-----END PRIVATE KEY-----"),
						PrivateKeyId: "my-own-kid",
						TeamId:       "some-team",
					}),
				})),
				"Bookmarks": ConsistOf(mockedBookmark, createBM),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{
				Id:       oauth2ClientMinimalConfigResp.ConfigNode.Id,
				Bookmark: updateBM,
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(oauth2ClientMinimalConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM),
				})))).
				Times(3).
				Return(oauth2ClientMinimalConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(oauth2ClientMinimalConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM),
				})))).
				Return(invalidProviderTypeConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(oauth2ClientMinimalConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM),
				})))).
				Return(invalidAuthStyleConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(oauth2ClientMinimalConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM),
				})))).
				Return(oauth2ClientMinimalConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(oauth2ClientMinimalConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM),
				})))).
				Times(2).
				Return(oauth2ClientFullConfigResp, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":        Equal(oauth2ClientMinimalConfigResp.ConfigNode.Id),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{Bookmark: deleteBM}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Error cases should be always first, easier to avoid missing mocks or incomplete plan
				{
					Config: `resource "indykite_oauth2_client" "wonka" {
						name = "wonka-oauth2-client"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "location" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_oauth2_client" "wonka" {
						location = "` + customerID + `"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "name" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_oauth2_client" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-oauth2-client"

						provider_type = "google.com"
						client_id = "abcdefghijkl"
						auth_style = "auto_detect"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "customer_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_oauth2_client" "wonka" {
						location = "` + customerID + `"
						app_space_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-oauth2-client"

						provider_type = "google.com"
						client_id = "abcdefghijkl"
						auth_style = "auto_detect"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "app_space_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_oauth2_client" "wonka" {
						location = "` + customerID + `"
						tenant_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-oauth2-client"

						provider_type = "google.com"
						client_id = "abcdefghijkl"
						auth_style = "auto_detect"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "tenant_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_oauth2_client" "wonka" {
						location = "something-invalid"
						name = "wonka-oauth2-client"

						provider_type = "google.com"
						client_id = "abcdefghijkl"
						auth_style = "auto_detect"
					}
					`,
					ExpectError: regexp.MustCompile(`expected to have 'gid:' prefix`),
				},
				{
					Config: `resource "indykite_oauth2_client" "wonka" {
						location = "` + customerID + `"
						name = "Invalid Name @#$"

						provider_type = "google.com"
						client_id = "abcdefghijkl"
						auth_style = "auto_detect"
					}
					`,
					ExpectError: regexp.MustCompile(`Value can have lowercase letters, digits, or hyphens.`),
				},
				{
					Config: `resource "indykite_oauth2_client" "wonka" {
						location = "` + customerID + `"
						name = "wonka-oauth2-client"

						provider_type = "google.com"
						client_id = "abcdefghijkl"
						auth_style = "invalid-style"
					}
					`,
					ExpectError: regexp.MustCompile(`expected auth_style to be one of \[.*\], got invalid-style`),
				},
				{
					Config: `resource "indykite_oauth2_client" "wonka" {
						location = "` + customerID + `"
						name = "wonka-oauth2-client"
						display_name = "Wonka-Slugworth client"

						provider_type = "brontosaurus"
						client_id = "abcdefghijkl"
						auth_style = "in_params"
					}
					`,
					ExpectError: regexp.MustCompile(`expected provider_type to be one of \[.*\], got brontosaurus`),
				},

				// ---- Run mocked tests here ----
				{
					// Minimal config - Checking Create and Read (oauth2ClientMinimalConfigResp)
					Config: `resource "indykite_oauth2_client" "wonka" {
						location = "` + tenantID + `"
						name = "wonka-oauth2-client"
						display_name = "Wonka-Slugworth client"

						provider_type = "google.com"
						client_id = "my-own-client-id"
						client_secret = "client-secret-for-google"
						auth_style = "auto_detect"
					}
					`,
					Check: resource.ComposeTestCheckFunc(testOAuth2ClientResourceDataExists(
						resourceName,
						oauth2ClientMinimalConfigResp,
						"google.com",
						"auto_detect",
						map[string]OmegaMatcher{
							"client_secret": Equal("client-secret-for-google"),
						},
					)),
				},
				{
					// Performs 1 read (oauth2ClientMinimalConfigResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: oauth2ClientMinimalConfigResp.ConfigNode.Id,
				},
				{
					// Performs 1 read (oauth2ClientMinimalConfigResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: oauth2ClientMinimalConfigResp.ConfigNode.Id,
					ExpectError: regexp.MustCompile(
						`unsupported OAuth2 Provider Type((?s).*)IndyKite plugin error, please report this issue`),
				},
				{
					// Performs 1 read (oauth2ClientMinimalConfigResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: oauth2ClientMinimalConfigResp.ConfigNode.Id,
					ExpectError: regexp.MustCompile(
						`unsupported OAuth2 AuthStyle((?s).*)IndyKite plugin error, please report this issue to us`),
				},
				{
					// Checking Read(oauth2ClientMinimalConfigResp), Update and Read(oauth2ClientFullConfigResp)
					Config: `resource "indykite_oauth2_client" "wonka" {
						location = "` + tenantID + `"
						name = "wonka-oauth2-client"
						description = "Description of the best OAuth2Client by Wonka inc."

						provider_type = "apple.com"
						client_id = "my-own-client-id"
						redirect_uri = ["http://unsecure.indykite.com", "https://indykite.com"]
						default_scopes = ["openid", "profile"]
						allowed_scopes = ["openid", "profile", "email", "phone"]
						allow_signup = true
						issuer = "https://example.com"
						authorization_endpoint = "https://example.com/auth"
						token_endpoint = "https://example.com/token"
						discovery_url = "https://example.com/discovery"
						userinfo_endpoint = "https://example.com/user-info"
						jwks_uri = "https://example.com/json-web-key"
						image_url = "https://example.com/oauth2.png"
						tenant = "some-tenant"
						hosted_domain = "indykite.com"
						auth_style = "in_params"
						private_key_pem = <<-EOT
							-----BEGIN PRIVATE KEY-----
							MEECAQAwEwYHKoZIzj0CAQYIKoZIzj0DAQcEJzAlAgEBBCCwZbKaPfdXYNDxejZD
							kQFncFsMtK1qo1e/Ol5gtxk1vw==
							-----END PRIVATE KEY-----
						EOT
						private_key_id = "my-own-kid"
						team_id = "some-team"
					}
					`,
					Check: resource.ComposeTestCheckFunc(testOAuth2ClientResourceDataExists(
						resourceName,
						oauth2ClientFullConfigResp,
						"apple.com",
						"in_params",
						map[string]OmegaMatcher{
							"client_secret":   BeEmpty(),
							"private_key_pem": HavePrefix("-----BEGIN PRIVATE KEY-----"),
						},
					)),
				},
			},
		})
	})
})

func testOAuth2ClientResourceDataExists(
	n string,
	data *configpb.ReadConfigNodeResponse,
	providerType, authStyle string,
	extraMatch map[string]OmegaMatcher,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != data.ConfigNode.Id {
			return errors.New("ID does not match")
		}
		attrs := rs.Primary.Attributes

		oauth2Cfg := data.ConfigNode.GetOauth2ClientConfig()

		keys := Keys{
			"id": Equal(data.ConfigNode.Id),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"location":     Not(BeEmpty()), // Response does not return this
			"customer_id":  Equal(data.ConfigNode.CustomerId),
			"app_space_id": Equal(data.ConfigNode.AppSpaceId),
			"tenant_id":    Equal(data.ConfigNode.TenantId),
			"name":         Equal(data.ConfigNode.Name),
			"display_name": Equal(data.ConfigNode.DisplayName),
			"description":  Equal(data.ConfigNode.GetDescription().GetValue()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),

			"provider_type":          Equal(providerType),
			"client_id":              Equal(oauth2Cfg.ClientId),
			"allow_signup":           Equal(strconv.FormatBool(oauth2Cfg.AllowSignup)),
			"issuer":                 Equal(oauth2Cfg.Issuer),
			"authorization_endpoint": Equal(oauth2Cfg.AuthorizationEndpoint),
			"token_endpoint":         Equal(oauth2Cfg.TokenEndpoint),
			"discovery_url":          Equal(oauth2Cfg.DiscoveryUrl),
			"userinfo_endpoint":      Equal(oauth2Cfg.UserinfoEndpoint),
			"jwks_uri":               Equal(oauth2Cfg.JwksUri),
			"image_url":              Equal(oauth2Cfg.ImageUrl),
			"tenant":                 Equal(oauth2Cfg.Tenant),
			"hosted_domain":          Equal(oauth2Cfg.HostedDomain),
			"auth_style":             Equal(authStyle),
			"private_key_id":         Equal(oauth2Cfg.PrivateKeyId),
			"team_id":                Equal(oauth2Cfg.TeamId),
		}

		addStringArrayToKeys(keys, "redirect_uri", oauth2Cfg.RedirectUri)
		addStringArrayToKeys(keys, "default_scopes", oauth2Cfg.DefaultScopes)
		addStringArrayToKeys(keys, "allowed_scopes", oauth2Cfg.AllowedScopes)

		for k, v := range extraMatch {
			keys[k] = v
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), attrs)
	}
}
