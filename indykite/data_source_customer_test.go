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
	"regexp"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"
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

var _ = Describe("Data Source customer", func() {
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

	It("Test Read Customer", func() {
		wonka := &configpb.ReadCustomerResponse{
			Customer: &configpb.Customer{
				Id:          customerID,
				Name:        "wonka",
				DisplayName: "wonka",
				Description: wrapperspb.String("Just some description"),
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
			},
		}

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadCustomer(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Name": Equal("acme")})),
				})))).
				Return(nil, status.Error(codes.NotFound, "unknown name")),

			mockConfigClient.EXPECT().
				ReadCustomer(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(customerID)})),
				})))).
				Times(5).
				Return(wonka, nil),

			mockConfigClient.EXPECT().
				ReadCustomer(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Name": Equal("wonka")})),
				})))).
				Times(5).
				Return(wonka, nil),
		)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
			},
			Steps: []resource.TestStep{
				// Error cases must be always first
				{
					Config: `data "indykite_customer" "error" {
						name = "wonka"
						customer_id = "` + customerID + `"
					}`,
					ExpectError: regexp.MustCompile("only one of `customer_id,name` can be specified"),
				},
				{
					Config:      `data "indykite_customer" "error" {}`,
					ExpectError: regexp.MustCompile("one of `customer_id,name` must be specified"),
				},
				{
					Config: `data "indykite_customer" "error" {
						name = "wonka-"
					}`,
					ExpectError: regexp.MustCompile("Value can have lowercase letters, digits, or hyphens"),
				},
				{
					Config: `data "indykite_customer" "error" {
						customer_id = "gid:not-valid-base64@#$"
					}`,
					ExpectError: regexp.MustCompile("valid Raw URL Base64 string with 'gid:' prefix"),
				},
				{
					Config:      `data "indykite_customer" "wonka" {name = "acme"}`,
					ExpectError: regexp.MustCompile("unknown name"),
				},

				// Success cases
				{
					Check: resource.ComposeTestCheckFunc(testDataSourceWonkaCustomer(wonka.Customer, customerID)),
					Config: `data "indykite_customer" "wonka" {
						customer_id = "` + customerID + `"
					}`,
				},
				{
					Check: resource.ComposeTestCheckFunc(testDataSourceWonkaCustomer(wonka.Customer, "")),
					Config: `data "indykite_customer" "wonka" {
						name = "wonka"
					}`,
				},
			},
		})
	})
})

func testDataSourceWonkaCustomer(data *configpb.Customer, customerID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["data.indykite_customer.wonka"]
		if !ok {
			return errors.New("not found: `indykite_customer.wonka`")
		}

		if rs.Primary.ID != data.Id {
			return errors.New("ID does not match")
		}

		keys := Keys{
			"id": Equal(data.Id),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"name":         Equal(data.Name),
			"display_name": Equal(data.DisplayName),
			"description":  Equal(data.Description.GetValue()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
		}
		if customerID != "" {
			keys["customer_id"] = Equal(customerID)
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}
