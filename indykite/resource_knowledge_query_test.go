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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource Knowledge Query config", func() {
	const resourceName = "indykite_knowledge_query.wonka"
	var (
		mockCtrl         *gomock.Controller
		mockConfigClient *configm.MockConfigManagementAPIClient
		provider         *schema.Provider

		authorizationPolicyID = "gid:AALikeGIDOfAuthZPolicyAA"
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
	})

	It("Test error cases", func() {
		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Error cases should be always first, easier to avoid missing mocks or incomplete plan
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						name = "wonka-knowledge-query-config"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "location" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "name" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-knowledge-query-config"

						query = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "status" is required, but no definition was found.`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-knowledge-query-config"

						query = "{}"
						status = "non-existing"
					}
					`,
					ExpectError: regexp.MustCompile(
						`The argument "policy_id" is required, but no definition was found.`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						name = "wonka-knowledge-query-config"

						query = "{}"
						status = "non-existing"
						policy_id = "gid:AALikeGIDOfAuthZPolicyAA"
					}
					`,
					ExpectError: regexp.MustCompile(
						`expected status to be one of \["active" "draft" "inactive"\], got non-existing`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						name = "wonka-knowledge-query-config"

						query = "{}"
						status = "inactive"
						policy_id = "abc"
					}
					`,
					ExpectError: regexp.MustCompile(`expected to have 'gid:' prefix`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						app_space_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-knowledge-query-config"

						query = "{}"
						status = "non-existing"
						policy_id = "gid:AALikeGIDOfAuthZPolicyAA"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "app_space_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						name = "wonka-knowledge-query-config"

						status = "active"
						policy_id = "gid:AALikeGIDOfAuthZPolicyAA"
					}
					`,
					ExpectError: regexp.MustCompile(`argument "query" is required`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						name = "wonka-knowledge-query-config"

						query = "not valid query"
						status = "active"
						policy_id = "gid:AALikeGIDOfAuthZPolicyAA"
					}
					`,
					ExpectError: regexp.MustCompile(`"query" contains an invalid JSON`),
				},
			},
		})
	})

	It("Test all CRUD", func() {
		knowledgeQueryConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				Id:          sampleID,
				Name:        "wonka-knowledge-query-config",
				DisplayName: "Wonka Query for chocolate receipts",
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_KnowledgeQueryConfig{
					KnowledgeQueryConfig: &configpb.KnowledgeQueryConfig{
						Query:    `{"something":["like","query"]}`,
						Status:   configpb.KnowledgeQueryConfig_STATUS_ACTIVE,
						PolicyId: authorizationPolicyID,
					},
				},
			},
		}

		knowledgeQueryInvalidResponse := proto.Clone(knowledgeQueryConfigResp).(*configpb.ReadConfigNodeResponse)
		knowledgeQueryInvalidResponse.ConfigNode.Config = &configpb.ConfigNode_AuditSinkConfig{}

		knowledgeQueryConfigUpdateResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				Id:          sampleID,
				Name:        "wonka-knowledge-query-config",
				Description: wrapperspb.String("Description of the best Knowledge Query by Wonka inc."),
				CreateTime:  knowledgeQueryConfigResp.GetConfigNode().GetCreateTime(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_KnowledgeQueryConfig{
					KnowledgeQueryConfig: &configpb.KnowledgeQueryConfig{
						Query:    `{"something":["like","another","query"]}`,
						Status:   configpb.KnowledgeQueryConfig_STATUS_DRAFT,
						PolicyId: authorizationPolicyID,
					},
				},
			},
		}

		// Create
		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(knowledgeQueryConfigResp.GetConfigNode().GetName()),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					knowledgeQueryConfigResp.GetConfigNode().GetDisplayName(),
				)})),
				"Description": BeNil(),
				"Location":    Equal(appSpaceID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{"KnowledgeQueryConfig": test.EqualProto(
					knowledgeQueryConfigResp.GetConfigNode().GetKnowledgeQueryConfig(),
				)})),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         knowledgeQueryConfigResp.GetConfigNode().GetId(),
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
			}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(knowledgeQueryConfigResp.GetConfigNode().GetId()),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					knowledgeQueryConfigUpdateResp.GetConfigNode().GetDescription().GetValue(),
				)})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"KnowledgeQueryConfig": test.EqualProto(
						knowledgeQueryConfigUpdateResp.GetConfigNode().GetKnowledgeQueryConfig(),
					),
				})),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{
				Id: knowledgeQueryConfigResp.GetConfigNode().GetId(),
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(knowledgeQueryConfigResp.GetConfigNode().GetId()),
				})))).
				Times(3).
				Return(knowledgeQueryConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(knowledgeQueryConfigResp.GetConfigNode().GetId()),
				})))).
				Return(knowledgeQueryInvalidResponse, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(knowledgeQueryConfigResp.GetConfigNode().GetId()),
				})))).
				Return(knowledgeQueryConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(knowledgeQueryConfigResp.GetConfigNode().GetId()),
				})))).
				Times(2).
				Return(knowledgeQueryConfigUpdateResp, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(knowledgeQueryConfigResp.GetConfigNode().GetId()),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Minimal config - Checking Create and Read (knowledgeQueryConfigResp)
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-knowledge-query-config"
						display_name = "Wonka Query for chocolate receipts"

						query = jsonencode({"something":["like", "query"]})
						status = "active"
						policy_id = "` + authorizationPolicyID + `"
					}`,

					Check: resource.ComposeTestCheckFunc(testKnowledgeQueryResourceDataExists(
						resourceName,
						knowledgeQueryConfigResp,
						nil,
					)),
				},
				{
					// Performs 1 read (knowledgeQueryConfigResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: knowledgeQueryConfigResp.GetConfigNode().GetId(),
				},
				{
					// Performs 1 read (knowledgeQueryInvalidResponse)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: knowledgeQueryConfigResp.GetConfigNode().GetId(),
					ExpectError: regexp.MustCompile(
						`not valid KnowledgeQueryConfig((?s).*)IndyKite plugin error, please report this issue`),
				},
				// Checking Read(knowledgeQueryConfigResp), Update and Read(knowledgeQueryConfigUpdateResp)
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-knowledge-query-config"
						description = "Description of the best Knowledge Query by Wonka inc."

						query = jsonencode({"something":["like", "another", "query"]})
						status = "draft"
						policy_id = "` + authorizationPolicyID + `"
					}
					`,
					Check: resource.ComposeTestCheckFunc(testKnowledgeQueryResourceDataExists(
						resourceName,
						knowledgeQueryConfigUpdateResp,
						nil,
					)),
				},
			},
		})
	})
})

func testKnowledgeQueryResourceDataExists(
	n string,
	data *configpb.ReadConfigNodeResponse,
	extraKeys Keys,
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
			"%":  Not(BeEmpty()), // This is Terraform helper

			"location":     Not(BeEmpty()), // Response does not return this
			"customer_id":  Equal(data.GetConfigNode().GetCustomerId()),
			"app_space_id": Equal(data.GetConfigNode().GetAppSpaceId()),
			"name":         Equal(data.GetConfigNode().GetName()),
			"display_name": Equal(data.GetConfigNode().GetDisplayName()),
			"description":  Equal(data.GetConfigNode().GetDescription().GetValue()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
			"query":        MatchJSON(data.GetConfigNode().GetKnowledgeQueryConfig().GetQuery()),
			"status": Equal(indykite.ReverseProtoEnumMap(
				indykite.KnowledgeQueryStatusTypes,
			)[data.GetConfigNode().GetKnowledgeQueryConfig().GetStatus()]),
			"policy_id": Equal(data.GetConfigNode().GetKnowledgeQueryConfig().GetPolicyId()),
		}

		for k, v := range extraKeys {
			keys[k] = v
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), attrs)
	}
}
