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
	identityv1beta1 "github.com/indykite/jarvis-sdk-go/gen/indykite/identity/v1beta1"
	knowledge_graphpb "github.com/indykite/jarvis-sdk-go/gen/indykite/knowledge_graph/v1beta1"
	configm "github.com/indykite/jarvis-sdk-go/test/config/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"
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

	It("Test all CRUD", func() {
		authzPolicyConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				TenantId:    tenantID,
				Id:          sampleID,
				Name:        "wonka-authorization-policy-config",
				DisplayName: "Wonka Authorization for chocolate receipts",
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_AuthorizationPolicyConfig{
					AuthorizationPolicyConfig: &configpb.AuthorizationPolicyConfig{
						Policy: &knowledge_graphpb.Policy{
							Path: &knowledge_graphpb.Path{
								SubjectId:  "sub",
								ResourceId: "res",
								Entities: []*knowledge_graphpb.Path_Entity{
									{Id: "sub", Labels: []string{"DigitalTwin"}},
									{Id: "res", Labels: []string{"Company"}},
								},
								Relationships: []*knowledge_graphpb.Path_Relationship{{
									Source: "sub",
									Target: "res",
									Types:  []string{"WORKS_AT"},
								}},
							},
							Actions: []string{"READ", "LIST"},
						},
					},
				},
			},
		}

		authzPolicyInvalidResponse := proto.Clone(authzPolicyConfigResp).(*configpb.ReadConfigNodeResponse)
		authzPolicyInvalidResponse.ConfigNode.Config = &configpb.ConfigNode_AuthFlowConfig{}

		authzPolicyConfigUpdateResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				TenantId:    tenantID,
				Id:          sampleID,
				Name:        "wonka-authorization-policy-config",
				Description: wrapperspb.String("Description of the best Authz Policies by Wonka inc."),
				CreateTime:  authzPolicyConfigResp.ConfigNode.CreateTime,
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_AuthorizationPolicyConfig{
					AuthorizationPolicyConfig: &configpb.AuthorizationPolicyConfig{
						Policy: &knowledge_graphpb.Policy{
							Path: &knowledge_graphpb.Path{
								SubjectId:  "sub",
								ResourceId: "res",
								Entities: []*knowledge_graphpb.Path_Entity{
									{
										Id:     "sub",
										Labels: []string{"DigitalTwin"},
										IdentityProperties: []*knowledge_graphpb.Path_Entity_IdentityProperty{{
											Property:              "email",
											Value:                 "wonka@indykite.com",
											MinimumAssuranceLevel: identityv1beta1.AssuranceLevel_ASSURANCE_LEVEL_LOW,
											AllowedIssuers:        []string{"google.com"},
											MustBePrimary:         true,
											AllowedVerifiers:      []string{"google.com"},
										}},
									},
									{Id: "res", Labels: []string{"Company"}},
									{
										Id:     "group",
										Labels: []string{"Group"},
										KnowledgeProperties: []*knowledge_graphpb.Path_Entity_KnowledgeProperty{{
											Property: "name",
											Value:    "chocolate-factory",
										}},
									},
								},
								Relationships: []*knowledge_graphpb.Path_Relationship{
									{Source: "sub", Target: "res", Types: []string{"WORKS_AT"}},
									{Source: "group", Target: "res", Types: []string{"OWNS"}, NonDirectional: true},
								},
							},
							Active:  true,
							Actions: []string{"READ", "WRITE", "UPDATE"},
						},
					},
				},
			},
		}

		// Create
		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(authzPolicyConfigResp.ConfigNode.Name),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					authzPolicyConfigResp.ConfigNode.DisplayName,
				)})),
				"Description": BeNil(),
				"Location":    Equal(tenantID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{"AuthorizationPolicyConfig": test.EqualProto(
					authzPolicyConfigResp.ConfigNode.GetAuthorizationPolicyConfig(),
				)})),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         authzPolicyConfigResp.ConfigNode.Id,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
			}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(authzPolicyConfigResp.ConfigNode.Id),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					authzPolicyConfigUpdateResp.ConfigNode.Description.GetValue(),
				)})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"AuthorizationPolicyConfig": test.EqualProto(
						authzPolicyConfigUpdateResp.ConfigNode.GetAuthorizationPolicyConfig(),
					),
				})),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{Id: authzPolicyConfigResp.ConfigNode.Id}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(authzPolicyConfigResp.ConfigNode.Id),
				})))).
				Times(3).
				Return(authzPolicyConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(authzPolicyConfigResp.ConfigNode.Id),
				})))).
				Return(authzPolicyInvalidResponse, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(authzPolicyConfigResp.ConfigNode.Id),
				})))).
				Return(authzPolicyConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(authzPolicyConfigResp.ConfigNode.Id),
				})))).
				Times(2).
				Return(authzPolicyConfigUpdateResp, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(authzPolicyConfigResp.ConfigNode.Id),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
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

						json_config = "{}"
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

						json_config = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "app_space_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						tenant_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-authorization-policy-config"

						json_config = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "tenant_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "something-invalid"
						name = "wonka-authorization-policy-config"

						json_config = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`expected to have 'gid:' prefix`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						name = "Invalid Name @#$"

						json_config = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`Value can have lowercase letters, digits, or hyphens.`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						name = "wonka-authorization-policy-config"
					}
					`,
					ExpectError: regexp.MustCompile(`argument "json_config" is required`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						name = "wonka-authorization-policy-config"

						json_config = "not valid json"
					}
					`,
					ExpectError: regexp.MustCompile(`"json_config" contains an invalid JSON`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						name = "wonka-authorization-policy-config"

						json_config = "[]"
					}
					`,
					ExpectError: regexp.MustCompile(
						`"json_config" cannot be unmarshalled into Proto message: .* unexpected token \[`,
					),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						name = "wonka-authorization-policy-config"

						json_config = "{\"policy\": {\"path\": {}}}"
					}
					`,
					ExpectError: regexp.MustCompile(`"json_config" has invalid Policy.Path.SubjectId: ` +
						`value length must be between 2 and 50 runes, inclusive`),
				},

				// ---- Run mocked tests here ----
				{
					// Minimal config - Checking Create and Read (authzPolicyConfigResp)
					// nolint:lll
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + tenantID + `"
						name = "wonka-authorization-policy-config"
						display_name = "Wonka Authorization for chocolate receipts"

						json_config = jsonencode({ "policy": {
							"active": false,
							"path": {
								"subjectId": "sub",
								"resourceId": "res",
								"entities": [
									{"id":"sub","labels":["DigitalTwin"],"identityProperties":[],"knowledgeProperties":[]},
									{"id": "res", "labels": ["Company"],"identityProperties":[],"knowledgeProperties":[]}
								],
								"relationships": [
									{"source": "sub", "target": "res", "types": ["WORKS_AT"], "nonDirectional" = false}
								]
							},
							"actions": ["READ", "LIST"]
						}})
					}
					`,
					Check: resource.ComposeTestCheckFunc(testAuthorizationPolicyResourceDataExists(
						resourceName,
						authzPolicyConfigResp,
					)),
				},
				{
					// Performs 1 read (authzPolicyConfigResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: authzPolicyConfigResp.ConfigNode.Id,
				},
				{
					// Performs 1 read (authzPolicyInvalidResponse)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: authzPolicyConfigResp.ConfigNode.Id,
					ExpectError: regexp.MustCompile(
						`not valid AuthorizationPolicyConfig((?s).*)IndyKite plugin error, please report this issue`),
				},
				{
					// Checking Read(authzPolicyConfigResp), Update and Read(authzPolicyConfigUpdateResp)
					// nolint:lll
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + tenantID + `"
						name = "wonka-authorization-policy-config"
						description = "Description of the best Authz Policies by Wonka inc."

						json_config = jsonencode({ "policy": {
							"active": true,
							"path": {
								"subjectId": "sub",
								"resourceId": "res",
								"entities": [
									{
										"id":"sub",
										"labels":["DigitalTwin"],
										"identityProperties":[{
											"property": "email",
											"value": "wonka@indykite.com",
											"minimumAssuranceLevel": "ASSURANCE_LEVEL_LOW",
											"allowedIssuers": ["google.com"],
											"allowedVerifiers": ["google.com"],
											"mustBePrimary": true,
											"verificationTime": null
										}],
										knowledgeProperties:[]
									},
									{"id": "res", "labels": ["Company"],"identityProperties":[],"knowledgeProperties":[]},
									{
										"id": "group",
										"labels": ["Group"],
										"identityProperties":[],
										"knowledgeProperties":[{"property": "name", "value": "chocolate-factory"}]
									}
								],
								"relationships": [
									{"source": "sub", "target": "res", "types": ["WORKS_AT"], "nonDirectional" = false},
									{"source": "group", "target": "res", "types": ["OWNS"], "nonDirectional" = true}
								]
							},
							"actions": ["READ", "WRITE", "UPDATE"]
						}})
					}
					`,
					Check: resource.ComposeTestCheckFunc(testAuthorizationPolicyResourceDataExists(
						resourceName,
						authzPolicyConfigUpdateResp,
					)),
				},
			},
		})
	})
})

func testAuthorizationPolicyResourceDataExists(
	n string,
	data *configpb.ReadConfigNodeResponse,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != data.ConfigNode.Id {
			return fmt.Errorf("ID does not match")
		}
		attrs := rs.Primary.Attributes

		expectedJSON, err := protojson.MarshalOptions{EmitUnpopulated: true}.
			Marshal(data.ConfigNode.GetAuthorizationPolicyConfig())
		if err != nil {
			return err
		}

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

			"json_config": MatchJSON(expectedJSON),
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), attrs)
	}
}
