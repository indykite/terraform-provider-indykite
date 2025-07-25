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
	"io"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("DataSource ApplicationAgent", func() {
	const resourceName = "data.indykite_application_agent.development"
	var (
		mockCtrl                        *gomock.Controller
		mockConfigClient                *configm.MockConfigManagementAPIClient
		mockListApplicationAgentsClient *configm.MockConfigManagementAPI_ListApplicationAgentsClient
		provider                        *schema.Provider
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)
		mockListApplicationAgentsClient = configm.NewMockConfigManagementAPI_ListApplicationAgentsClient(mockCtrl)
		provider = indykite.Provider()
		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc =
			func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
				client, _ := config.NewTestClient(ctx, mockConfigClient)
				ctx = indykite.WithClient(ctx, client)
				return cfgFunc(ctx, data)
			}
	})

	It("Test load by ID and name", func() {
		appAgentResp := &configpb.ApplicationAgent{
			CustomerId:           customerID,
			AppSpaceId:           appSpaceID,
			ApplicationId:        applicationID,
			Id:                   appAgentID,
			Name:                 "acme",
			DisplayName:          "Some Cool Display name",
			Description:          wrapperspb.String("ApplicationAgent description"),
			ApiAccessRestriction: []string{"Authorization", "Capture"},
			CreateTime:           timestamppb.Now(),
			UpdateTime:           timestamppb.Now(),
		}

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{
						"Name": PointTo(MatchFields(IgnoreExtras, Fields{
							"Name":     Equal(appAgentResp.GetName()),
							"Location": Equal(appSpaceID),
						})),
					})),
				})))).
				Return(nil, status.Error(codes.Unknown, "unknown name")),

			mockConfigClient.EXPECT().
				ReadApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(appAgentID)})),
				})))).
				Times(5).
				Return(&configpb.ReadApplicationAgentResponse{ApplicationAgent: appAgentResp}, nil),

			mockConfigClient.EXPECT().
				ReadApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{
						"Name": PointTo(MatchFields(IgnoreExtras, Fields{
							"Name":     Equal(appAgentResp.GetName()),
							"Location": Equal(appSpaceID),
						})),
					})),
				})))).
				Times(5).
				Return(&configpb.ReadApplicationAgentResponse{ApplicationAgent: appAgentResp}, nil),
		)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					Config: `data "indykite_application_agent" "development" {
						customer_id = "` + customerID + `"
						app_space_id = "` + appSpaceID + `"
						name = "acme"
						api_permissions = ["Authorization","Capture"]
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "customer_id"`),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						application_id = "` + applicationID + `"
						app_space_id = "` + appSpaceID + `"
						name = "acme"
						api_permissions = ["Authorization","Capture"]
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "application_id"`),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						name = "acme"
						api_permissions = ["Authorization","Capture"]
						app_agent_id = "` + applicationID + `"
					}`,
					ExpectError: regexp.MustCompile("only one of `app_agent_id,name` can be specified"),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						display_name = "anything"
						api_permissions = ["Authorization","Capture"]
					}`,
					ExpectError: regexp.MustCompile("one of `app_agent_id,name` must be specified"),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						name = "anything"
						api_permissions = ["Authorization","Capture"]
					}`,
					ExpectError: regexp.MustCompile("\"name\": all of `app_space_id,name` must be specified"),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						app_space_id = "` + appSpaceID + `"
						name = "acme"
						api_permissions = ["Authorization","Capture"]
					}`,
					ExpectError: regexp.MustCompile("unknown name"),
				},
				// Success test cases
				{
					Check: resource.ComposeTestCheckFunc(
						testApplicationAgentDataExists(resourceName, appAgentResp, appAgentID),
					),
					Config: `data "indykite_application_agent" "development" {
						app_agent_id = "` + appAgentID + `"
						api_permissions = ["Authorization","Capture"]
					}`,
				},
				{
					Check: resource.ComposeTestCheckFunc(
						testApplicationAgentDataExists(resourceName, appAgentResp, "")),
					Config: `data "indykite_application_agent" "development" {
						app_space_id = "` + appSpaceID + `"
						name = "acme"
						api_permissions = ["Authorization","Capture"]
					}`,
				},
			},
		})
	})

	It("Test list by multiple names", func() {
		appAgentResp := &configpb.ApplicationAgent{
			CustomerId:           customerID,
			AppSpaceId:           appSpaceID,
			ApplicationId:        applicationID,
			Id:                   appAgentID,
			Name:                 "loompaland",
			DisplayName:          "Some Cool Display name",
			Description:          wrapperspb.String("Just some ApplicationAgent description"),
			CreateTime:           timestamppb.Now(),
			UpdateTime:           timestamppb.Now(),
			ApiAccessRestriction: []string{"Authorization", "Capture"},
		}
		appAgentResp2 := &configpb.ApplicationAgent{
			CustomerId:           customerID,
			AppSpaceId:           appSpaceID,
			ApplicationId:        applicationID,
			Id:                   sampleID,
			Name:                 "wonka-opa-agent",
			CreateTime:           timestamppb.Now(),
			UpdateTime:           timestamppb.Now(),
			ApiAccessRestriction: []string{"Authorization", "Capture"},
		}

		mockConfigClient.EXPECT().
			ListApplicationAgents(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Match":      ConsistOf("loompaland", "some-another-name", "wonka-opa-agent"),
				"AppSpaceId": Equal(appSpaceID),
			})))).
			Times(5).
			DoAndReturn(
				func(
					_, _ any,
					_ ...any,
				) (*configm.MockConfigManagementAPI_ListApplicationAgentsClient, error) {
					mockListApplicationAgentsClient.EXPECT().Recv().
						Return(&configpb.ListApplicationAgentsResponse{ApplicationAgent: appAgentResp}, nil)
					mockListApplicationAgentsClient.EXPECT().Recv().
						Return(&configpb.ListApplicationAgentsResponse{ApplicationAgent: appAgentResp2}, nil)
					mockListApplicationAgentsClient.EXPECT().Recv().Return(nil, io.EOF)
					return mockListApplicationAgentsClient, nil
				},
			)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					Config: `data "indykite_application_agents" "development" {
						filter = "acme"
						app_space_id = "` + appSpaceID + `"
					}`,
					ExpectError: regexp.MustCompile(
						`Inappropriate value for attribute "filter": list of string required`),
				},
				{
					Config: `data "indykite_application_agents" "development" {
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile(`The argument "app_space_id" is required`),
				},
				{
					Config: `data "indykite_application_agents" "development" {
						filter = []
						app_space_id = "` + appSpaceID + `"
					}`,
					ExpectError: regexp.MustCompile("Attribute filter requires 1 item minimum, but config has only 0"),
				},
				{
					Config: `data "indykite_application_agents" "development" {
						app_space_id = "` + appSpaceID + `"
						filter = [123]
					}`,
					ExpectError: regexp.MustCompile("Value can have lowercase letters, digits, or hyphens."),
				},
				{
					Config: `data "indykite_application_agents" "development" {
						app_space_id = "abc"
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile("expected to have 'gid:' prefix"),
				},
				{
					Config: `data "indykite_application_agents" "development" {
						app_space_id = "` + appSpaceID + `"
						filter = ["acme"]
						app_agents = []
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "app_agents":`),
				},
				// Success test cases
				{
					Check: resource.ComposeTestCheckFunc(testApplicationAgentListDataExists(
						"data.indykite_application_agents.development",
						appAgentResp,
						appAgentResp2)),
					Config: `data "indykite_application_agents" "development" {
						app_space_id = "` + appSpaceID + `"
						filter = ["loompaland", "some-another-name", "wonka-opa-agent"]
					}`,
				},
			},
		})
	})
})

func testApplicationAgentDataExists(
	n string,
	data *configpb.ApplicationAgent,
	appAgentID string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != data.GetId() {
			return errors.New("ID does not match")
		}

		keys := Keys{
			"id":                Equal(data.GetId()),
			"%":                 Not(BeEmpty()), // This is Terraform helper
			"api_permissions.#": Equal("2"),
			"api_permissions.0": Equal("Authorization"),
			"api_permissions.1": Equal("Capture"),

			"customer_id":    Equal(data.GetCustomerId()),
			"app_space_id":   Equal(data.GetAppSpaceId()),
			"application_id": Equal(data.GetApplicationId()),
			"name":           Equal(data.GetName()),
			"display_name":   Equal(data.GetDisplayName()),
			"description":    Equal(data.GetDescription().GetValue()),
			"create_time":    Not(BeEmpty()),
			"update_time":    Not(BeEmpty()),
		}
		if appAgentID != "" {
			keys["app_agent_id"] = Equal(appAgentID)
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}

func testApplicationAgentListDataExists(n string, data ...*configpb.ApplicationAgent) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		expectedID := "gid:AAAAAmluZHlraURlgAABDwAAAAA/app_agents/loompaland,some-another-name,wonka-opa-agent"
		if rs.Primary.ID != expectedID {
			return fmt.Errorf("expected ID to be '%s' got '%s'", expectedID, rs.Primary.ID)
		}

		keys := Keys{
			"id": Equal(expectedID),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"app_space_id": Equal(appSpaceID),

			"app_agents.#": Equal(strconv.Itoa(len(data))), // This is Terraform helper
			"filter.#":     Equal("3"),
			"filter.0":     Equal("loompaland"),
			"filter.1":     Equal("some-another-name"),
			"filter.2":     Equal("wonka-opa-agent"),
		}

		for i, d := range data {
			k := "app_agents." + strconv.Itoa(i) + "."
			keys[k+"%"] = Not(BeEmpty()) // This is Terraform helper
			keys[k+"api_permissions.#"] = Equal("2")
			keys[k+"api_permissions.0"] = Equal("Authorization")
			keys[k+"api_permissions.1"] = Equal("Capture")
			keys[k+"id"] = Equal(d.GetId())
			keys[k+"customer_id"] = Equal(d.GetCustomerId())
			keys[k+"app_space_id"] = Equal(d.GetAppSpaceId())
			keys[k+"application_id"] = Equal(d.GetApplicationId())
			keys[k+"name"] = Equal(d.GetName())
			keys[k+"display_name"] = Equal(d.GetDisplayName())
			keys[k+"description"] = Equal(d.GetDescription().GetValue())
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}
