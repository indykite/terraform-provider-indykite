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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"
	"github.com/pborman/uuid"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource IngestPipeline", func() {
	//nolint:gosec,lll // there are no secrets
	const (
		resourceName  = "indykite_ingest_pipeline.development"
		appAgentToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaWQ6QUFBQUJXbHVaSGxyYVVSbGdBQUZEd0FBQUFBIiwic3ViIjoiZ2lkOkFBQUFCV2x1WkhscmFVUmxnQUFGRHdBQUFBQSIsImV4cCI6MjUzNDAyMjYxMTk5LCJpYXQiOjE1MTYyMzkwMjJ9.39Kc7pL8Vjf1S4qA6NHBGMP06TahR5Y9JOGSWKOo5Rw"
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
		mockedBookmark = "for-ingest-pipeline-cfg" + uuid.NewRandom().String()
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

	It("Test CRUD of Ingest Pipeline configuration", func() {
		tfConfigDef :=
			`resource "indykite_ingest_pipeline" "development" {
				location = "%s"
				name = "%s"
				%s
			}`
		expectedResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				Id:          sampleID,
				Name:        "my-first-ingest-pipeline",
				DisplayName: "Display name of Ingest Pipeline configuration",
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_IngestPipelineConfig{
					IngestPipelineConfig: &configpb.IngestPipelineConfig{
						Sources: []string{"source1", "source2"},
						Operations: []configpb.IngestPipelineOperation{
							configpb.IngestPipelineOperation_INGEST_PIPELINE_OPERATION_UPSERT_NODE,
							configpb.IngestPipelineOperation_INGEST_PIPELINE_OPERATION_UPSERT_RELATIONSHIP,
						},
						AppAgentToken: appAgentToken,
					},
				},
			},
		}
		expectedUpdatedResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				Id:          sampleID,
				Name:        "my-first-ingest-pipeline",
				Description: wrapperspb.String("ingest pipeline description"),
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				CreateTime:  expectedResp.ConfigNode.CreateTime,
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_IngestPipelineConfig{
					IngestPipelineConfig: &configpb.IngestPipelineConfig{
						Sources: []string{"source1", "source2", "source3"},
						Operations: []configpb.IngestPipelineOperation{
							configpb.IngestPipelineOperation_INGEST_PIPELINE_OPERATION_UPSERT_NODE,
							configpb.IngestPipelineOperation_INGEST_PIPELINE_OPERATION_UPSERT_RELATIONSHIP,
							configpb.IngestPipelineOperation_INGEST_PIPELINE_OPERATION_DELETE_NODE,
						},
						AppAgentToken: appAgentToken,
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
					"IngestPipelineConfig": test.EqualProto(
						expectedResp.GetConfigNode().GetIngestPipelineConfig(),
					),
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
					expectedUpdatedResp.GetConfigNode().GetDescription().GetValue(),
				)})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"IngestPipelineConfig": test.EqualProto(
						expectedUpdatedResp.GetConfigNode().GetIngestPipelineConfig(),
					),
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

					"location":     Equal(data.ConfigNode.AppSpaceId), // Ingest Pipeline is always on AppSpace level
					"customer_id":  Equal(data.ConfigNode.CustomerId),
					"app_space_id": Equal(data.ConfigNode.AppSpaceId),
					"name":         Equal(data.ConfigNode.Name),
					"display_name": Equal(data.ConfigNode.DisplayName),
					"description":  Equal(data.ConfigNode.GetDescription().GetValue()),
					"create_time":  Not(BeEmpty()),
					"update_time":  Not(BeEmpty()),

					"app_agent_token": Equal(data.GetConfigNode().GetIngestPipelineConfig().GetAppAgentToken()),
				}

				operations := data.GetConfigNode().GetIngestPipelineConfig().GetOperations()
				keys["operations.#"] = Equal(strconv.Itoa(len(operations)))
				for i, op := range operations {
					keys[fmt.Sprintf("operations.%d", i)] = Equal(indykite.IngestPipelineOperationTypesReverse[op])
				}

				sources := data.GetConfigNode().GetIngestPipelineConfig().GetSources()
				keys["sources.#"] = Equal(strconv.Itoa(len(sources)))
				addStringArrayToKeys(keys, "sources", sources)

				return convertOmegaMatcherToError(MatchAllKeys(keys), attrs)
			}
		}

		validSettings := `
		sources = ["source1", "source2"]
		operations = ["OPERATION_UPSERT_NODE", "OPERATION_UPSERT_RELATIONSHIP"]
		app_agent_token = "` + appAgentToken + `"
		`

		validSettingsUpdate := `
		sources = ["source1", "source2", "source3"]
		operations = ["OPERATION_UPSERT_NODE", "OPERATION_UPSERT_RELATIONSHIP", "OPERATION_DELETE_NODE"]
		app_agent_token = "` + appAgentToken + `"
		`

		// Invalid token
		invalidSettings1 := `
		sources = ["source1", "source2"]
		operations = ["OPERATION_UPSERT_NODE", "OPERATION_UPSERT_RELATIONSHIP"]
		app_agent_token = "invalid-token"
		`

		// Missing required argument
		invalidSettings2 := `
		operations = ["OPERATION_UPSERT_NODE", "OPERATION_UPSERT_RELATIONSHIP", "OPERATION_DELETE_NODE"]
		app_agent_token = "` + appAgentToken + `"
		`

		// Empty slice
		invalidSettings3 := `
		sources = ["source1", "source2"]
		operations = []
		app_agent_token = "` + appAgentToken + `"
		`
		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors case must be always first
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", invalidSettings1),
					ExpectError: regexp.MustCompile("invalid value for app_agent_token"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, "ccc", "name", validSettings),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", invalidSettings2),
					ExpectError: regexp.MustCompile("Missing required argument"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", invalidSettings3),
					ExpectError: regexp.MustCompile("Not enough list items"),
				},

				{
					// Checking Create and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-ingest-pipeline",
						`display_name = "Display name of Ingest Pipeline configuration"
						`+validSettings+``,
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
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-ingest-pipeline",
						`description = "ingest pipeline description"
						`+validSettingsUpdate+``,
					),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(resourceName, expectedUpdatedResp),
					),
				},
			},
		})
	})
})
