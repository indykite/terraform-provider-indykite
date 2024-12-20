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
	"reflect"
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
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource ExternalDataResolver", func() {
	const (
		resourceName = "indykite_external_data_resolver.development"
	)
	var (
		mockCtrl         *gomock.Controller
		mockConfigClient *configm.MockConfigManagementAPIClient
		provider         *schema.Provider
		tfConfigDef      string
		validSettings    string
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

		tfConfigDef = `resource "indykite_external_data_resolver" "development" {
			location = "%s"
			name = "%s"
			%s
		}`

		validSettings = `
		url = "https://example.com/source2"
		method = "GET"

		headers {
		  name   = "Authorization"
		  values = ["Bearer edolkUTY"]
		}

		headers {
		  name   = "Content-Type"
		  values = ["application/json"]
		}

		request_type = "json"
		request_payload  = "{\"key\": \"value\"}"
		response_type = "json"
		response_selector = "."
		`
	})

	Describe("Error cases", func() {
		It("should handle invalid configurations", func() {
			resource.Test(GinkgoT(), resource.TestCase{
				Providers: map[string]*schema.Provider{
					"indykite": provider,
				},
				Steps: []resource.TestStep{
					{
						Config:      fmt.Sprintf(tfConfigDef, "ccc", "name", validSettings),
						ExpectError: regexp.MustCompile("Invalid ID value"),
					},
					{
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name",
							`url = "https://example.com/source2"

							headers {
							  name   = "Authorization"
							  values = ["Bearer edolkUTY"]
							}

							headers {
							  name   = "Content-Type"
							  values = ["application/json"]
							}

							request_type = "json"
							response_type = "json"
							response_selector = "."
							`,
						),
						ExpectError: regexp.MustCompile(
							`The argument "method" is required, but no definition was found`),
					},
					{
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name",
							`url = "https://example.com/source2"
							method = "GET"
							headers {
							  name   = "Content-Type"
							  values = []
							}
							request_type = "json"
							response_type = "json"
							response_selector = "."
							`),
						ExpectError: regexp.MustCompile(
							`Attribute headers.0.values requires 1 item minimum, but config has only 0`),
					},
				},
			})
		})
	})

	Describe("Valid configurations", func() {
		It("Test CRUD of ExternalDataResolver configuration", func() {
			expectedResp := &configpb.ReadConfigNodeResponse{
				ConfigNode: &configpb.ConfigNode{
					Id:          sampleID,
					Name:        "my-first-external-data-resolver1",
					DisplayName: "Display name of ExternalDataResolver configuration",
					CustomerId:  customerID,
					AppSpaceId:  appSpaceID,
					CreateTime:  timestamppb.Now(),
					UpdateTime:  timestamppb.Now(),
					Config: &configpb.ConfigNode_ExternalDataResolverConfig{
						ExternalDataResolverConfig: &configpb.ExternalDataResolverConfig{
							Url:    "https://example.com/source2",
							Method: "GET",
							Headers: map[string]*configpb.ExternalDataResolverConfig_Header{
								"Authorization": {Values: []string{"Bearer edolkUTY"}},
								"Content-Type":  {Values: []string{"application/json"}},
							},
							RequestType:      configpb.ExternalDataResolverConfig_CONTENT_TYPE_JSON,
							ResponseType:     configpb.ExternalDataResolverConfig_CONTENT_TYPE_JSON,
							RequestPayload:   []byte(`{"key": "value"}`),
							ResponseSelector: ".",
						},
					},
				},
			}
			expectedUpdatedResp := &configpb.ReadConfigNodeResponse{
				ConfigNode: &configpb.ConfigNode{
					Id:          sampleID,
					Name:        "my-first-external-data-resolver1",
					DisplayName: "Display name of ExternalDataResolver configuration",
					CustomerId:  customerID,
					AppSpaceId:  appSpaceID,
					CreateTime:  expectedResp.ConfigNode.CreateTime,
					UpdateTime:  timestamppb.Now(),
					Config: &configpb.ConfigNode_ExternalDataResolverConfig{
						ExternalDataResolverConfig: &configpb.ExternalDataResolverConfig{
							Url:    "https://example.com/source2",
							Method: "GET",
							Headers: map[string]*configpb.ExternalDataResolverConfig_Header{
								"Authorization": {Values: []string{"Bearer edolkUTY"}},
								"Content-Type":  {Values: []string{"application/json"}},
							},
							RequestType:      configpb.ExternalDataResolverConfig_CONTENT_TYPE_JSON,
							ResponseType:     configpb.ExternalDataResolverConfig_CONTENT_TYPE_JSON,
							RequestPayload:   []byte(`{"key2": "value2"}`),
							ResponseSelector: ".",
						},
					},
				},
			}
			expectedUpdatedResp2 := &configpb.ReadConfigNodeResponse{
				ConfigNode: &configpb.ConfigNode{
					Id:         sampleID,
					Name:       "my-first-external-data-resolver1",
					CustomerId: customerID,
					AppSpaceId: appSpaceID,
					CreateTime: expectedResp.ConfigNode.CreateTime,
					UpdateTime: timestamppb.Now(),
					Config: &configpb.ConfigNode_ExternalDataResolverConfig{
						ExternalDataResolverConfig: &configpb.ExternalDataResolverConfig{
							Url:    "https://example.com/source2",
							Method: "GET",
							Headers: map[string]*configpb.ExternalDataResolverConfig_Header{
								"Authorization": {Values: []string{"Bearer edolkUTY"}},
								"Content-Type":  {Values: []string{"application/json"}},
							},
							RequestType:      configpb.ExternalDataResolverConfig_CONTENT_TYPE_JSON,
							ResponseType:     configpb.ExternalDataResolverConfig_CONTENT_TYPE_JSON,
							ResponseSelector: ".",
						},
					},
				},
			}

			// Create
			mockConfigClient.EXPECT().
				CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Name": Equal(expectedResp.ConfigNode.Name),
					"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
						expectedResp.ConfigNode.DisplayName,
					)})),
					"Description": BeNil(),
					"Location":    Equal(appSpaceID),
					"Config": PointTo(MatchFields(IgnoreExtras, Fields{
						"ExternalDataResolverConfig": test.EqualProto(
							expectedResp.GetConfigNode().GetExternalDataResolverConfig()),
					})),
				})))).
				Return(&configpb.CreateConfigNodeResponse{
					Id:         sampleID,
					CreateTime: timestamppb.Now(),
					UpdateTime: timestamppb.Now(),
				}, nil)

			// Update
			mockConfigClient.EXPECT().
				UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(sampleID),
					"Config": PointTo(MatchFields(IgnoreExtras, Fields{
						"ExternalDataResolverConfig": test.EqualProto(
							expectedUpdatedResp.GetConfigNode().GetExternalDataResolverConfig()),
					})),
				})))).
				Return(&configpb.UpdateConfigNodeResponse{Id: sampleID}, nil)

			// Second Update
			mockConfigClient.EXPECT().
				UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(sampleID),
					"Config": PointTo(MatchFields(IgnoreExtras, Fields{
						"ExternalDataResolverConfig": test.EqualProto(
							expectedUpdatedResp2.GetConfigNode().GetExternalDataResolverConfig()),
					})),
				})))).
				Return(&configpb.UpdateConfigNodeResponse{Id: sampleID}, nil)

			// Read in given order
			gomock.InOrder(
				mockConfigClient.EXPECT().
					ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
						"Id": Equal(sampleID),
					})))).
					Times(3).
					Return(expectedResp, nil),

				mockConfigClient.EXPECT().
					ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
						"Id": Equal(sampleID),
					})))).
					Times(3).
					Return(expectedUpdatedResp, nil),
				mockConfigClient.EXPECT().
					ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
						"Id": Equal(sampleID),
					})))).
					Times(2).
					Return(expectedUpdatedResp2, nil),
			)

			// Delete
			mockConfigClient.EXPECT().
				DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(sampleID),
				})))).
				Return(&configpb.DeleteConfigNodeResponse{}, nil)

			resource.Test(GinkgoT(), resource.TestCase{
				Providers: map[string]*schema.Provider{
					"indykite": provider,
				},
				Steps: []resource.TestStep{
					{
						// Checking Create and Read
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-external-data-resolver1",
							`display_name = "Display name of ExternalDataResolver configuration"
							url = "https://example.com/source2"
							method = "GET"

							headers {
							  name   = "Authorization"
							  values = ["Bearer edolkUTY"]
							}

							headers {
							  name   = "Content-Type"
							  values = ["application/json"]
							}

							request_type = "json"
							response_type = "json"
							request_payload  = "{\"key\": \"value\"}"
							response_selector = "."
							`,
						),
						Check: resource.ComposeTestCheckFunc(
							testResourceDataExists(resourceName, expectedResp),
						),
					},
					{
						// Checking Update and Read
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-external-data-resolver1",
							`display_name = "Display name of ExternalDataResolver configuration"
							url = "https://example.com/source2"
							method = "GET"
							headers {
								name   = "Authorization"
								values = ["Bearer edolkUTY"]
							}

							headers {
								name   = "Content-Type"
								values = ["application/json"]
							}

							request_type = "json"
							response_type = "json"
							request_payload  = "{\"key2\": \"value2\"}"
							response_selector = "."
						`),
						Check: resource.ComposeTestCheckFunc(
							testResourceDataExists(resourceName, expectedUpdatedResp),
						),
					},
					{
						// Checking Second Update and Read
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-external-data-resolver1",
							`
							url = "https://example.com/source2"
							method = "GET"

							headers {
							  name   = "Authorization"
							  values = ["Bearer edolkUTY"]
							}

							headers {
							  name   = "Content-Type"
							  values = ["application/json"]
							}

							request_type = "json"
							response_type = "json"
							response_selector = "."
							`,
						),
						Check: resource.ComposeTestCheckFunc(
							testResourceDataExists(resourceName, expectedUpdatedResp2),
						),
					},
				},
			})
		})
	})
})

func testResourceDataExists(
	n string,
	data *configpb.ReadConfigNodeResponse,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != data.ConfigNode.Id {
			return errors.New("ID does not match")
		}
		attrs := rs.Primary.Attributes

		expectedHeaders := data.GetConfigNode().GetExternalDataResolverConfig().GetHeaders()
		actualHeadersCount, err := strconv.Atoi(attrs["headers.#"])
		if err != nil || actualHeadersCount != len(expectedHeaders) {
			return errors.New("headers count mismatch")
		}

		// Build a map from the resource attributes for easier access
		headersMap := make(map[string]map[string][]string)
		for i := range make([]struct{}, actualHeadersCount) {
			prefix := fmt.Sprintf("headers.%d.", i)
			name := attrs[prefix+"name"]
			valuesCountKey := prefix + "values.#"
			valuesCount, err := strconv.Atoi(attrs[valuesCountKey])
			if err != nil {
				return fmt.Errorf("error reading values count for header %s: %w", name, err)
			}

			values := make([]string, valuesCount)
			for j := range make([]struct{}, valuesCount) {
				valueKey := fmt.Sprintf("%svalues.%d", prefix, j)
				values[j] = attrs[valueKey]
			}
			headersMap[name] = map[string][]string{
				"values": values,
			}
		}

		// Verify headers and their values
		for name, expectedHeader := range expectedHeaders {
			actualHeader, exists := headersMap[name]
			if !exists {
				return fmt.Errorf("header %s not found", name)
			}
			if !reflect.DeepEqual(actualHeader["values"], expectedHeader.GetValues()) {
				return fmt.Errorf("values for header %s do not match", name)
			}
		}

		keys := Keys{
			"id": Equal(data.ConfigNode.Id),
			"%":  Not(BeEmpty()),

			"location":     Equal(data.ConfigNode.AppSpaceId),
			"customer_id":  Equal(data.ConfigNode.CustomerId),
			"app_space_id": Equal(data.ConfigNode.AppSpaceId),
			"name":         Equal(data.ConfigNode.Name),
			"display_name": Equal(data.ConfigNode.DisplayName),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),

			"url":    Equal(data.GetConfigNode().GetExternalDataResolverConfig().GetUrl()),
			"method": Equal(data.GetConfigNode().GetExternalDataResolverConfig().GetMethod()),
		}
		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
	}
}
