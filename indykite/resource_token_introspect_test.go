// Copyright (c) 2024 IndyKite
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
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"
	"github.com/pborman/uuid"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource TokenIntrospect", func() {
	const (
		resourceName = "indykite_token_introspect.development"
	)
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
		mockedBookmark = "for-token-introspect-cfg" + uuid.NewRandom().String()
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

	It("Test CRUD of Token Introspect configuration", func() {
		tfConfigDef :=
			`resource "indykite_token_introspect" "development" {
				location = "%s"
				name = "%s"
				%s
			}`
		expectedResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				Id:          sampleID,
				Name:        "my-first-token-introspect",
				DisplayName: "Display name of Token Introspect configuration",
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_TokenIntrospectConfig{
					TokenIntrospectConfig: &configpb.TokenIntrospectConfig{
						TokenMatcher: &configpb.TokenIntrospectConfig_Jwt{Jwt: &configpb.TokenIntrospectConfig_JWT{
							Issuer:   "https://example.com",
							Audience: "audience-id",
						}},
						Validation: &configpb.TokenIntrospectConfig_Online_{
							Online: &configpb.TokenIntrospectConfig_Online{
								CacheTtl: durationpb.New(600 * time.Second),
							},
						},
						ClaimsMapping: map[string]*configpb.TokenIntrospectConfig_Claim{
							"email": {Selector: "mail"},
							"name":  {Selector: "full_name"},
						},
						IkgNodeType:   "MyUser",
						PerformUpsert: true,
					},
				},
			},
		}
		expectedUpdatedResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				Id:          sampleID,
				Name:        "my-first-token-introspect",
				Description: wrapperspb.String("token introspect description"),
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				CreateTime:  expectedResp.ConfigNode.CreateTime,
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_TokenIntrospectConfig{
					TokenIntrospectConfig: &configpb.TokenIntrospectConfig{
						TokenMatcher: &configpb.TokenIntrospectConfig_Opaque_{
							Opaque: &configpb.TokenIntrospectConfig_Opaque{}},
						Validation: &configpb.TokenIntrospectConfig_Online_{
							Online: &configpb.TokenIntrospectConfig_Online{
								UserinfoEndpoint: "https://data.example.com/userinfo",
							},
						},
						IkgNodeType: "MyUser",
					},
				},
			},
		}
		expectedUpdatedResp2 := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				Id:          sampleID,
				Name:        "my-first-token-introspect",
				Description: wrapperspb.String("token introspect description"),
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				CreateTime:  expectedResp.ConfigNode.CreateTime,
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_TokenIntrospectConfig{
					TokenIntrospectConfig: &configpb.TokenIntrospectConfig{
						TokenMatcher: &configpb.TokenIntrospectConfig_Jwt{Jwt: &configpb.TokenIntrospectConfig_JWT{
							Issuer:   "https://example.com",
							Audience: "audience-id",
						}},
						Validation: &configpb.TokenIntrospectConfig_Offline_{
							Offline: &configpb.TokenIntrospectConfig_Offline{
								PublicJwks: [][]byte{
									[]byte(`{"kid":"abc","use":"sig","alg":"RS256","n":"--nothing-real-just-random-xyqwerasf--","kty":"RSA"}`), //nolint:lll
									[]byte(`{"kid":"jkl","use":"sig","alg":"RS256","n":"--nothing-real-just-random-435asdf43--","kty":"RSA"}`), //nolint:lll
								},
							},
						},
						IkgNodeType: "MyUser",
					},
				},
			},
		}

		// Create
		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(expectedResp.ConfigNode.Name),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					expectedResp.ConfigNode.DisplayName,
				)})),
				"Description": BeNil(),
				"Location":    Equal(appSpaceID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"TokenIntrospectConfig": test.EqualProto(expectedResp.GetConfigNode().GetTokenIntrospectConfig()),
				})),
				"Bookmarks": ConsistOf(mockedBookmark),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         sampleID,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
			}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(sampleID),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					expectedUpdatedResp.ConfigNode.Description.GetValue(),
				)})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"TokenIntrospectConfig": test.EqualProto(
						expectedUpdatedResp.GetConfigNode().GetTokenIntrospectConfig()),
				})),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{Id: sampleID}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(sampleID),
				"DisplayName": BeNil(),
				"Description": BeNil(),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"TokenIntrospectConfig": test.EqualProto(
						expectedUpdatedResp2.GetConfigNode().GetTokenIntrospectConfig()),
				})),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{Id: sampleID}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(sampleID),
				})))).
				Times(4).
				Return(expectedResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(sampleID),
				})))).
				Times(2).
				Return(expectedUpdatedResp, nil),
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(sampleID),
				})))).
				Times(3).
				Return(expectedUpdatedResp2, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(sampleID),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{}, nil)

		testResourceDataExists := func(
			n string,
			data *configpb.ReadConfigNodeResponse,
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

				keys := Keys{
					"id": Equal(data.ConfigNode.Id),
					"%":  Not(BeEmpty()), // This is Terraform helper

					"location":     Equal(data.ConfigNode.AppSpaceId), // Token Introspect is always on AppSpace level
					"customer_id":  Equal(data.ConfigNode.CustomerId),
					"app_space_id": Equal(data.ConfigNode.AppSpaceId),
					"name":         Equal(data.ConfigNode.Name),
					"display_name": Equal(data.ConfigNode.DisplayName),
					"description":  Equal(data.ConfigNode.GetDescription().GetValue()),
					"create_time":  Not(BeEmpty()),
					"update_time":  Not(BeEmpty()),

					"ikg_node_type": Equal(data.GetConfigNode().GetTokenIntrospectConfig().GetIkgNodeType()),
					"perform_upsert": Equal(strconv.FormatBool(
						data.GetConfigNode().GetTokenIntrospectConfig().GetPerformUpsert())),
				}

				switch validation := data.GetConfigNode().GetTokenIntrospectConfig().GetValidation().(type) {
				case *configpb.TokenIntrospectConfig_Offline_:
					keys["offline_validation.#"] = Equal("1")
					keys["offline_validation.0.%"] = Equal("1")
					keys["online_validation.#"] = Equal("0")
					addStringArrayToKeys(keys, "offline_validation.0.public_jwks", validation.Offline.GetPublicJwks())
				case *configpb.TokenIntrospectConfig_Online_:
					keys["offline_validation.#"] = Equal("0")
					keys["online_validation.#"] = Equal("1")
					keys["online_validation.0.%"] = Equal("2")

					keys["online_validation.0.cache_ttl"] = Equal(
						strconv.FormatInt(validation.Online.GetCacheTtl().GetSeconds(), 10))
					keys["online_validation.0.user_info_endpoint"] = Equal(validation.Online.GetUserinfoEndpoint())
				}

				switch matcher := data.GetConfigNode().GetTokenIntrospectConfig().GetTokenMatcher().(type) {
				case *configpb.TokenIntrospectConfig_Jwt:
					keys["opaque_matcher.#"] = Equal("0")
					keys["jwt_matcher.#"] = Equal("1")
					keys["jwt_matcher.0.%"] = Equal("2")
					keys["jwt_matcher.0.issuer"] = Equal(matcher.Jwt.GetIssuer())
					keys["jwt_matcher.0.audience"] = Equal(matcher.Jwt.GetAudience())
				case *configpb.TokenIntrospectConfig_Opaque_:
					keys["jwt_matcher.#"] = Equal("0")
					keys["opaque_matcher.#"] = Equal("1")
					keys["opaque_matcher.0.%"] = Equal("0")
				}

				rawClaimsMapping := make(map[string]string)
				for k, v := range data.GetConfigNode().GetTokenIntrospectConfig().GetClaimsMapping() {
					rawClaimsMapping[k] = v.GetSelector()
				}
				addStringMapMatcherToKeys(keys, "claims_mapping", rawClaimsMapping, true)

				return convertOmegaMatcherToError(MatchAllKeys(keys), attrs)
			}
		}

		validSettings := `
		opaque_matcher {}
		online_validation {
			user_info_endpoint = "https://data.example.com/userinfo"
		}
		ikg_node_type = "MyUser"
		`

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors case must be always first
				{
					Config:      fmt.Sprintf(tfConfigDef, "ccc", "name", validSettings),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name", `ikg_node_type = "MyUser"`),
					ExpectError: regexp.MustCompile(
						"one of `offline_validation,online_validation` must be(\\s+)specified"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", `ikg_node_type = "MyUser"`),
					ExpectError: regexp.MustCompile("one of `jwt_matcher,opaque_matcher` must be specified"),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name",
						`opaque_matcher {}
						online_validation {}
						ikg_node_type = "MyUser"`),
					ExpectError: regexp.MustCompile(
						"`online_validation.0.user_info_endpoint,opaque_matcher` must be specified"),
				},

				{
					// Checking Create and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-token-introspect",
						`display_name = "Display name of Token Introspect configuration"
						jwt_matcher {
							issuer = "https://example.com"
							audience = "audience-id"
						}
						online_validation {
							cache_ttl = 600
						}
						claims_mapping = {
							"email" = "mail",
							"name" = "full_name"
						}
						ikg_node_type = "MyUser"
						perform_upsert = true
						`,
					),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(resourceName, expectedResp),
					),
				},
				{
					// Performs 1 read
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: sampleID,
				},
				{
					// Checking Update and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-token-introspect",
						`description = "token introspect description"
						opaque_matcher { }
						online_validation {
							user_info_endpoint = "https://data.example.com/userinfo"
						}
						ikg_node_type = "MyUser"
					`,
					),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(resourceName, expectedUpdatedResp),
					),
				},
				{
					// Checking Update and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-token-introspect",
						`description = "token introspect description"
						jwt_matcher {
							issuer = "https://example.com"
							audience = "audience-id"
						}
						offline_validation {
							public_jwks = [
								jsonencode({
									"kid": "abc",
									"use": "sig",
									"alg": "RS256",
									"n": "--nothing-real-just-random-xyqwerasf--",
									"kty": "RSA"
								}),
								jsonencode({
									"kid": "jkl",
									"use": "sig",
									"alg": "RS256",
									"n": "--nothing-real-just-random-435asdf43--",
									"kty": "RSA"
								})
							]
						}
						ikg_node_type = "MyUser"
					`,
					),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(resourceName, expectedUpdatedResp2),
					),
				},
			},
		})
	})
})
