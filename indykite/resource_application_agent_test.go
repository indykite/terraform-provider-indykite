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

var _ = Describe("Resource ApplicationAgent", func() {
	const resourceName = "indykite_application_agent.development"
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
		mockedBookmark = "for-app-agent" + uuid.NewRandom().String()
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

		createBM := "created-app-agent" + uuid.NewRandom().String()
		updateBM := "updated-app-agent" + uuid.NewRandom().String()
		updateBM2 := "updated-app-agent-2" + uuid.NewRandom().String()
		deleteBM := "deleted-app-agent" + uuid.NewRandom().String()

		// Create
		mockConfigClient.EXPECT().
			CreateApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"ApplicationId": Equal(applicationID),
				"Name":          Equal(initialAppAgentResp.Name),
				"DisplayName":   BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(initialAppAgentResp.Description.Value),
				})),
				"Bookmarks": ConsistOf(mockedBookmark),
			})))).
			Return(&configpb.CreateApplicationAgentResponse{Id: initialAppAgentResp.Id, Bookmark: createBM}, nil)

		// 2x update
		mockConfigClient.EXPECT().
			UpdateApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(initialAppAgentResp.Id),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter1stUpdateResp.Description.Value),
				})),
				"Bookmarks": ConsistOf(mockedBookmark, createBM),
			})))).
			Return(&configpb.UpdateApplicationAgentResponse{Id: initialAppAgentResp.Id, Bookmark: updateBM}, nil)

		mockConfigClient.EXPECT().
			UpdateApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialAppAgentResp.Id),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter2ndUpdateResp.DisplayName),
				})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal("")})),
				"Bookmarks":   ConsistOf(mockedBookmark, createBM, updateBM),
			})))).
			Return(&configpb.UpdateApplicationAgentResponse{Id: initialAppAgentResp.Id, Bookmark: updateBM2}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialAppAgentResp.Id)})),
					"Bookmarks":  ConsistOf(mockedBookmark, createBM),
				})))).
				Times(4).
				Return(&configpb.ReadApplicationAgentResponse{ApplicationAgent: initialAppAgentResp}, nil),

			mockConfigClient.EXPECT().
				ReadApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialAppAgentResp.Id)})),
					"Bookmarks":  ConsistOf(mockedBookmark, createBM, updateBM),
				})))).
				Times(3).
				Return(&configpb.ReadApplicationAgentResponse{ApplicationAgent: readAfter1stUpdateResp}, nil),

			mockConfigClient.EXPECT().
				ReadApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialAppAgentResp.Id)})),
					"Bookmarks":  ConsistOf(mockedBookmark, createBM, updateBM, updateBM2),
				})))).
				Times(5).
				Return(&configpb.ReadApplicationAgentResponse{ApplicationAgent: readAfter2ndUpdateResp}, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteApplicationAgent(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":        Equal(initialAppAgentResp.Id),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM, updateBM2),
			})))).
			Return(&configpb.DeleteApplicationAgentResponse{Bookmark: deleteBM}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
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
			return errors.New("ID does not match")
		}

		keys := Keys{
			"id": Equal(data.Id),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"customer_id":         Equal(data.CustomerId),
			"app_space_id":        Equal(data.AppSpaceId),
			"application_id":      Equal(data.ApplicationId),
			"name":                Equal(data.Name),
			"display_name":        Equal(data.DisplayName),
			"description":         Equal(data.Description.GetValue()),
			"create_time":         Not(BeEmpty()),
			"update_time":         Not(BeEmpty()),
			"deletion_protection": Not(BeEmpty()),
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}
