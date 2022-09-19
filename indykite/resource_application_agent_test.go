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
	configm "github.com/indykite/jarvis-sdk-go/test/config/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource ApplicationAgent", func() {
	const resourceName = "indykite_application_agent.development"
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
		turnOffDelProtection := "deletion_protection=false"
		// Terraform created config must be in sync with returned data in expectedApp and expectedUpdatedApp
		// otherwise "After applying this test step, the plan was not empty" error is thrown
		tfConfigDef :=
			`resource "indykite_application_agent" "development" {
				application_id = "` + applicationID + `"
				name = "acme"
				display_name = "%s"
				description = "%s"
				%s
			}`
		initialAppAgentResp := &configpb.ApplicationAgent{
			CustomerId:    customerID,
			AppSpaceId:    appSpaceID,
			ApplicationId: applicationID,
			Id:            appAgentID,
			Name:          "acme",
			DisplayName:   "acme",
			Description:   wrapperspb.String("Just some App description"),
			CreateTime:    timestamppb.Now(),
			UpdateTime:    timestamppb.Now(),
		}

		readAfter1stUpdateResp := &configpb.ApplicationAgent{
			CustomerId:    initialAppAgentResp.CustomerId,
			AppSpaceId:    initialAppAgentResp.AppSpaceId,
			ApplicationId: initialAppAgentResp.ApplicationId,
			Id:            initialAppAgentResp.Id,
			Name:          "acme",
			DisplayName:   "acme",
			Description:   wrapperspb.String("Another App description"),
			CreateTime:    initialAppAgentResp.CreateTime,
			UpdateTime:    timestamppb.Now(),
		}
		readAfter2ndUpdateResp := &configpb.ApplicationAgent{
			CustomerId:    initialAppAgentResp.CustomerId,
			AppSpaceId:    initialAppAgentResp.AppSpaceId,
			ApplicationId: initialAppAgentResp.ApplicationId,
			Id:            initialAppAgentResp.Id,
			Name:          "acme",
			DisplayName:   "Some new display name",
			Description:   nil,
			CreateTime:    initialAppAgentResp.CreateTime,
			UpdateTime:    timestamppb.Now(),
		}

		// MOCKS
		// There are 5 test steps
		// 1. step call: Create + Read
		// 2. step call: Read, Update, Read
		// 3. step call: Read, Update, Read
		// 4. step call: Read + delete (going to fail)
		// 5. step call: Read (changes only in deletion_protection do not trigger API)
		// after steps Delete is called

		// Create
		mockConfigClient.EXPECT().
			CreateApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"ApplicationId": Equal(applicationID),
				"Name":          Equal(initialAppAgentResp.Name),
				"DisplayName":   BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(initialAppAgentResp.Description.Value),
				})),
			})))).
			Return(&configpb.CreateApplicationAgentResponse{Id: initialAppAgentResp.Id}, nil)

		// 2x update
		mockConfigClient.EXPECT().
			UpdateApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(initialAppAgentResp.Id),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter1stUpdateResp.Description.Value),
				})),
			})))).
			Return(&configpb.UpdateApplicationAgentResponse{Id: initialAppAgentResp.Id}, nil)

		mockConfigClient.EXPECT().
			UpdateApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialAppAgentResp.Id),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter2ndUpdateResp.DisplayName),
				})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal("")})),
			})))).
			Return(&configpb.UpdateApplicationAgentResponse{Id: initialAppAgentResp.Id}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialAppAgentResp.Id)})),
				})))).
				Times(4).
				Return(&configpb.ReadApplicationAgentResponse{ApplicationAgent: initialAppAgentResp}, nil),

			mockConfigClient.EXPECT().
				ReadApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialAppAgentResp.Id)})),
				})))).
				Times(3).
				Return(&configpb.ReadApplicationAgentResponse{ApplicationAgent: readAfter1stUpdateResp}, nil),

			mockConfigClient.EXPECT().
				ReadApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialAppAgentResp.Id)})),
				})))).
				Times(5).
				Return(&configpb.ReadApplicationAgentResponse{ApplicationAgent: readAfter2ndUpdateResp}, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialAppAgentResp.Id),
			})))).
			Return(&configpb.DeleteApplicationAgentResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					Config:      fmt.Sprintf(tfConfigDef, "", "", `customer_id = "`+customerID+`"`),
					ExpectError: regexp.MustCompile("Value for unconfigurable attribute"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, "", "", `app_space_id = "`+appSpaceID+`"`),
					ExpectError: regexp.MustCompile("Value for unconfigurable attribute"),
				},
				{
					// Checking Create and Read (initialAppAgentResp)
					Config: fmt.Sprintf(tfConfigDef, "", initialAppAgentResp.Description.Value, ""),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentResourceDataExists(resourceName, initialAppAgentResp),
					),
				},
				{
					// Performs 1 read (initialAppAgentResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: initialAppAgentResp.Id,
				},
				{
					// Checking Read (initialAppAgentResp), Update and Read(readAfter1stUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, "", readAfter1stUpdateResp.Description.Value, ""),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentResourceDataExists(resourceName, readAfter1stUpdateResp),
					),
				},
				{
					// Checking Read(readAfter1stUpdateResp), Update and Read(readAfter2ndUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.DisplayName, "", ""),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentResourceDataExists(resourceName, readAfter2ndUpdateResp),
					),
				},
				{
					// Checking Read(readAfter2ndUpdateResp) -> no changes but tries to destroy with error
					Config:      fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.DisplayName, "", ""),
					Destroy:     true,
					ExpectError: regexp.MustCompile("Cannot destroy instance"),
				},
				{
					// Checking Read(readAfter2ndUpdateResp), Update (del protection, no API call)
					// and final Read (readAfter2ndUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.DisplayName, "", turnOffDelProtection),
				},
			},
		})
	})
})

func testAppAgentResourceDataExists(n string, data *configpb.ApplicationAgent) resource.TestCheckFunc {
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
			//
			if data.DisplayName != data.Name {
				return fmt.Errorf("invalid display name: %s", v)
			}
		}
		if v, has := rs.Primary.Attributes["description"]; !has || v != data.Description.GetValue() {
			return fmt.Errorf("invalid description: %s", v)
		}

		return nil
	}
}
