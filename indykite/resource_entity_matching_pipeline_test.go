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

var _ = Describe("Resource EntityMatchingPipeline", func() {
	const (
		resourceName = "indykite_entity_matching_pipeline.development"
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

		tfConfigDef = `resource "indykite_entity_matching_pipeline" "development" {
			location = "%s"
			name = "%s"
			%s
		}`

		validSettings = `
		source_node_filter = ["Person"]
		target_node_filter = ["Person"]
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
							`target_node_filter = ["Person"]
							`),
						ExpectError: regexp.MustCompile(
							`The argument "source_node_filter" is required, but no definition was found.`),
					},
				},
			})
		})
	})

	Describe("Valid configurations", func() {
		It("Test CRUD of EntityMatchingPipeline configuration", func() {
			expectedResp := &configpb.ReadConfigNodeResponse{
				ConfigNode: &configpb.ConfigNode{
					Id:          sampleID,
					Name:        "my-first-entity-matching-pipeline",
					DisplayName: "Display name of EntityMatchingPipeline configuration",
					CustomerId:  customerID,
					AppSpaceId:  appSpaceID,
					CreateTime:  timestamppb.Now(),
					UpdateTime:  timestamppb.Now(),
					Config: &configpb.ConfigNode_EntityMatchingPipelineConfig{
						EntityMatchingPipelineConfig: &configpb.EntityMatchingPipelineConfig{
							NodeFilter: &configpb.EntityMatchingPipelineConfig_NodeFilter{
								SourceNodeTypes: []string{"Person"},
								TargetNodeTypes: []string{"Person"},
							},
						},
					},
				},
			}
			expectedUpdatedResp := &configpb.ReadConfigNodeResponse{
				ConfigNode: &configpb.ConfigNode{
					Id:          sampleID,
					Name:        "my-first-entity-matching-pipeline",
					DisplayName: "Display name of EntityMatchingPipeline configuration",
					CustomerId:  customerID,
					AppSpaceId:  appSpaceID,
					CreateTime:  expectedResp.GetConfigNode().GetCreateTime(),
					UpdateTime:  timestamppb.Now(),
					Config: &configpb.ConfigNode_EntityMatchingPipelineConfig{
						EntityMatchingPipelineConfig: &configpb.EntityMatchingPipelineConfig{
							NodeFilter: &configpb.EntityMatchingPipelineConfig_NodeFilter{
								SourceNodeTypes: []string{"Person"},
								TargetNodeTypes: []string{"Person"},
							},
							RerunInterval: "1 day",
						},
					},
				},
			}

			expectedUpdatedResp2 := &configpb.ReadConfigNodeResponse{
				ConfigNode: &configpb.ConfigNode{
					Id:          sampleID2,
					Name:        "my-first-entity-matching-pipeline",
					DisplayName: "Display name of EntityMatchingPipeline configuration",
					CustomerId:  customerID,
					AppSpaceId:  appSpaceID,
					CreateTime:  expectedResp.GetConfigNode().GetCreateTime(),
					UpdateTime:  timestamppb.Now(),
					Config: &configpb.ConfigNode_EntityMatchingPipelineConfig{
						EntityMatchingPipelineConfig: &configpb.EntityMatchingPipelineConfig{
							NodeFilter: &configpb.EntityMatchingPipelineConfig_NodeFilter{
								SourceNodeTypes: []string{"Person"},
								TargetNodeTypes: []string{"Car"},
							},
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
						"EntityMatchingPipelineConfig": test.EqualProto(
							expectedResp.GetConfigNode().GetEntityMatchingPipelineConfig()),
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
						"EntityMatchingPipelineConfig": test.EqualProto(
							expectedUpdatedResp.GetConfigNode().GetEntityMatchingPipelineConfig()),
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
						"EntityMatchingPipelineConfig": test.EqualProto(
							expectedUpdatedResp2.GetConfigNode().GetEntityMatchingPipelineConfig()),
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
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-entity-matching-pipeline",
							`display_name = "Display name of EntityMatchingPipeline configuration"
							source_node_filter = ["Person"]
							target_node_filter = ["Person"]
							`,
						),
						Check: resource.ComposeTestCheckFunc(
							testResourceEntityMatchingPipelineExists(resourceName, expectedResp),
						),
					},
					{
						// Checking Update and Read
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-entity-matching-pipeline",
							`display_name = "Display name of EntityMatchingPipeline configuration"
							source_node_filter = ["Person"]
							target_node_filter = ["Person"]
							rerun_interval = "1 day"
							`),
						Check: resource.ComposeTestCheckFunc(
							testResourceEntityMatchingPipelineExists(resourceName, expectedUpdatedResp),
						),
					},
					{
						// Checking Recreate and Read
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-entity-matching-pipeline",
							`display_name = "Display name of EntityMatchingPipeline configuration"
							source_node_filter = ["Person"]
							target_node_filter = ["Car"]
							`),
						// ExpectError: regexp.MustCompile(
						// "InvalidArgument desc = update source or target node is not allowed"),
						Check: resource.ComposeTestCheckFunc(
							testResourceEntityMatchingPipelineExists(resourceName, expectedUpdatedResp2),
						),
					},
				},
			})
		})
	})
})

func testResourceEntityMatchingPipelineExists(
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
		}

		sourceNodeFilter := data.GetConfigNode().
			GetEntityMatchingPipelineConfig().
			GetNodeFilter().
			GetSourceNodeTypes()
		keys["source_node_filter.#"] = Equal(strconv.Itoa(len(sourceNodeFilter)))
		addStringArrayToKeys(keys, "source_node_filter", sourceNodeFilter)

		targetNodeFilter := data.GetConfigNode().
			GetEntityMatchingPipelineConfig().
			GetNodeFilter().
			GetTargetNodeTypes()
		keys["target_node_filter.#"] = Equal(strconv.Itoa(len(targetNodeFilter)))
		addStringArrayToKeys(keys, "target_node_filter", targetNodeFilter)

		rerunInterval := data.GetConfigNode().
			GetEntityMatchingPipelineConfig().
			GetRerunInterval()
		keys["rerun_interval"] = Equal(rerunInterval)

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
	}
}
