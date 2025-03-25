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

var _ = Describe("Resource Authorization Policy config", func() {
	const resourceName = "indykite_authorization_policy.wonka"
	var (
		mockCtrl         *gomock.Controller
		mockConfigClient *configm.MockConfigManagementAPIClient
		provider         *schema.Provider
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

	It("Test all CRUD", func() {
		authzPolicyConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				Id:          sampleID,
				Name:        "wonka-authorization-policy-config",
				DisplayName: "Wonka Authorization for chocolate receipts",
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_AuthorizationPolicyConfig{
					AuthorizationPolicyConfig: &configpb.AuthorizationPolicyConfig{
						//nolint:lll
						Policy: "{\"meta\":{\"policyVersion\":\"1.0-indykite\"},\"subject\":{\"type\":\"Person\"},\"actions\":[\"CAN_DRIVE\",\"CAN_PERFORM_SERVICE\"],\"resource\":{\"type\":\"Car\"},\"condition\":{\"cypher\":\"MATCH (subject:Person)-[:PART_OF]->(:Household)-[:DISPOSES]->(resource:Car)\"}}",
						Status: configpb.AuthorizationPolicyConfig_STATUS_ACTIVE,
						Tags:   nil,
					},
				},
			},
		}

		authzPolicyInvalidResponse := proto.Clone(authzPolicyConfigResp).(*configpb.ReadConfigNodeResponse)
		authzPolicyInvalidResponse.ConfigNode.Config = &configpb.ConfigNode_EventSinkConfig{}

		authzPolicyConfigUpdateResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				Id:          sampleID,
				Name:        "wonka-authorization-policy-config",
				Description: wrapperspb.String("Description of the best Authz Policies by Wonka inc."),
				CreateTime:  authzPolicyConfigResp.GetConfigNode().GetCreateTime(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_AuthorizationPolicyConfig{
					AuthorizationPolicyConfig: &configpb.AuthorizationPolicyConfig{
						//nolint:lll
						Policy: "{\"meta\":{\"policyVersion\":\"1.0-indykite\"},\"subject\":{\"type\":\"Person\"},\"actions\":[\"CAN_DRIVE\",\"CAN_PERFORM_SERVICE\"],\"resource\":{\"type\":\"Car\"},\"condition\":{\"cypher\":\"MATCH (subject:Person)-[:PART_OF]->(:Household)-[:DISPOSES]->(resource:Car)\"}}",
						Status: configpb.AuthorizationPolicyConfig_STATUS_ACTIVE,
						Tags:   []string{"test", "wonka"},
					},
				},
			},
		}

		// Create
		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(authzPolicyConfigResp.GetConfigNode().GetName()),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					authzPolicyConfigResp.GetConfigNode().GetDisplayName(),
				)})),
				"Description": BeNil(),
				"Location":    Equal(appSpaceID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{"AuthorizationPolicyConfig": test.EqualProto(
					authzPolicyConfigResp.GetConfigNode().GetAuthorizationPolicyConfig(),
				)})),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         authzPolicyConfigResp.GetConfigNode().GetId(),
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
			}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(authzPolicyConfigResp.GetConfigNode().GetId()),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					authzPolicyConfigUpdateResp.GetConfigNode().GetDescription().GetValue(),
				)})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"AuthorizationPolicyConfig": test.EqualProto(
						authzPolicyConfigUpdateResp.GetConfigNode().GetAuthorizationPolicyConfig(),
					),
				})),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{
				Id: authzPolicyConfigResp.GetConfigNode().GetId(),
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(authzPolicyConfigResp.GetConfigNode().GetId()),
				})))).
				Times(3).
				Return(authzPolicyConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(authzPolicyConfigResp.GetConfigNode().GetId()),
				})))).
				Return(authzPolicyInvalidResponse, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(authzPolicyConfigResp.GetConfigNode().GetId()),
				})))).
				Return(authzPolicyConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(authzPolicyConfigResp.GetConfigNode().GetId()),
				})))).
				Times(2).
				Return(authzPolicyConfigUpdateResp, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(authzPolicyConfigResp.GetConfigNode().GetId()),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Error cases should be always first, easier to avoid missing mocks or incomplete plan
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						name = "wonka-authorization-policy-config"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "location" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "name" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-authorization-policy-config"

						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`The argument "status" is required, but no definition was found.`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-authorization-policy-config"
						status = "active"

						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "customer_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						app_space_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-authorization-policy-config"
						status = "active"

						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "app_space_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "something-invalid"
						name = "wonka-authorization-policy-config"
						status = "active"

						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`expected to have 'gid:' prefix`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						name = "Invalid Name @#$"
						status = "active"

						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`Value can have lowercase letters, digits, or hyphens.`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						name = "wonka-authorization-policy-config"
						status = "active"
					}
					`,
					ExpectError: regexp.MustCompile(`argument "json" is required`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						name = "wonka-authorization-policy-config"
						status = "active"

						json = "not valid json"
					}
					`,
					ExpectError: regexp.MustCompile(`"json" contains an invalid JSON`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = ""
						name = "wonka-authorization-policy-config"
						status = "active"
						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`expected to have 'gid:' prefix`),
				},

				// ---- Run mocked tests here ----
				// Minimal config - Checking Create and Read (authzPolicyConfigResp)
				{
					//nolint:lll
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-authorization-policy-config"
						display_name = "Wonka Authorization for chocolate receipts"
						status = "active"

						json = "{\"meta\":{\"policyVersion\":\"1.0-indykite\"},\"subject\":{\"type\":\"Person\"},\"actions\":[\"CAN_DRIVE\",\"CAN_PERFORM_SERVICE\"],\"resource\":{\"type\":\"Car\"},\"condition\":{\"cypher\":\"MATCH (subject:Person)-[:PART_OF]->(:Household)-[:DISPOSES]->(resource:Car)\"}}"
					}`,

					Check: resource.ComposeTestCheckFunc(testAuthorizationPolicyResourceDataExists(
						resourceName,
						authzPolicyConfigResp,
						nil,
					)),
				},
				{
					// Performs 1 read (authzPolicyConfigResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: authzPolicyConfigResp.GetConfigNode().GetId(),
				},
				{
					// Performs 1 read (authzPolicyInvalidResponse)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: authzPolicyConfigResp.GetConfigNode().GetId(),
					ExpectError: regexp.MustCompile(
						`not valid AuthorizationPolicyConfig((?s).*)IndyKite plugin error, please report this issue`),
				},
				// Checking Read(authzPolicyConfigResp), Update and Read(authzPolicyConfigUpdateResp)
				{
					//nolint:lll
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-authorization-policy-config"
						description = "Description of the best Authz Policies by Wonka inc."
						status = "active"
						tags = ["test", "wonka"]

						json = "{\"meta\":{\"policyVersion\":\"1.0-indykite\"},\"subject\":{\"type\":\"Person\"},\"actions\":[\"CAN_DRIVE\",\"CAN_PERFORM_SERVICE\"],\"resource\":{\"type\":\"Car\"},\"condition\":{\"cypher\":\"MATCH (subject:Person)-[:PART_OF]->(:Household)-[:DISPOSES]->(resource:Car)\"}}"
					}
					`,
					Check: resource.ComposeTestCheckFunc(testAuthorizationPolicyResourceDataExists(
						resourceName,
						authzPolicyConfigUpdateResp,
						addStringArrayToKeys(
							Keys{},
							"tags",
							authzPolicyConfigUpdateResp.GetConfigNode().GetAuthorizationPolicyConfig().GetTags(),
						),
					)),
				},
			},
		})
	})
})

func testAuthorizationPolicyResourceDataExists(
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

		expectedJSON := data.GetConfigNode().GetAuthorizationPolicyConfig().GetPolicy()

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
			"json":         MatchJSON(expectedJSON),
			"status": Equal(indykite.ReverseProtoEnumMap(
				indykite.AuthorizationPolicyStatusTypes,
			)[data.GetConfigNode().GetAuthorizationPolicyConfig().GetStatus()]),
		}

		for k, v := range extraKeys {
			keys[k] = v
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), attrs)
	}
}
