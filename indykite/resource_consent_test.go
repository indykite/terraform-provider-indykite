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
		mockedBookmark   string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)

		// Bookmark must be longer than 40 chars - have just 1 added before the first write to test all cases
		mockedBookmark = "for-consent" + uuid.NewRandom().String()
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
		consentInvalidResponse.ConfigNode.Config = &configpb.ConfigNode_AuditSinkConfig{}

		consentConfigUpdateResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				Id:          sampleID,
				Name:        "wonka-consent-config",
				Description: wrapperspb.String("Description of the best Consent by Wonka inc."),
				CreateTime:  consentConfigResp.ConfigNode.CreateTime,
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

		createBM := "created-consent" + uuid.NewRandom().String()
		updateBM := "updated-consent" + uuid.NewRandom().String()
		deleteBM := "deleted-consent" + uuid.NewRandom().String()

		// Create
		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(consentConfigResp.ConfigNode.Name),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					consentConfigResp.ConfigNode.DisplayName,
				)})),
				"Description": BeNil(),
				"Location":    Equal(appSpaceID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{"ConsentConfig": test.EqualProto(
					consentConfigResp.ConfigNode.GetConsentConfig(),
				)})),
				"Bookmarks": ConsistOf(mockedBookmark),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         consentConfigResp.ConfigNode.Id,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
				Bookmark:   createBM,
			}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(consentConfigResp.ConfigNode.Id),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					consentConfigUpdateResp.ConfigNode.Description.GetValue(),
				)})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"ConsentConfig": test.EqualProto(
						consentConfigUpdateResp.ConfigNode.GetConsentConfig(),
					),
				})),
				"Bookmarks": ConsistOf(mockedBookmark, createBM),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{
				Id:       consentConfigResp.ConfigNode.Id,
				Bookmark: updateBM,
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(consentConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM),
				})))).
				Times(2).
				Return(consentConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(consentConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM),
				})))).
				Return(consentInvalidResponse, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(consentConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM),
				})))).
				Return(consentConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(consentConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM),
				})))).
				Times(2).
				Return(consentConfigUpdateResp, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":        Equal(consentConfigResp.ConfigNode.Id),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{
				Bookmark: deleteBM,
			}, nil)

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
					ImportStateId: consentConfigResp.ConfigNode.Id,
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
		if rs.Primary.ID != data.GetConfigNode().Id {
			return errors.New("ID does not match")
		}
		attrs := rs.Primary.Attributes

		expectedJSON := data.GetConfigNode().GetConsentConfig()

		keys := Keys{
			"id": Equal(data.GetConfigNode().Id),
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
