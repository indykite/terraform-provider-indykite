// Copyright (c) 2023 IndyKite
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

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"
	"github.com/pborman/uuid"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource Customer/AppSpace/Tenant configuration", func() {
	const (
		customerResourceName = "indykite_customer_configuration.development"
		appSpaceResourceName = "indykite_application_space_configuration.development"
		tenantResourceName   = "indykite_tenant_configuration.development"
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
		mockedBookmark = "for-container-cfg" + uuid.NewRandom().String()
		bmOnce := &sync.Once{}

		provider = indykite.Provider()
		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc =
			func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
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

	It("Test CRU of Customer configuration", func() {
		// Terraform created config must be in sync with data in expectedCustomerResp and expectedUpdatedCustomerResp
		// otherwise "After applying this test step, the plan was not empty" error is thrown
		tfConfigDef :=
			`resource "indykite_customer_configuration" "development" {
				customer_id = "%s"
				%s
			}`
		expectedCustomerResp := &configpb.ReadCustomerConfigResponse{
			Id:     customerID,
			Config: &configpb.CustomerConfig{},
		}
		expectedUpdatedCustomerResp := &configpb.ReadCustomerConfigResponse{
			Id: customerID,
			Config: &configpb.CustomerConfig{
				DefaultAuthFlowId:     sampleID,
				DefaultEmailServiceId: sampleID2,
			},
		}

		createUpdateBM := "created-updated-customer-cfg" + uuid.NewRandom().String()
		createUpdateBM2 := "created-updated-customer-cfg-2" + uuid.NewRandom().String()

		// Create/Update
		mockConfigClient.EXPECT().
			UpdateCustomerConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(&configpb.UpdateCustomerConfigRequest{
				Id:        customerID,
				Config:    expectedCustomerResp.Config,
				Bookmarks: []string{mockedBookmark},
			}))).
			Return(&configpb.UpdateCustomerConfigResponse{
				Id:       customerID,
				Bookmark: createUpdateBM,
			}, nil)

		mockConfigClient.EXPECT().
			UpdateCustomerConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(&configpb.UpdateCustomerConfigRequest{
				Id:        customerID,
				Config:    expectedUpdatedCustomerResp.Config,
				Bookmarks: []string{mockedBookmark, createUpdateBM},
			}))).
			Return(&configpb.UpdateCustomerConfigResponse{
				Id:       customerID,
				Bookmark: createUpdateBM2,
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadCustomerConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(&configpb.ReadCustomerConfigRequest{
					Id:        customerID,
					Bookmarks: []string{mockedBookmark, createUpdateBM},
				}))).
				Times(3).
				Return(expectedCustomerResp, nil),

			mockConfigClient.EXPECT().
				ReadCustomerConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(&configpb.ReadCustomerConfigRequest{
					Id:        customerID,
					Bookmarks: []string{mockedBookmark, createUpdateBM, createUpdateBM2},
				}))).
				Times(2).
				Return(expectedUpdatedCustomerResp, nil),
		)

		testResourceDataExists := func(n string, data *configpb.ReadCustomerConfigResponse) resource.TestCheckFunc {
			return func(s *terraform.State) error {
				rs, ok := s.RootModule().Resources[n]
				if !ok {
					return fmt.Errorf("not found: %s", n)
				}

				if rs.Primary.ID != "container:"+data.Id {
					return errors.New("ID does not match")
				}

				keys := Keys{
					"id":          Equal("container:" + data.Id),
					"customer_id": Equal(data.Id),
					"%":           Not(BeEmpty()), // This is Terraform helper

					"default_auth_flow_id":     Equal(data.Config.DefaultAuthFlowId),
					"default_email_service_id": Equal(data.Config.DefaultEmailServiceId),
				}

				return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
			}
		}

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors case must be always first
				{
					Config:      fmt.Sprintf(tfConfigDef, "ccc", ""),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, customerID, `default_auth_flow_id = "ccc"`),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, customerID, `default_email_service_id = "ccc"`),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					// Performs 1 read (appAgentJWKCredResp)
					ResourceName:  customerResourceName,
					ImportState:   true,
					ImportStateId: customerID,
					ExpectError:   regexp.MustCompile(`indykite_customer_configuration doesn't support import`),
				},
				{
					// Checking Create and Read (appAgentJWKCredResp)
					Config: fmt.Sprintf(tfConfigDef, customerID, ""),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(customerResourceName, expectedCustomerResp),
					),
				},
				{
					// Checking Create and Read (appAgentJWKCredResp)
					Config: fmt.Sprintf(tfConfigDef, customerID, `default_auth_flow_id = "`+sampleID+`"
					default_email_service_id = "`+sampleID2+`"`),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(customerResourceName, expectedUpdatedCustomerResp),
					),
				},
			},
		})
	})

	It("Test CRU of ApplicationSpace configuration", func() {
		// Terraform created config must be in sync with data in expectedAppSpaceResp and expectedUpdatedAppSpaceResp
		// otherwise "After applying this test step, the plan was not empty" error is thrown
		tfConfigDef :=
			`resource "indykite_application_space_configuration" "development" {
				app_space_id = "%s"
				%s
			}`
		expectedAppSpaceResp := &configpb.ReadApplicationSpaceConfigResponse{
			Id:     appSpaceID,
			Config: &configpb.ApplicationSpaceConfig{},
		}
		expectedUpdatedAppSpaceResp := &configpb.ReadApplicationSpaceConfigResponse{
			Id: appSpaceID,
			Config: &configpb.ApplicationSpaceConfig{
				DefaultAuthFlowId:     sampleID,
				DefaultEmailServiceId: sampleID2,
				DefaultTenantId:       tenantID,
			},
		}

		createUpdateBM := "created-updated-app-space-cfg" + uuid.NewRandom().String()
		createUpdateBM2 := "created-updated-app-space-cfg-2" + uuid.NewRandom().String()

		// Create/Update
		mockConfigClient.EXPECT().
			UpdateApplicationSpaceConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(
				&configpb.UpdateApplicationSpaceConfigRequest{
					Id:        appSpaceID,
					Config:    expectedAppSpaceResp.Config,
					Bookmarks: []string{mockedBookmark},
				},
			))).
			Return(&configpb.UpdateApplicationSpaceConfigResponse{
				Id:       appSpaceID,
				Bookmark: createUpdateBM,
			}, nil)

		mockConfigClient.EXPECT().
			UpdateApplicationSpaceConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(
				&configpb.UpdateApplicationSpaceConfigRequest{
					Id:        appSpaceID,
					Config:    expectedUpdatedAppSpaceResp.Config,
					Bookmarks: []string{mockedBookmark, createUpdateBM},
				},
			))).
			Return(&configpb.UpdateApplicationSpaceConfigResponse{
				Id:       appSpaceID,
				Bookmark: createUpdateBM2,
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadApplicationSpaceConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(
					&configpb.ReadApplicationSpaceConfigRequest{
						Id:        appSpaceID,
						Bookmarks: []string{mockedBookmark, createUpdateBM},
					},
				))).
				Times(3).
				Return(expectedAppSpaceResp, nil),

			mockConfigClient.EXPECT().
				ReadApplicationSpaceConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(
					&configpb.ReadApplicationSpaceConfigRequest{
						Id:        appSpaceID,
						Bookmarks: []string{mockedBookmark, createUpdateBM, createUpdateBM2},
					},
				))).
				Times(2).
				Return(expectedUpdatedAppSpaceResp, nil),
		)

		testDataExists := func(n string, data *configpb.ReadApplicationSpaceConfigResponse) resource.TestCheckFunc {
			return func(s *terraform.State) error {
				rs, ok := s.RootModule().Resources[n]
				if !ok {
					return fmt.Errorf("not found: %s", n)
				}

				if rs.Primary.ID != "container:"+data.Id {
					return errors.New("ID does not match")
				}

				keys := Keys{
					"id":           Equal("container:" + data.Id),
					"app_space_id": Equal(data.Id),
					"%":            Not(BeEmpty()), // This is Terraform helper

					"default_auth_flow_id":     Equal(data.Config.DefaultAuthFlowId),
					"default_email_service_id": Equal(data.Config.DefaultEmailServiceId),
					"default_tenant_id":        Equal(data.Config.DefaultTenantId),
				}

				return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
			}
		}

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors case must be always first
				{
					Config:      fmt.Sprintf(tfConfigDef, "ccc", ""),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, `default_auth_flow_id = "ccc"`),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, `default_email_service_id = "ccc"`),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, `default_tenant_id = "ccc"`),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					// Performs 1 read (appAgentJWKCredResp)
					ResourceName:  appSpaceResourceName,
					ImportState:   true,
					ImportStateId: appSpaceID,

					ExpectError: regexp.MustCompile(`indykite_application_space_configuration doesn't support import`),
				},
				{
					// Checking Create and Read (appAgentJWKCredResp)
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, ""),
					Check: resource.ComposeTestCheckFunc(
						testDataExists(appSpaceResourceName, expectedAppSpaceResp),
					),
				},
				{
					// Checking Create and Read (appAgentJWKCredResp)
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, `default_auth_flow_id = "`+sampleID+`"
					default_email_service_id = "`+sampleID2+`"
					default_tenant_id = "`+tenantID+`"`),
					Check: resource.ComposeTestCheckFunc(
						testDataExists(appSpaceResourceName, expectedUpdatedAppSpaceResp),
					),
				},
			},
		})
	})

	It("Test CRU of Tenant configuration", func() {
		// Terraform created config must be in sync with data in expectedAppSpaceResp and expectedUpdatedAppSpaceResp
		// otherwise "After applying this test step, the plan was not empty" error is thrown
		tfConfigDef :=
			`resource "indykite_tenant_configuration" "development" {
				tenant_id = "%s"
				%s
			}`
		expectedAppSpaceResp := &configpb.ReadTenantConfigResponse{
			Id:     appSpaceID,
			Config: &configpb.TenantConfig{},
		}
		expectedUpdatedAppSpaceResp := &configpb.ReadTenantConfigResponse{
			Id: appSpaceID,
			Config: &configpb.TenantConfig{
				DefaultAuthFlowId:     sampleID,
				DefaultEmailServiceId: sampleID2,
			},
		}

		createUpdateBM := "created-updated-tenant-cfg" + uuid.NewRandom().String()
		createUpdateBM2 := "created-updated-tenant-cfg-2" + uuid.NewRandom().String()

		// Create/Update
		mockConfigClient.EXPECT().
			UpdateTenantConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(&configpb.UpdateTenantConfigRequest{
				Id:        appSpaceID,
				Config:    expectedAppSpaceResp.Config,
				Bookmarks: []string{mockedBookmark},
			}))).
			Return(&configpb.UpdateTenantConfigResponse{
				Id:       appSpaceID,
				Bookmark: createUpdateBM,
			}, nil)

		mockConfigClient.EXPECT().
			UpdateTenantConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(&configpb.UpdateTenantConfigRequest{
				Id:        appSpaceID,
				Config:    expectedUpdatedAppSpaceResp.Config,
				Bookmarks: []string{mockedBookmark, createUpdateBM},
			}))).
			Return(&configpb.UpdateTenantConfigResponse{
				Id:       appSpaceID,
				Bookmark: createUpdateBM2,
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadTenantConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(
					&configpb.ReadTenantConfigRequest{
						Id:        appSpaceID,
						Bookmarks: []string{mockedBookmark, createUpdateBM},
					},
				))).
				Times(3).
				Return(expectedAppSpaceResp, nil),

			mockConfigClient.EXPECT().
				ReadTenantConfig(gomock.Any(), test.WrapMatcher(test.EqualProto(
					&configpb.ReadTenantConfigRequest{
						Id:        appSpaceID,
						Bookmarks: []string{mockedBookmark, createUpdateBM, createUpdateBM2},
					},
				))).
				Times(2).
				Return(expectedUpdatedAppSpaceResp, nil),
		)

		testDataExists := func(n string, data *configpb.ReadTenantConfigResponse) resource.TestCheckFunc {
			return func(s *terraform.State) error {
				rs, ok := s.RootModule().Resources[n]
				if !ok {
					return fmt.Errorf("not found: %s", n)
				}

				if rs.Primary.ID != "container:"+data.Id {
					return errors.New("ID does not match")
				}

				keys := Keys{
					"id":        Equal("container:" + data.Id),
					"tenant_id": Equal(data.Id),
					"%":         Not(BeEmpty()), // This is Terraform helper

					"default_auth_flow_id":     Equal(data.Config.DefaultAuthFlowId),
					"default_email_service_id": Equal(data.Config.DefaultEmailServiceId),
				}

				return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
			}
		}

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors case must be always first
				{
					Config:      fmt.Sprintf(tfConfigDef, "ccc", ""),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, `default_auth_flow_id = "ccc"`),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, `default_email_service_id = "ccc"`),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					// Performs 1 read (appAgentJWKCredResp)
					ResourceName:  tenantResourceName,
					ImportState:   true,
					ImportStateId: appSpaceID,

					ExpectError: regexp.MustCompile(`indykite_tenant_configuration doesn't support import`),
				},
				{
					// Checking Create and Read (appAgentJWKCredResp)
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, ""),
					Check: resource.ComposeTestCheckFunc(
						testDataExists(tenantResourceName, expectedAppSpaceResp),
					),
				},
				{
					// Checking Create and Read (appAgentJWKCredResp)
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, `default_auth_flow_id = "`+sampleID+`"
					default_email_service_id = "`+sampleID2+`"`),
					Check: resource.ComposeTestCheckFunc(
						testDataExists(tenantResourceName, expectedUpdatedAppSpaceResp),
					),
				},
			},
		})
	})
})
