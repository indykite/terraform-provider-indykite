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
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource Tenant", func() {
	const resourceName = "indykite_tenant.development"
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
			`resource "indykite_tenant" "development" {
				issuer_id = "` + issuerID + `"
				name = "acme"
				display_name = "%s"
				description = "%s"
				%s
			}`
		initialTenantResp := &configpb.Tenant{
			CustomerId:  customerID,
			AppSpaceId:  appSpaceID,
			IssuerId:    issuerID,
			Id:          tenantID,
			Name:        "acme",
			DisplayName: "acme",
			Description: wrapperspb.String("Just some Tenant description"),
			CreateTime:  timestamppb.Now(),
			UpdateTime:  timestamppb.Now(),
		}

		readAfter1stUpdateResp := &configpb.Tenant{
			CustomerId:  initialTenantResp.CustomerId,
			AppSpaceId:  initialTenantResp.AppSpaceId,
			IssuerId:    initialTenantResp.IssuerId,
			Id:          initialTenantResp.Id,
			Name:        "acme",
			DisplayName: "acme",
			Description: wrapperspb.String("Another Tenant description"),
			CreateTime:  initialTenantResp.CreateTime,
			UpdateTime:  timestamppb.Now(),
		}
		readAfter2ndUpdateResp := &configpb.Tenant{
			CustomerId:  initialTenantResp.CustomerId,
			AppSpaceId:  initialTenantResp.AppSpaceId,
			IssuerId:    initialTenantResp.IssuerId,
			Id:          initialTenantResp.Id,
			Name:        "acme",
			DisplayName: "Some new display name",
			Description: nil,
			CreateTime:  initialTenantResp.CreateTime,
			UpdateTime:  timestamppb.Now(),
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
			CreateTenant(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"IssuerId":    Equal(issuerID),
				"Name":        Equal(initialTenantResp.Name),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(initialTenantResp.Description.Value),
				})),
			})))).
			Return(&configpb.CreateTenantResponse{Id: initialTenantResp.Id}, nil)

		// 2x update
		mockConfigClient.EXPECT().
			UpdateTenant(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(initialTenantResp.Id),
				"DisplayName": BeNil(),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter1stUpdateResp.Description.Value),
				})),
			})))).
			Return(&configpb.UpdateTenantResponse{Id: initialTenantResp.Id}, nil)

		mockConfigClient.EXPECT().
			UpdateTenant(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialTenantResp.Id),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(readAfter2ndUpdateResp.DisplayName),
				})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal("")})),
			})))).
			Return(&configpb.UpdateTenantResponse{Id: initialTenantResp.Id}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadTenant(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialTenantResp.Id)})),
				})))).
				Times(4).
				Return(&configpb.ReadTenantResponse{Tenant: initialTenantResp}, nil),

			mockConfigClient.EXPECT().
				ReadTenant(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialTenantResp.Id)})),
				})))).
				Times(3).
				Return(&configpb.ReadTenantResponse{Tenant: readAfter1stUpdateResp}, nil),

			mockConfigClient.EXPECT().
				ReadTenant(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(initialTenantResp.Id)})),
				})))).
				Times(5).
				Return(&configpb.ReadTenantResponse{Tenant: readAfter2ndUpdateResp}, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteTenant(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(initialTenantResp.Id),
			})))).
			Return(&configpb.DeleteTenantResponse{}, nil)

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
					// Checking Create and Read (initialTenantResp)
					Config: fmt.Sprintf(tfConfigDef, "", initialTenantResp.Description.Value, ""),
					Check: resource.ComposeTestCheckFunc(
						testTenantResourceDataExists(resourceName, initialTenantResp),
					),
				},
				{
					// Performs 1 read (initialTenantResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: initialTenantResp.Id,
				},
				{
					// Checking Read (initialTenantResp), Update and Read(readAfter1stUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, "", readAfter1stUpdateResp.Description.Value, ""),
					Check: resource.ComposeTestCheckFunc(
						testTenantResourceDataExists(resourceName, readAfter1stUpdateResp),
					),
				},
				{
					// Checking Read(readAfter1stUpdateResp), Update and Read(readAfter2ndUpdateResp)
					Config: fmt.Sprintf(tfConfigDef, readAfter2ndUpdateResp.DisplayName, "", ""),
					Check: resource.ComposeTestCheckFunc(
						testTenantResourceDataExists(resourceName, readAfter2ndUpdateResp),
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

func testTenantResourceDataExists(n string, data *configpb.Tenant) resource.TestCheckFunc {
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
			"issuer_id":           Equal(data.IssuerId),
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
