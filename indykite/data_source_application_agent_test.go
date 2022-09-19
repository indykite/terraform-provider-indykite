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
	"io"
	"regexp"
	"strconv"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/jarvis-sdk-go/config"
	configpb "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/jarvis-sdk-go/test/config/v1beta1"
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
		indykiteProviderFactory         func() (*schema.Provider, error)
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)
		mockListApplicationAgentsClient = configm.NewMockConfigManagementAPI_ListApplicationAgentsClient(mockCtrl)

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

	It("Test load by ID and name", func() {
		appAgentResp := &configpb.ApplicationAgent{
			CustomerId:    customerID,
			AppSpaceId:    appSpaceID,
			ApplicationId: applicationID,
			Id:            appAgentID,
			Name:          "acme",
			DisplayName:   "Some Cool Display name",
			Description:   wrapperspb.String("ApplicationAgent description"),
			CreateTime:    timestamppb.Now(),
			UpdateTime:    timestamppb.Now(),
		}

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{
						"Name": PointTo(MatchFields(IgnoreExtras, Fields{
							"Name":     Equal(appAgentResp.Name),
							"Location": Equal(appSpaceID),
						})),
					})),
				})))).
				Return(nil, status.Error(codes.NotFound, "unknown name")),

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
							"Name":     Equal(appAgentResp.Name),
							"Location": Equal(appSpaceID),
						})),
					})),
				})))).
				Times(5).
				Return(&configpb.ReadApplicationAgentResponse{ApplicationAgent: appAgentResp}, nil),
		)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					Config: `data "indykite_application_agent" "development" {
						customer_id = "` + customerID + `"
						app_space_id = "` + appSpaceID + `"
						name = "acme"
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "customer_id"`),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						application_id = "` + applicationID + `"
						app_space_id = "` + appSpaceID + `"
						name = "acme"
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "application_id"`),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						name = "acme"
						app_agent_id = "` + applicationID + `"
					}`,
					ExpectError: regexp.MustCompile("only one of `app_agent_id,name` can be specified"),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						display_name = "anything"
					}`,
					ExpectError: regexp.MustCompile("one of `app_agent_id,name` must be specified"),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						name = "anything"
					}`,
					ExpectError: regexp.MustCompile("\"name\": all of `app_space_id,name` must be specified"),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						app_space_id = "` + appSpaceID + `"
						name = "acme"
					}`,
					ExpectError: regexp.MustCompile("unknown name"),
				},

				// Success test cases
				{
					Check: resource.ComposeTestCheckFunc(testApplicationAgentDataExists(resourceName, appAgentResp)),
					Config: `data "indykite_application_agent" "development" {
						app_agent_id = "` + appAgentID + `"
					}`,
				},
				{
					Check: resource.ComposeTestCheckFunc(testApplicationAgentDataExists(resourceName, appAgentResp)),
					Config: `data "indykite_application_agent" "development" {
						app_space_id = "` + appSpaceID + `"
						name = "acme"
					}`,
				},
			},
		})
	})

	It("Test list by multiple names", func() {
		appAgentResp := &configpb.ApplicationAgent{
			CustomerId:    customerID,
			AppSpaceId:    appSpaceID,
			ApplicationId: applicationID,
			Id:            appAgentID,
			Name:          "loompaland",
			DisplayName:   "Some Cool Display name",
			Description:   wrapperspb.String("Just some ApplicationAgent description"),
			CreateTime:    timestamppb.Now(),
			UpdateTime:    timestamppb.Now(),
		}
		appAgentResp2 := &configpb.ApplicationAgent{
			CustomerId:    customerID,
			AppSpaceId:    appSpaceID,
			ApplicationId: applicationID,
			Id:            sampleID,
			Name:          "wonka-opa-agent",
			CreateTime:    timestamppb.Now(),
			UpdateTime:    timestamppb.Now(),
		}

		mockConfigClient.EXPECT().
			ListApplicationAgents(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Match":      ConsistOf("loompaland", "some-another-name", "wonka-opa-agent"),
				"AppSpaceId": Equal(appSpaceID),
			})))).
			Times(5).
			DoAndReturn(
				func(
					_, _ interface{},
					_ ...interface{},
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
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
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

func testApplicationAgentDataExists(n string, data *configpb.ApplicationAgent) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID != data.Id {
			return fmt.Errorf("ID does not match")
		}
		if v, has := rs.Primary.Attributes["customer_id"]; !has || v != data.CustomerId {
			return fmt.Errorf("invalid customer_id: %s", v)
		}
		if v, has := rs.Primary.Attributes["app_space_id"]; !has || v != data.AppSpaceId {
			return fmt.Errorf("invalid app_space_id: %s", v)
		}
		if v, has := rs.Primary.Attributes["application_id"]; !has || v != data.ApplicationId {
			return fmt.Errorf("invalid application_id: %s", v)
		}
		if v, has := rs.Primary.Attributes["name"]; !has || v != data.Name {
			return fmt.Errorf("invalid name: %s", v)
		}
		if v, has := rs.Primary.Attributes["display_name"]; !has || v != data.DisplayName {
			return fmt.Errorf("invalid display name: %s", v)
		}
		if v, has := rs.Primary.Attributes["description"]; !has || v != data.Description.GetValue() {
			return fmt.Errorf("invalid description: %s", v)
		}

		return nil
	}
}

func testApplicationAgentListDataExists(n string, data ...*configpb.ApplicationAgent) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		attrs := rs.Primary.Attributes

		expectedID := "gid:AAAAAmluZHlraURlgAABDwAAAAA/app_agents/loompaland,some-another-name,wonka-opa-agent"
		if rs.Primary.ID != expectedID {
			return fmt.Errorf("expected ID to be '%s' got '%s'", expectedID, rs.Primary.ID)
		}

		if attrs["app_agents.#"] != strconv.Itoa(len(data)) {
			return fmt.Errorf("expected %d app_agents, but got %s", len(data), attrs["app_agents.#"])
		}
		for i, d := range data {
			k := "app_agents." + strconv.Itoa(i) + "."
			if v, has := attrs[k+"id"]; !has || v != d.Id {
				return fmt.Errorf("%d. entry for 'id' expected '%s', got '%s'", i, d.Id, v)
			}

			if v, has := attrs[k+"customer_id"]; !has || v != d.CustomerId {
				return fmt.Errorf("%d. entry for 'customer_id' expected '%s', got '%s'", i, d.CustomerId, v)
			}
			if v, has := attrs[k+"app_space_id"]; !has || v != d.AppSpaceId {
				return fmt.Errorf("%d. entry for 'app_space_id' expected '%s', got '%s'", i, d.AppSpaceId, v)
			}
			if v, has := attrs[k+"application_id"]; !has || v != d.ApplicationId {
				return fmt.Errorf("%d. entry for 'application_id' expected '%s', got '%s'", i, d.ApplicationId, v)
			}

			if v, has := attrs[k+"name"]; !has || v != d.Name {
				return fmt.Errorf("%d. entry for 'name' expected '%s', got '%s'", i, d.Name, v)
			}
			if v, has := attrs[k+"display_name"]; !has || v != d.DisplayName {
				return fmt.Errorf("%d. entry for 'display_name' expected '%s', got '%s'", i, d.DisplayName, v)
			}
			if v, has := attrs[k+"description"]; !has || v != d.Description.GetValue() {
				return fmt.Errorf("%d. entry for 'description' expected '%s', got '%s'", i, d.Description.GetValue(), v)
			}
		}
		return nil
	}
}
