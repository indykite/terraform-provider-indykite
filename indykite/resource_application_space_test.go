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
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource Application Space", func() {
	const resourceName = "indykite_application_space.development"
	const resourceNameSimple = "indykite_application_space.developmentSimple"
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

	It("Test all CRUD Simple", func() {
		turnOffDelProtection := "deletion_protection=false"
		// Terraform create config must be in sync with returned data in expectedAppSpace and expectedUpdatedAppSpace
		// otherwise "After applying this test step, the plan was not empty" error is thrown
		tfConfigDefSimple :=
			`resource "indykite_application_space" "developmentSimple" {
				customer_id = "` + customerID + `"
				name = "acme0"
				display_name = "%s"
				description = "%s"
				region = "europe-west1"
				ikg_size = "4GB"
				%s
			}`

		initialAppSpaceRespSimple := &configpb.ApplicationSpace{
			CustomerId:  customerID,
			Id:          appSpaceID,
			Name:        "acme0",
			DisplayName: "acme0",
			Description: wrapperspb.String("Just some AppSpace description"),
			CreateTime:  timestamppb.Now(),
			UpdateTime:  timestamppb.Now(),
			Region:      "europe-west1",
			IkgSize:     "4GB",
		}

		readAfter1stUpdateRespSimple := &configpb.ApplicationSpace{
			CustomerId:  customerID,
			Id:          initialAppSpaceRespSimple.GetId(),
			Name:        "acme0",
			DisplayName: "acme0",
			Description: wrapperspb.String("Another AppSpace description"),
			CreateTime:  initialAppSpaceRespSimple.GetCreateTime(),
			UpdateTime:  timestamppb.Now(),
			Region:      "europe-west1",
			IkgSize:     "4GB",
		}
		readAfter2ndUpdateRespSimple := &configpb.ApplicationSpace{
			CustomerId:  customerID,
			Id:          initialAppSpaceRespSimple.GetId(),
			Name:        "acme0",
			DisplayName: "Some new display name",
			Description: nil,
			CreateTime:  initialAppSpaceRespSimple.GetCreateTime(),
			UpdateTime:  timestamppb.Now(),
			Region:      "europe-west1",
			IkgSize:     "4GB",
		}

		// Create1
		mockConfigClient.EXPECT().
			CreateApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"CustomerId":  Equal(customerID),
				"Name":        Equal(initialAppSpaceRespSimple.GetName()),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(initialAppSpaceRespSimple.GetDescription().GetValue()),
				})),
				"Region":  Equal(initialAppSpaceRespSimple.GetRegion()),
				"IkgSize": Equal(initialAppSpaceRespSimple.GetIkgSize()),
			})))).
			Return(&configpb.CreateApplicationSpaceResponse{Id: initialAppSpaceRespSimple.GetId()}, nil)

		// 2x update
		mockConfigClient.EXPECT().
			UpdateApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(initialAppSpaceRespSimple.GetId()),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter1stUpdateRespSimple.GetDescription().GetValue()),
				})),
			})))).
			Return(&configpb.UpdateApplicationSpaceResponse{Id: initialAppSpaceRespSimple.GetId()}, nil)

		mockConfigClient.EXPECT().
			UpdateApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialAppSpaceRespSimple.GetId()),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter2ndUpdateRespSimple.GetDisplayName()),
				})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal("")})),
			})))).
			Return(&configpb.UpdateApplicationSpaceResponse{Id: initialAppSpaceRespSimple.GetId()}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{
						"Id": Equal(initialAppSpaceRespSimple.GetId())})),
				})))).
				Times(4).
				Return(&configpb.ReadApplicationSpaceResponse{AppSpace: initialAppSpaceRespSimple}, nil),

			mockConfigClient.EXPECT().
				ReadApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{
						"Id": Equal(initialAppSpaceRespSimple.GetId())})),
				})))).
				Times(3).
				Return(&configpb.ReadApplicationSpaceResponse{AppSpace: readAfter1stUpdateRespSimple}, nil),

			mockConfigClient.EXPECT().
				ReadApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{
						"Id": Equal(initialAppSpaceRespSimple.GetId())})),
				})))).
				Times(5).
				Return(&configpb.ReadApplicationSpaceResponse{AppSpace: readAfter2ndUpdateRespSimple}, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialAppSpaceRespSimple.GetId()),
			})))).
			Return(&configpb.DeleteApplicationSpaceResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					// Checking Create and Read (initialAppSpaceRespSimple)
					Config: fmt.Sprintf(
						tfConfigDefSimple, "", initialAppSpaceRespSimple.GetDescription().GetValue(), ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceNameSimple, initialAppSpaceRespSimple),
					),
				},
				{
					// Performs 1 read (initialAppSpaceRespSimple)
					ResourceName:  resourceNameSimple,
					ImportState:   true,
					ImportStateId: initialAppSpaceRespSimple.GetId(),
				},
				{
					// Checking Read (initialAppSpaceRespSimple), Update and Read(readAfter1stUpdateRespSimple)
					Config: fmt.Sprintf(
						tfConfigDefSimple, "", readAfter1stUpdateRespSimple.GetDescription().GetValue(), ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceNameSimple, readAfter1stUpdateRespSimple),
					),
				},
				{
					// Checking Read(readAfter1stUpdateRespSimple), Update and Read(readAfter2ndUpdateRespSimple)
					Config: fmt.Sprintf(tfConfigDefSimple, readAfter2ndUpdateRespSimple.GetDisplayName(), "", ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceNameSimple, readAfter2ndUpdateRespSimple),
					),
				},
				{
					// Checking Read(readAfter2ndUpdateRespSimple) -> no changes but tries to destroy with error
					Config:      fmt.Sprintf(tfConfigDefSimple, readAfter2ndUpdateRespSimple.GetDisplayName(), "", ""),
					Destroy:     true,
					ExpectError: regexp.MustCompile("Cannot destroy instance"),
				},
				{
					// Checking Read(readAfter2ndUpdateRespSimple), Update (del protection, no API call)
					// and final Read (readAfter2ndUpdateRespSimple)
					Config: fmt.Sprintf(
						tfConfigDefSimple, readAfter2ndUpdateRespSimple.GetDisplayName(), "", turnOffDelProtection),
				},
			},
		})
	})

	It("Test all CRUD", func() {
		turnOffDelProtection := "deletion_protection=false"
		// Terraform create config must be in sync with returned data in expectedAppSpace and expectedUpdatedAppSpace
		// otherwise "After applying this test step, the plan was not empty" error is thrown
		tfConfigDef :=
			`resource "indykite_application_space" "development" {
					customer_id = "` + customerID + `"
					name = "acme"
					display_name = "%s"
					description = "%s"
					region = "us-east1"
					ikg_size = "4GB"
					replica_region = "us-west1"
					%s
				}`

		initialAppSpaceResp := &configpb.ApplicationSpace{
			CustomerId:    customerID,
			Id:            appSpaceID,
			Name:          "acme",
			DisplayName:   "acme",
			Description:   wrapperspb.String("Just some AppSpace description"),
			CreateTime:    timestamppb.Now(),
			UpdateTime:    timestamppb.Now(),
			Region:        "us-east1",
			IkgSize:       "4GB",
			ReplicaRegion: "us-west1",
		}

		readAfter1stUpdateResp := &configpb.ApplicationSpace{
			CustomerId:    customerID,
			Id:            initialAppSpaceResp.GetId(),
			Name:          "acme",
			DisplayName:   "acme",
			Description:   wrapperspb.String("Another AppSpace description"),
			CreateTime:    initialAppSpaceResp.GetCreateTime(),
			UpdateTime:    timestamppb.Now(),
			Region:        "us-east1",
			IkgSize:       "4GB",
			ReplicaRegion: "us-west1",
		}
		readAfter2ndUpdateResp := &configpb.ApplicationSpace{
			CustomerId:    customerID,
			Id:            initialAppSpaceResp.GetId(),
			Name:          "acme",
			DisplayName:   "Some new display name",
			Description:   nil,
			CreateTime:    initialAppSpaceResp.GetCreateTime(),
			UpdateTime:    timestamppb.Now(),
			Region:        "us-east1",
			IkgSize:       "4GB",
			ReplicaRegion: "us-west1",
		}

		// Create2
		mockConfigClient.EXPECT().
			CreateApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"CustomerId":  Equal(customerID),
				"Name":        Equal(initialAppSpaceResp.GetName()),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(initialAppSpaceResp.GetDescription().GetValue()),
				})),
				"Region": Equal(initialAppSpaceResp.GetRegion()),
			})))).
			Return(&configpb.CreateApplicationSpaceResponse{Id: initialAppSpaceResp.GetId()}, nil)

		// 2x update
		mockConfigClient.EXPECT().
			UpdateApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(initialAppSpaceResp.GetId()),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter1stUpdateResp.GetDescription().GetValue()),
				})),
			})))).
			Return(&configpb.UpdateApplicationSpaceResponse{Id: initialAppSpaceResp.GetId()}, nil)

		mockConfigClient.EXPECT().
			UpdateApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialAppSpaceResp.GetId()),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter2ndUpdateResp.GetDisplayName()),
				})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal("")})),
			})))).
			Return(&configpb.UpdateApplicationSpaceResponse{Id: initialAppSpaceResp.GetId()}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialAppSpaceResp.GetId())})),
				})))).
				Times(4).
				Return(&configpb.ReadApplicationSpaceResponse{AppSpace: initialAppSpaceResp}, nil),

			mockConfigClient.EXPECT().
				ReadApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialAppSpaceResp.GetId())})),
				})))).
				Times(3).
				Return(&configpb.ReadApplicationSpaceResponse{AppSpace: readAfter1stUpdateResp}, nil),

			mockConfigClient.EXPECT().
				ReadApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialAppSpaceResp.GetId())})),
				})))).
				Times(5).
				Return(&configpb.ReadApplicationSpaceResponse{AppSpace: readAfter2ndUpdateResp}, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialAppSpaceResp.GetId()),
			})))).
			Return(&configpb.DeleteApplicationSpaceResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					// Checking Create and Read (initialAppSpaceResp)
					Config: fmt.Sprintf(tfConfigDef, "", initialAppSpaceResp.GetDescription().GetValue(), ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceName, initialAppSpaceResp),
					),
				},
				{
					// Performs 1 read (initialAppSpaceResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: initialAppSpaceResp.GetId(),
				},
				{
					// Checking Read (initialAppSpaceResp), Update and Read(readAfter1stUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, "", readAfter1stUpdateResp.GetDescription().GetValue(), ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceName, readAfter1stUpdateResp),
					),
				},
				{
					// Checking Read(readAfter1stUpdateResp), Update and Read(readAfter2ndUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.GetDisplayName(), "", ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceName, readAfter2ndUpdateResp),
					),
				},
				{
					// Checking Read(readAfter2ndUpdateResp) -> no changes but tries to destroy with error
					Config:      fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.GetDisplayName(), "", ""),
					Destroy:     true,
					ExpectError: regexp.MustCompile("Cannot destroy instance"),
				},
				{
					// Checking Read(readAfter2ndUpdateResp), Update (del protection, no API call)
					// and final Read (readAfter2ndUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.GetDisplayName(), "", turnOffDelProtection),
				},
			},
		})
	})
})

func testAppSpaceResourceDataExists(n string, data *configpb.ApplicationSpace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID != data.GetId() {
			return errors.New("ID does not match")
		}
		keys := Keys{
			"id": Equal(data.GetId()),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"customer_id":         Equal(data.GetCustomerId()),
			"name":                Equal(data.GetName()),
			"display_name":        Equal(data.GetDisplayName()),
			"description":         Equal(data.GetDescription().GetValue()),
			"create_time":         Not(BeEmpty()),
			"update_time":         Not(BeEmpty()),
			"deletion_protection": Not(BeEmpty()),
			"region":              Equal(data.GetRegion()),
			"ikg_size":            Equal(data.GetIkgSize()),
			"replica_region":      Equal(data.GetReplicaRegion()),
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}
