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

var _ = Describe("DataSource Application Space", func() {
	const resourceName = "data.indykite_application_space.development"
	var (
		mockCtrl                *gomock.Controller
		mockConfigClient        *configm.MockConfigManagementAPIClient
		mockListAppSpacesClient *configm.MockConfigManagementAPI_ListApplicationSpacesClient
		indykiteProviderFactory func() (*schema.Provider, error)
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)
		mockListAppSpacesClient = configm.NewMockConfigManagementAPI_ListApplicationSpacesClient(mockCtrl)

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
		appSpaceResp := &configpb.ApplicationSpace{
			CustomerId:  customerID,
			Id:          appSpaceID,
			IssuerId:    issuerID,
			Name:        "acme",
			DisplayName: "Some Cool Display name",
			Description: wrapperspb.String("Just some AppSpace description"),
			CreateTime:  timestamppb.Now(),
			UpdateTime:  timestamppb.Now(),
		}

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{
						"Name": PointTo(MatchFields(IgnoreExtras, Fields{
							"Name":     Equal(appSpaceResp.Name),
							"Location": Equal(customerID),
						})),
					})),
				})))).
				Return(nil, status.Error(codes.NotFound, "unknown name")),

			mockConfigClient.EXPECT().
				ReadApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{"Id": Equal(appSpaceID)})),
				})))).
				Times(5).
				Return(&configpb.ReadApplicationSpaceResponse{AppSpace: appSpaceResp}, nil),

			mockConfigClient.EXPECT().
				ReadApplicationSpace(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Identifier": PointTo(MatchFields(IgnoreExtras, Fields{
						"Name": PointTo(MatchFields(IgnoreExtras, Fields{
							"Name":     Equal(appSpaceResp.Name),
							"Location": Equal(customerID),
						})),
					})),
				})))).
				Times(5).
				Return(&configpb.ReadApplicationSpaceResponse{AppSpace: appSpaceResp}, nil),
		)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					Config: `data "indykite_application_space" "development" {
						name = "acme"
						app_space_id = "` + appSpaceID + `"
					}`,
					ExpectError: regexp.MustCompile("only one of `app_space_id,name` can be specified"),
				},
				{
					Config: `data "indykite_application_space" "development" {
						display_name = "anything"
					}`,
					ExpectError: regexp.MustCompile("one of `app_space_id,name` must be specified"),
				},
				{
					Config: `data "indykite_application_space" "development" {
						name = "anything"
					}`,
					ExpectError: regexp.MustCompile("\"name\": all of `customer_id,name` must be specified"),
				},
				{
					Config: `data "indykite_application_space" "development" {
						customer_id = "` + customerID + `"
						name = "acme"
					}`,
					ExpectError: regexp.MustCompile("unknown name"),
				},

				// Success test cases
				{
					Check: resource.ComposeTestCheckFunc(testAppSpaceDataExists(resourceName, appSpaceResp)),
					Config: `data "indykite_application_space" "development" {
						app_space_id = "` + appSpaceID + `"
					}`,
				},
				{
					Check: resource.ComposeTestCheckFunc(testAppSpaceDataExists(resourceName, appSpaceResp)),
					Config: `data "indykite_application_space" "development" {
						customer_id = "` + customerID + `"
						name = "acme"
					}`,
				},
			},
		})
	})

	It("Test list by multple names", func() {
		appSpaceResp := &configpb.ApplicationSpace{
			CustomerId:  customerID,
			Id:          appSpaceID,
			IssuerId:    issuerID,
			Name:        "acme",
			DisplayName: "Some Cool Display name",
			Description: wrapperspb.String("Just some AppSpace description"),
			CreateTime:  timestamppb.Now(),
			UpdateTime:  timestamppb.Now(),
		}
		appSpaceResp2 := &configpb.ApplicationSpace{
			CustomerId: customerID,
			Id:         sampleID,
			IssuerId:   issuerID,
			Name:       "wonka",
			CreateTime: timestamppb.Now(),
			UpdateTime: timestamppb.Now(),
		}

		mockConfigClient.EXPECT().
			ListApplicationSpaces(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Match":      ConsistOf("acme", "some-another-name", "wonka"),
				"CustomerId": Equal(customerID),
			})))).
			Times(5).
			DoAndReturn(
				func(
					_, _ interface{},
					_ ...interface{},
				) (*configm.MockConfigManagementAPI_ListApplicationSpacesClient, error) {
					mockListAppSpacesClient.EXPECT().Recv().
						Return(&configpb.ListApplicationSpacesResponse{AppSpace: appSpaceResp}, nil)
					mockListAppSpacesClient.EXPECT().Recv().
						Return(&configpb.ListApplicationSpacesResponse{AppSpace: appSpaceResp2}, nil)
					mockListAppSpacesClient.EXPECT().Recv().Return(nil, io.EOF)
					return mockListAppSpacesClient, nil
				},
			)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
			},
			Steps: []resource.TestStep{
				// Errors cases must be always first
				{
					Config: `data "indykite_application_spaces" "development" {
						filter = "acme"
						customer_id = "` + customerID + `"
					}`,
					ExpectError: regexp.MustCompile(
						`Inappropriate value for attribute "filter": list of string required`),
				},
				{
					Config: `data "indykite_application_spaces" "development" {
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile(`The argument "customer_id" is required`),
				},
				{
					Config: `data "indykite_application_spaces" "development" {
						filter = []
						customer_id = "` + customerID + `"
					}`,
					ExpectError: regexp.MustCompile("Attribute filter requires 1 item minimum, but config has only 0"),
				},
				{
					Config: `data "indykite_application_spaces" "development" {
						customer_id = "` + customerID + `"
						filter = [123]
					}`,
					ExpectError: regexp.MustCompile("Value can have lowercase letters, digits, or hyphens."),
				},
				{
					Config: `data "indykite_application_spaces" "development" {
						customer_id = "abc"
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile("expected to have 'gid:' prefix"),
				},
				{
					Config: `data "indykite_application_spaces" "development" {
						customer_id = "` + customerID + `"
						filter = ["acme"]
						app_spaces = []
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "app_spaces":`),
				},

				// Success test cases
				{
					Check: resource.ComposeTestCheckFunc(testAppSpaceListDataExists(
						"data.indykite_application_spaces.development",
						appSpaceResp,
						appSpaceResp2)),
					Config: `data "indykite_application_spaces" "development" {
						customer_id = "` + customerID + `"
						filter = ["acme", "some-another-name", "wonka"]
					}`,
				},
			},
		})
	})
})

func testAppSpaceDataExists(n string, data *configpb.ApplicationSpace) resource.TestCheckFunc {
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
		if v, has := rs.Primary.Attributes["name"]; !has || v != data.Name {
			return fmt.Errorf("invalid name: %s", v)
		}
		if v, has := rs.Primary.Attributes["issuer_id"]; !has || v != data.IssuerId {
			return fmt.Errorf("invalid issuer_id: %s", v)
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

func testAppSpaceListDataExists(n string, data ...*configpb.ApplicationSpace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		attrs := rs.Primary.Attributes

		expectedID := "gid:AAAAAWluZHlraURlgAAAAAAAAA8/appSpaces/acme,some-another-name,wonka"
		if rs.Primary.ID != expectedID {
			return fmt.Errorf("expected ID to be '%s' got '%s'", expectedID, rs.Primary.ID)
		}

		if attrs["app_spaces.#"] != strconv.Itoa(len(data)) {
			return fmt.Errorf("expected %d app_spaces, but got %s", len(data), attrs["app_spaces.#"])
		}
		for i, d := range data {
			k := "app_spaces." + strconv.Itoa(i) + "."
			if v, has := attrs[k+"id"]; !has || v != d.Id {
				return fmt.Errorf("%d. entry for 'id' expected '%s', got '%s'", i, d.Id, v)
			}
			if v, has := attrs[k+"customer_id"]; !has || v != d.CustomerId {
				return fmt.Errorf("%d. entry for 'customer_id' expected '%s', got '%s'", i, d.CustomerId, v)
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
			if v, has := attrs[k+"issuer_id"]; !has || v != d.IssuerId {
				return fmt.Errorf("%d. entry for 'issuer_id' expected '%s', got '%s'", i, d.IssuerId, v)
			}
		}
		return nil
	}
}
