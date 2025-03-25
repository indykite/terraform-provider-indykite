// Copyright (c) 2024 IndyKite
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
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource Consent config", func() {
	const resourceName = "indykite_consent.wonka"
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

	It("Test all CRUD", func() {
		consentConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				Id:          sampleID,
				Name:        "wonka-consent-config",
				DisplayName: "Wonka Consent for chocolate receipts",
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_ConsentConfig{
					ConsentConfig: &configpb.ConsentConfiguration{
						Purpose:        "To allow Wonka to store and process chocolate receipts",
						ApplicationId:  applicationID,
						RevokeAfterUse: true,
						ValidityPeriod: 86400,
						DataPoints:     []string{`{"returns": [{"properties": ["name", "location"]}]}`},
					},
				},
			},
		}

		consentInvalidResponse := proto.Clone(consentConfigResp).(*configpb.ReadConfigNodeResponse)
		consentInvalidResponse.ConfigNode.Config = &configpb.ConfigNode_EventSinkConfig{}

		consentConfigUpdateResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				Id:          sampleID,
				Name:        "wonka-consent-config",
				Description: wrapperspb.String("Description of the best Consent by Wonka inc."),
				CreateTime:  consentConfigResp.GetConfigNode().GetCreateTime(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_ConsentConfig{
					ConsentConfig: &configpb.ConsentConfiguration{
						Purpose:        "To allow Wonka to store and process the new chocolate sauce",
						ApplicationId:  applicationID,
						RevokeAfterUse: false,
						ValidityPeriod: 96400,
						DataPoints:     []string{`{"returns": [{"properties": ["name", "location"]}]}`},
					},
				},
			},
		}

		// Create
		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(consentConfigResp.GetConfigNode().GetName()),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					consentConfigResp.GetConfigNode().GetDisplayName(),
				)})),
				"Description": BeNil(),
				"Location":    Equal(appSpaceID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{"ConsentConfig": test.EqualProto(
					consentConfigResp.GetConfigNode().GetConsentConfig(),
				)})),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         consentConfigResp.GetConfigNode().GetId(),
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
			}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(consentConfigResp.GetConfigNode().GetId()),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					consentConfigUpdateResp.GetConfigNode().GetDescription().GetValue(),
				)})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"ConsentConfig": test.EqualProto(
						consentConfigUpdateResp.GetConfigNode().GetConsentConfig(),
					),
				})),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{
				Id: consentConfigResp.GetConfigNode().GetId(),
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(consentConfigResp.GetConfigNode().GetId()),
				})))).
				Times(2).
				Return(consentConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(consentConfigResp.GetConfigNode().GetId()),
				})))).
				Return(consentInvalidResponse, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(consentConfigResp.GetConfigNode().GetId()),
				})))).
				Return(consentConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(consentConfigResp.GetConfigNode().GetId()),
				})))).
				Times(2).
				Return(consentConfigUpdateResp, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(consentConfigResp.GetConfigNode().GetId()),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Error cases should be always first, easier to avoid missing mocks or incomplete plan
				{
					Config: `resource "indykite_consent" "wonka" {
						name = "wonka-consent-config"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "location" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_consent" "wonka" {
						location = "` + appSpaceID + `"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "name" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_consent" "wonka" {
						location = "` + appSpaceID + `"
						customer_id = "` + applicationID + `"
						name = "wonka-consent-config"

						purpose = "To allow Wonka to store and process chocolate receipts"
						application_id = "` + applicationID + `"
						validity_period = 86400
						revoke_after_use = true
						data_points = ["{\"returns\": [{\"properties\": [\"name\", \"location\"]}]}"]
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "customer_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_consent" "wonka" {
						location = "something-invalid"
						name = "wonka-consent-config"

						purpose = "To allow Wonka to store and process chocolate receipts"
						application_id = "` + applicationID + `"
						validity_period = 86400
						revoke_after_use = true
						data_points = ["{\"returns\": [{\"properties\": [\"name\", \"location\"]}]}"]
					}
					`,
					ExpectError: regexp.MustCompile(`expected to have 'gid:' prefix`),
				},
				{
					Config: `resource "indykite_consent" "wonka" {
						location = "` + appSpaceID + `"
						name = "Invalid Name @#$"

						purpose = "To allow Wonka to store and process chocolate receipts"
						application_id = "` + applicationID + `"
						validity_period = 86400
						revoke_after_use = true
						data_points = ["{\"returns\": [{\"properties\": [\"name\", \"location\"]}]}"]
					}
					`,
					ExpectError: regexp.MustCompile(`Value can have lowercase letters, digits, or hyphens.`),
				},
				{
					Config: `resource "indykite_consent" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-consent-config"
					}
					`,
					ExpectError: regexp.MustCompile(`argument "application_id" is required`),
				},
				{
					Config: `resource "indykite_consent" "wonka" {
						location = ""
						name = "wonka-consent-config"

						purpose = "To allow Wonka to store and process chocolate receipts"
						application_id = "` + applicationID + `"
						validity_period = 86400
						revoke_after_use = true
						data_points = ["{\"returns\": [{\"properties\": [\"name\", \"location\"]}]}"]
					}
					`,
					ExpectError: regexp.MustCompile(`expected to have 'gid:' prefix`),
				},
				{
					Config: `resource "indykite_consent" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-consent-config"
						display_name = "Wonka Consent for chocolate receipts"

						purpose = "To allow Wonka to store and process chocolate receipts"
						application_id = "` + applicationID + `"
						validity_period = 86400
						revoke_after_use = true
						data_points = ["\"returns\": [{\"properties\": [\"name\", \"location\"]}]"]
					}`,
					//nolint:lll
					ExpectError: regexp.MustCompile(`contains an invalid JSON: invalid character ':' after top-level value`),
				},
				// ---- Run mocked tests here ----
				// Minimal config - Checking Create and Read (consentConfigResp)
				{
					Config: `resource "indykite_consent" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-consent-config"
						display_name = "Wonka Consent for chocolate receipts"

						purpose = "To allow Wonka to store and process chocolate receipts"
						application_id = "` + applicationID + `"
						validity_period = 86400
						revoke_after_use = true
						data_points = ["{\"returns\": [{\"properties\": [\"name\", \"location\"]}]}"]
					}`,

					Check: resource.ComposeTestCheckFunc(testConsentResourceDataExists(
						resourceName,
						consentConfigResp,
						Keys{
							"data_points.#": Equal(strconv.Itoa(len(consentConfigResp.GetConfigNode().
								GetConsentConfig().GetDataPoints()))),
							"data_points.0": Equal(consentConfigResp.GetConfigNode().GetConsentConfig().
								GetDataPoints()[0]),
						},
					)),
				},
				{
					// Performs 1 read (consentConfigResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: consentConfigResp.GetConfigNode().GetId(),
				},
				// Checking Read(consentConfigResp), Update and Read(consentConfigUpdateResp)
				{
					Config: `resource "indykite_consent" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-consent-config"
						description = "Description of the best Consent by Wonka inc."

						purpose = "To allow Wonka to store and process the new chocolate sauce"
						application_id = "` + applicationID + `"
						validity_period = 96400
						revoke_after_use = false
						data_points = ["{\"returns\": [{\"properties\": [\"name\", \"location\"]}]}"]
					}
					`,
					Check: resource.ComposeTestCheckFunc(testConsentResourceDataExists(
						resourceName,
						consentConfigUpdateResp,
						Keys{
							"data_points.#": Equal(strconv.Itoa(len(consentConfigUpdateResp.GetConfigNode().
								GetConsentConfig().GetDataPoints()))),
							"data_points.0": Equal(consentConfigUpdateResp.GetConfigNode().GetConsentConfig().
								GetDataPoints()[0]),
						},
					)),
				},
			},
		})
	})
})

func testConsentResourceDataExists(
	n string,
	data *configpb.ReadConfigNodeResponse,
	extraKeys Keys,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != data.GetConfigNode().GetId() {
			return errors.New("ID does not match")
		}
		attrs := rs.Primary.Attributes

		expectedJSON := data.GetConfigNode().GetConsentConfig()

		keys := Keys{
			"id": Equal(data.GetConfigNode().GetId()),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"location":         Not(BeEmpty()), // Response does not return this
			"customer_id":      Equal(data.GetConfigNode().GetCustomerId()),
			"app_space_id":     Equal(data.GetConfigNode().GetAppSpaceId()),
			"name":             Equal(data.GetConfigNode().GetName()),
			"display_name":     Equal(data.GetConfigNode().GetDisplayName()),
			"description":      Equal(data.GetConfigNode().GetDescription().GetValue()),
			"create_time":      Not(BeEmpty()),
			"update_time":      Not(BeEmpty()),
			"purpose":          Equal(expectedJSON.GetPurpose()),
			"application_id":   Equal(expectedJSON.GetApplicationId()),
			"validity_period":  Equal(strconv.FormatUint(expectedJSON.GetValidityPeriod(), 10)),
			"revoke_after_use": Equal(strconv.FormatBool(expectedJSON.GetRevokeAfterUse())),
		}

		for k, v := range extraKeys {
			keys[k] = v
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), attrs)
	}
}
