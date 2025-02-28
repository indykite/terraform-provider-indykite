// Copyright (c) 2025 IndyKite
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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource TrustScoreProfile", func() {
	const (
		resourceName = "indykite_trust_score_profile.development"
	)
	var (
		mockCtrl         *gomock.Controller
		mockConfigClient *configm.MockConfigManagementAPIClient
		provider         *schema.Provider
		tfConfigDef      string
		validSettings    string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)

		provider = indykite.Provider()
		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc =
			func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
				client, _ := config.NewTestClient(ctx, mockConfigClient)
				ctx = indykite.WithClient(ctx, client)
				return cfgFunc(ctx, data)
			}

		tfConfigDef = `resource "indykite_trust_score_profile" "development" {
			location = "%s"
			name = "%s"
			%s
		}`

		validSettings = `
		node_classification = "Person"
		dimensions {
		  name   = "NAME_FRESHNESS"
		  weight = 1.0
		}
		schedule = "UPDATE_FREQUENCY_DAILY"
		`
	})

	Describe("Error cases", func() {
		It("should handle invalid configurations", func() {
			resource.Test(GinkgoT(), resource.TestCase{
				Providers: map[string]*schema.Provider{
					"indykite": provider,
				},
				Steps: []resource.TestStep{
					{
						Config:      fmt.Sprintf(tfConfigDef, "ccc", "name", validSettings),
						ExpectError: regexp.MustCompile("Invalid ID value"),
					},
					{
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name",
							`node_classification = "Person"
							`),
						ExpectError: regexp.MustCompile(
							`The argument "schedule" is required, but no definition was found.`),
					},
				},
			})
		})
	})

	Describe("Valid configurations", func() {
		It("Test CRUD of TrustScoreProfile configuration", func() {
			expectedResp := &configpb.ReadConfigNodeResponse{
				ConfigNode: &configpb.ConfigNode{
					Id:          sampleID,
					Name:        "my-first-trust-score-profile",
					DisplayName: "Display name of TrustScoreProfile configuration",
					CustomerId:  customerID,
					AppSpaceId:  appSpaceID,
					CreateTime:  timestamppb.Now(),
					UpdateTime:  timestamppb.Now(),
					Config: &configpb.ConfigNode_TrustScoreProfileConfig{
						TrustScoreProfileConfig: &configpb.TrustScoreProfileConfig{
							NodeClassification: "Person",
							Dimensions: []*configpb.TrustScoreDimension{
								{
									Name:   configpb.TrustScoreDimension_NAME_FRESHNESS,
									Weight: 1.0,
								},
								{
									Name:   configpb.TrustScoreDimension_NAME_ORIGIN,
									Weight: 1.0,
								},
							},
							Schedule: configpb.TrustScoreProfileConfig_UPDATE_FREQUENCY_DAILY,
						},
					},
				},
			}
			expectedUpdatedResp := &configpb.ReadConfigNodeResponse{
				ConfigNode: &configpb.ConfigNode{
					Id:          sampleID,
					Name:        "my-first-trust-score-profile",
					DisplayName: "Display name of TrustScoreProfile configuration",
					CustomerId:  customerID,
					AppSpaceId:  appSpaceID,
					CreateTime:  expectedResp.GetConfigNode().GetCreateTime(),
					UpdateTime:  timestamppb.Now(),
					Config: &configpb.ConfigNode_TrustScoreProfileConfig{
						TrustScoreProfileConfig: &configpb.TrustScoreProfileConfig{
							NodeClassification: "Person",
							Dimensions: []*configpb.TrustScoreDimension{
								{
									Name:   configpb.TrustScoreDimension_NAME_COMPLETENESS,
									Weight: 1.0,
								},
								{
									Name:   configpb.TrustScoreDimension_NAME_ORIGIN,
									Weight: 1.0,
								},
							},
							Schedule: configpb.TrustScoreProfileConfig_UPDATE_FREQUENCY_SIX_HOURS,
						},
					},
				},
			}

			expectedUpdatedResp2 := &configpb.ReadConfigNodeResponse{
				ConfigNode: &configpb.ConfigNode{
					Id:          sampleID2,
					Name:        "my-first-trust-score-profile",
					DisplayName: "Display name of TrustScoreProfile configuration",
					CustomerId:  customerID,
					AppSpaceId:  appSpaceID,
					CreateTime:  expectedResp.GetConfigNode().GetCreateTime(),
					UpdateTime:  timestamppb.Now(),
					Config: &configpb.ConfigNode_TrustScoreProfileConfig{
						TrustScoreProfileConfig: &configpb.TrustScoreProfileConfig{
							NodeClassification: "Employee",
							Dimensions: []*configpb.TrustScoreDimension{
								{
									Name:   configpb.TrustScoreDimension_NAME_FRESHNESS,
									Weight: 1.0,
								},
							},
							Schedule: configpb.TrustScoreProfileConfig_UPDATE_FREQUENCY_SIX_HOURS,
						},
					},
				},
			}

			// Create
			mockConfigClient.EXPECT().
				CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Name": Equal(expectedResp.GetConfigNode().GetName()),
					"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
						expectedResp.GetConfigNode().GetDisplayName(),
					)})),
					"Description": BeNil(),
					"Location":    Equal(appSpaceID),
					"Config": PointTo(MatchFields(IgnoreExtras, Fields{
						"TrustScoreProfileConfig": test.EqualProto(
							expectedResp.GetConfigNode().GetTrustScoreProfileConfig()),
					})),
				})))).
				Return(&configpb.CreateConfigNodeResponse{
					Id:         sampleID,
					CreateTime: timestamppb.Now(),
					UpdateTime: timestamppb.Now(),
				}, nil)

			// Update
			mockConfigClient.EXPECT().
				UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(sampleID),
					"Config": PointTo(MatchFields(IgnoreExtras, Fields{
						"TrustScoreProfileConfig": test.EqualProto(
							expectedUpdatedResp.GetConfigNode().GetTrustScoreProfileConfig()),
					})),
				})))).
				Return(&configpb.UpdateConfigNodeResponse{Id: sampleID}, nil)

			// Update2
			mockConfigClient.EXPECT().
				DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(sampleID),
				})))).
				Return(&configpb.DeleteConfigNodeResponse{}, nil)
			mockConfigClient.EXPECT().
				CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Name": Equal(expectedUpdatedResp2.GetConfigNode().GetName()),
					"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
						expectedUpdatedResp2.GetConfigNode().GetDisplayName(),
					)})),
					"Description": BeNil(),
					"Location":    Equal(appSpaceID),
					"Config": PointTo(MatchFields(IgnoreExtras, Fields{
						"TrustScoreProfileConfig": test.EqualProto(
							expectedUpdatedResp2.GetConfigNode().GetTrustScoreProfileConfig()),
					})),
				})))).
				Return(&configpb.CreateConfigNodeResponse{
					Id:         sampleID2,
					CreateTime: timestamppb.Now(),
					UpdateTime: timestamppb.Now(),
				}, nil)

			// Read in given order
			gomock.InOrder(
				mockConfigClient.EXPECT().
					ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
						"Id": Equal(sampleID),
					})))).
					Times(3).
					Return(expectedResp, nil),

				mockConfigClient.EXPECT().
					ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
						"Id": Equal(sampleID),
					})))).
					Times(3).
					Return(expectedUpdatedResp, nil),

				mockConfigClient.EXPECT().
					ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
						"Id": Equal(sampleID2),
					})))).
					Times(2).
					Return(expectedUpdatedResp2, nil),
			)

			// Delete
			mockConfigClient.EXPECT().
				DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(sampleID2),
				})))).
				Return(&configpb.DeleteConfigNodeResponse{}, nil)

			resource.Test(GinkgoT(), resource.TestCase{
				Providers: map[string]*schema.Provider{
					"indykite": provider,
				},
				Steps: []resource.TestStep{
					{
						// Checking Create and Read
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-trust-score-profile",
							`display_name = "Display name of TrustScoreProfile configuration"
							node_classification = "Person"
							dimensions {
								name   = "NAME_FRESHNESS"
								weight = 1.0
							}
							dimensions {
								name   = "NAME_ORIGIN"
								weight = 1.0
							}
							schedule = "UPDATE_FREQUENCY_DAILY"
							`,
						),
						Check: resource.ComposeTestCheckFunc(
							testResourceTrustScoreProfileExists(resourceName, expectedResp),
						),
					},
					{
						// Checking Update and Read
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-trust-score-profile",
							`display_name = "Display name of TrustScoreProfile configuration"
							node_classification = "Person"
							dimensions {
							name   = "NAME_COMPLETENESS"
							weight = 1.0
							}
							dimensions {
								name   = "NAME_ORIGIN"
								weight = 1.0
							}
							schedule = "UPDATE_FREQUENCY_SIX_HOURS"
							`),
						Check: resource.ComposeTestCheckFunc(
							testResourceTrustScoreProfileExists(resourceName, expectedUpdatedResp),
						),
					},
					{
						// Checking Recreate and Read
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-trust-score-profile",
							`display_name = "Display name of TrustScoreProfile configuration"
							node_classification = "Employee"
							dimensions {
							name   = "NAME_FRESHNESS"
							weight = 1.0
							}
							schedule = "UPDATE_FREQUENCY_SIX_HOURS"
							`),
						// ExpectError: regexp.MustCompile(
						// "InvalidArgument desc = update source or target node is not allowed"),
						Check: resource.ComposeTestCheckFunc(
							testResourceTrustScoreProfileExists(resourceName, expectedUpdatedResp2),
						),
					},
				},
			})
		})
	})
})

func testResourceTrustScoreProfileExists(
	n string,
	data *configpb.ReadConfigNodeResponse,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != data.GetConfigNode().GetId() {
			return errors.New("ID does not match")
		}
		attrs := rs.Primary.Attributes

		keys := Keys{
			"id": Equal(data.GetConfigNode().GetId()),
			"%":  Not(BeEmpty()),

			"location":     Equal(data.GetConfigNode().GetAppSpaceId()),
			"customer_id":  Equal(data.GetConfigNode().GetCustomerId()),
			"app_space_id": Equal(data.GetConfigNode().GetAppSpaceId()),
			"name":         Equal(data.GetConfigNode().GetName()),
			"display_name": Equal(data.GetConfigNode().GetDisplayName()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
			"node_classification": Equal(data.GetConfigNode().
				GetTrustScoreProfileConfig().
				GetNodeClassification()),
			"schedule": Equal(indykite.ReverseProtoEnumMap(
				indykite.TrustScoreProfileScheduleFrequencies,
			)[data.GetConfigNode().GetTrustScoreProfileConfig().GetSchedule()]),
		}

		dimensions := data.GetConfigNode().
			GetTrustScoreProfileConfig().
			GetDimensions()
		mapDimensions := make([]map[string]any, len(dimensions))
		for k, v := range dimensions {
			if v != nil {
				if mapDimensions[k] == nil {
					mapDimensions[k] = make(map[string]any)
				}
				mapDimensions[k]["name"] = indykite.ReverseProtoEnumMap(indykite.TrustScoreDimensionNames)[v.GetName()]
				mapDimensions[k]["weight"] = strconv.FormatFloat(float64(v.GetWeight()), 'f', -1, 64)
			}
		}
		addSliceMapMatcherToKeys(keys, "dimensions", mapDimensions, true)
		keys["dimensions.#"] = Equal(strconv.Itoa(len(dimensions)))
		keys["dimensions.0.%"] = Equal("2")

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
	}
}
