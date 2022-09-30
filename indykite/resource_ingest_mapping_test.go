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

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/jarvis-sdk-go/config"
	configpb "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
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

var _ = Describe("Resource Ingest Mapping config", func() {
	const resourceName = "indykite_ingest_mapping.wonka"
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
		ingestMappingConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				TenantId:    tenantID,
				Id:          sampleID,
				Name:        "wonka-ingest-mapping-config",
				DisplayName: "Wonka Ingesting chocolate",
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_IngestMappingConfig{
					IngestMappingConfig: &configpb.IngestMappingConfig{
						IngestType: &configpb.IngestMappingConfig_Upsert{
							Upsert: &configpb.IngestMappingConfig_UpsertData{
								Entities: []*configpb.IngestMappingConfig_Entity{{
									Labels: []string{"Flavour"},
									ExternalId: &configpb.IngestMappingConfig_Property{
										SourceName: "smak",
										MappedName: "flavour",
									},
								}},
							},
						},
					},
				},
			},
		}

		ingestMappingInvalidResponse := proto.Clone(ingestMappingConfigResp).(*configpb.ReadConfigNodeResponse)
		ingestMappingInvalidResponse.ConfigNode.Config = &configpb.ConfigNode_AuthFlowConfig{}

		ingestMappingConfigUpdateResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				TenantId:    tenantID,
				Id:          sampleID,
				Name:        "wonka-ingest-mapping-config",
				Description: wrapperspb.String("Description of the best Ingestion config by Wonka inc."),
				CreateTime:  ingestMappingConfigResp.ConfigNode.CreateTime,
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_IngestMappingConfig{
					IngestMappingConfig: &configpb.IngestMappingConfig{
						IngestType: &configpb.IngestMappingConfig_Upsert{
							Upsert: &configpb.IngestMappingConfig_UpsertData{
								Entities: []*configpb.IngestMappingConfig_Entity{{
									Labels:   []string{"Flavour"},
									TenantId: tenantID,
									ExternalId: &configpb.IngestMappingConfig_Property{
										SourceName: "smak",
										MappedName: "flavour",
									},
									Properties: []*configpb.IngestMappingConfig_Property{{
										SourceName: "barva",
										MappedName: "color",
										IsRequired: true,
									}},
									Relationships: []*configpb.IngestMappingConfig_Relationship{{
										ExternalId: "abc",
										Type:       "MANUFACTURED_AT",
										Direction:  configpb.IngestMappingConfig_DIRECTION_OUTBOUND,
										MatchLabel: "Company",
									}},
								}},
							},
						},
					},
				},
			},
		}

		// Create
		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(ingestMappingConfigResp.ConfigNode.Name),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					ingestMappingConfigResp.ConfigNode.DisplayName,
				)})),
				"Description": BeNil(),
				"Location":    Equal(tenantID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"IngestMappingConfig": test.EqualProto(ingestMappingConfigResp.ConfigNode.GetIngestMappingConfig()),
				})),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         ingestMappingConfigResp.ConfigNode.Id,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
			}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(ingestMappingConfigResp.ConfigNode.Id),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					ingestMappingConfigUpdateResp.ConfigNode.Description.GetValue(),
				)})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"IngestMappingConfig": test.EqualProto(
						ingestMappingConfigUpdateResp.ConfigNode.GetIngestMappingConfig(),
					),
				})),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{Id: ingestMappingConfigResp.ConfigNode.Id}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(ingestMappingConfigResp.ConfigNode.Id),
				})))).
				Times(3).
				Return(ingestMappingConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(ingestMappingConfigResp.ConfigNode.Id),
				})))).
				Return(ingestMappingInvalidResponse, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(ingestMappingConfigResp.ConfigNode.Id),
				})))).
				Return(ingestMappingConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(ingestMappingConfigResp.ConfigNode.Id),
				})))).
				Times(2).
				Return(ingestMappingConfigUpdateResp, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(ingestMappingConfigResp.ConfigNode.Id),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
			},
			Steps: []resource.TestStep{
				// Error cases should be always first, easier to avoid missing mocks or incomplete plan
				{
					Config: `resource "indykite_ingest_mapping" "wonka" {
						name = "wonka-ingest-mapping-config"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "location" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "` + customerID + `"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "name" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-ingest-mapping-config"

						json_config = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "customer_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "` + customerID + `"
						app_space_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-ingest-mapping-config"

						json_config = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "app_space_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "` + customerID + `"
						tenant_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-ingest-mapping-config"

						json_config = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "tenant_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "something-invalid"
						name = "wonka-ingest-mapping-config"

						json_config = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`expected to have 'gid:' prefix`),
				},
				{
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "` + customerID + `"
						name = "Invalid Name @#$"

						json_config = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`Value can have lowercase letters, digits, or hyphens.`),
				},
				{
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "` + customerID + `"
						name = "wonka-ingest-mapping-config"
					}
					`,
					ExpectError: regexp.MustCompile(`argument "json_config" is required`),
				},
				{
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "` + customerID + `"
						name = "wonka-ingest-mapping-config"

						json_config = "not valid json"
					}
					`,
					ExpectError: regexp.MustCompile(`"json_config" contains an invalid JSON`),
				},
				{
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "` + customerID + `"
						name = "wonka-ingest-mapping-config"

						json_config = "[]"
					}
					`,
					ExpectError: regexp.MustCompile(
						`"json_config" cannot be unmarshalled into Proto message: .* unexpected token \[`,
					),
				},
				{
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "` + customerID + `"
						name = "wonka-ingest-mapping-config"

						json_config = "{\"upsert\": {\"entities\": []}}"
					}
					`,
					ExpectError: regexp.MustCompile(`"json_config" has invalid Upsert.Entities: ` +
						`value must contain between 1 and 10 items, inclusive`),
				},

				// ---- Run mocked tests here ----
				{
					// Minimal config - Checking Create and Read (ingestMappingConfigResp)
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "` + tenantID + `"
						name = "wonka-ingest-mapping-config"
						display_name = "Wonka Ingesting chocolate"

						json_config = jsonencode({"upsert": {"entities": [{
							"tenantId": "",
							"labels":["Flavour"],
							"externalId": {"sourceName": "smak", mappedName: "flavour", "isRequired": false},
							"properties": [],
							"relationships": []
						}]}})
					}
					`,
					Check: resource.ComposeTestCheckFunc(testIngestMappingResourceDataExists(
						resourceName,
						ingestMappingConfigResp,
					)),
				},
				{
					// Performs 1 read (ingestMappingConfigResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: ingestMappingConfigResp.ConfigNode.Id,
				},
				{
					// Performs 1 read (ingestMappingInvalidResponse)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: ingestMappingConfigResp.ConfigNode.Id,
					ExpectError: regexp.MustCompile(
						`response is not valid IngestMappingConfig((?s).*)IndyKite plugin error, please report this`),
				},
				{
					// Checking Read(ingestMappingConfigResp), Update and Read(ingestMappingConfigUpdateResp)
					Config: `resource "indykite_ingest_mapping" "wonka" {
						location = "` + tenantID + `"
						name = "wonka-ingest-mapping-config"
						description = "Description of the best Ingestion config by Wonka inc."

						json_config = jsonencode({"upsert": {"entities": [{
							"tenantId": "` + tenantID + `",
							"labels":["Flavour"],
							"externalId": {"sourceName": "smak", mappedName: "flavour", "isRequired": false},
							"properties": [{"sourceName": "barva", mappedName: "color", "isRequired": true}],
							"relationships": [{
								"externalId":"abc",
								"type": "MANUFACTURED_AT",
								"direction": "DIRECTION_OUTBOUND",
								"matchLabel": "Company"
							}]
						}]}})
					}
					`,
					Check: resource.ComposeTestCheckFunc(testIngestMappingResourceDataExists(
						resourceName,
						ingestMappingConfigUpdateResp,
					)),
				},
			},
		})
	})
})

func testIngestMappingResourceDataExists(
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

		expectedJSON, err := protojson.MarshalOptions{EmitUnpopulated: true}.
			Marshal(data.ConfigNode.GetIngestMappingConfig())
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
