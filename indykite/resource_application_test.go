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

var _ = Describe("Resource Application", func() {
	const resourceName = "indykite_application.development"
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
		mockedBookmark = "for-application" + uuid.NewRandom().String()
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
			`resource "indykite_application" "development" {
				app_space_id = "` + appSpaceID + `"
				name = "acme"
				display_name = "%s"
				description = "%s"
				%s
			}`
		initialApplicationResp := &configpb.Application{
			CustomerId:  customerID,
			AppSpaceId:  appSpaceID,
			Id:          applicationID,
			Name:        "acme",
			DisplayName: "acme",
			Description: wrapperspb.String("Just some App description"),
			CreateTime:  timestamppb.Now(),
			UpdateTime:  timestamppb.Now(),
		}

		readAfter1stUpdateResp := &configpb.Application{
			CustomerId:  initialApplicationResp.CustomerId,
			AppSpaceId:  initialApplicationResp.AppSpaceId,
			Id:          initialApplicationResp.Id,
			Name:        "acme",
			DisplayName: "acme",
			Description: wrapperspb.String("Another App description"),
			CreateTime:  initialApplicationResp.CreateTime,
			UpdateTime:  timestamppb.Now(),
		}
		readAfter2ndUpdateResp := &configpb.Application{
			CustomerId:  initialApplicationResp.CustomerId,
			AppSpaceId:  initialApplicationResp.AppSpaceId,
			Id:          initialApplicationResp.Id,
			Name:        "acme",
			DisplayName: "Some new display name",
			Description: nil,
			CreateTime:  initialApplicationResp.CreateTime,
			UpdateTime:  timestamppb.Now(),
		}

		createBM := "created-application" + uuid.NewRandom().String()
		updateBM := "updated-application" + uuid.NewRandom().String()
		updateBM2 := "updated-application-2" + uuid.NewRandom().String()
		deleteBM := "deleted-application" + uuid.NewRandom().String()

		// Create
		mockConfigClient.EXPECT().
			CreateApplication(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"AppSpaceId":  Equal(appSpaceID),
				"Name":        Equal(initialApplicationResp.Name),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(initialApplicationResp.Description.Value),
				})),
				"Bookmarks": ConsistOf(mockedBookmark),
			})))).
			Return(&configpb.CreateApplicationResponse{Id: initialApplicationResp.Id, Bookmark: createBM}, nil)

		// 2x update
		mockConfigClient.EXPECT().
			UpdateApplication(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(initialApplicationResp.Id),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter1stUpdateResp.Description.Value),
				})),
				"Bookmarks": ConsistOf(mockedBookmark, createBM),
			})))).
			Return(&configpb.UpdateApplicationResponse{Id: initialApplicationResp.Id, Bookmark: updateBM}, nil)

		mockConfigClient.EXPECT().
			UpdateApplication(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialApplicationResp.Id),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter2ndUpdateResp.DisplayName),
				})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal("")})),
				"Bookmarks":   ConsistOf(mockedBookmark, createBM, updateBM),
			})))).
			Return(&configpb.UpdateApplicationResponse{Id: initialApplicationResp.Id, Bookmark: updateBM2}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadApplication(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialApplicationResp.Id)})),
					"Bookmarks":  ConsistOf(mockedBookmark, createBM),
				})))).
				Times(4).
				Return(&configpb.ReadApplicationResponse{Application: initialApplicationResp}, nil),

			mockConfigClient.EXPECT().
				ReadApplication(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialApplicationResp.Id)})),
					"Bookmarks":  ConsistOf(mockedBookmark, createBM, updateBM),
				})))).
				Times(3).
				Return(&configpb.ReadApplicationResponse{Application: readAfter1stUpdateResp}, nil),

			mockConfigClient.EXPECT().
				ReadApplication(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialApplicationResp.Id)})),
					"Bookmarks":  ConsistOf(mockedBookmark, createBM, updateBM, updateBM2),
				})))).
				Times(5).
				Return(&configpb.ReadApplicationResponse{Application: readAfter2ndUpdateResp}, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteApplication(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":        Equal(initialApplicationResp.Id),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM, updateBM2),
			})))).
			Return(&configpb.DeleteApplicationResponse{Bookmark: deleteBM}, nil)

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
					// Checking Create and Read (initialApplicationResp)
					Config: fmt.Sprintf(tfConfigDef, "", initialApplicationResp.Description.Value, ""),
					Check: resource.ComposeTestCheckFunc(
						testAppResourceDataExists(resourceName, initialApplicationResp),
					),
				},
				{
					// Performs 1 read (initialApplicationResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: initialApplicationResp.Id,
				},
				{
					// Checking Read (initialApplicationResp), Update and Read(readAfter1stUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, "", readAfter1stUpdateResp.Description.Value, ""),
					Check: resource.ComposeTestCheckFunc(
						testAppResourceDataExists(resourceName, readAfter1stUpdateResp),
					),
				},
				{
					// Checking Read(readAfter1stUpdateResp), Update and Read(readAfter2ndUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.DisplayName, "", ""),
					Check: resource.ComposeTestCheckFunc(
						testAppResourceDataExists(resourceName, readAfter2ndUpdateResp),
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

func testAppResourceDataExists(n string, data *configpb.Application) resource.TestCheckFunc {
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
