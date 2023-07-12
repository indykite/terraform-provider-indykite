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
	"strconv"
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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource Auth Flow", func() {
	const resourceName = "indykite_auth_flow.wonka"
	var (
		mockCtrl         *gomock.Controller
		mockConfigClient *configm.MockConfigManagementAPIClient
		provider         *schema.Provider
		mockedBookmark   string

		// gid:/customer/1/appSpace/1/tenant/1/authFlow/1
		authFlowIDForTenant = "gid:L2N1c3RvbWVyLzEvYXBwU3BhY2UvMS90ZW5hbnQvMS9hdXRoRmxvdy8x"
		// gid:/customer/1/appSpace/1/authFlow/1
		authFlowIDForAppSpace = "gid:L2N1c3RvbWVyLzEvYXBwU3BhY2UvMS9hdXRoRmxvdy8x"
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)

		// Bookmark must be longer than 40 chars - have just 1 added before the first write to test all cases
		mockedBookmark = "for-auth-flow" + uuid.NewRandom().String()
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

	It("Test all CRUD", func() {
		jsonAuthFlowConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				TenantId:    tenantID,
				Id:          authFlowIDForTenant,
				Name:        "wonka-auth-flow",
				DisplayName: "Wonka ChocoFlow",
				Description: wrapperspb.String("Description of the best ChocoFlow by Wonka inc."),
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_AuthFlowConfig{
					AuthFlowConfig: &configpb.AuthFlowConfig{
						SourceFormat: configpb.AuthFlowConfig_FORMAT_BARE_JSON,
						Source:       []byte(`{ "key2": 456, "key": "For testing this valid JSON is enough" }`),
					},
				},
			},
		}

		authFlowUIResp := proto.Clone(jsonAuthFlowConfigResp).(*configpb.ReadConfigNodeResponse)
		authCfg := authFlowUIResp.GetConfigNode().GetAuthFlowConfig()
		authCfg.Source = []byte("whatever UI here")
		authCfg.SourceFormat = configpb.AuthFlowConfig_FORMAT_RICH_JSON

		jsonAuthFlowAfterUpdateResp := proto.Clone(jsonAuthFlowConfigResp).(*configpb.ReadConfigNodeResponse)
		jsonAuthFlowAfterUpdateResp.GetConfigNode().DisplayName = jsonAuthFlowAfterUpdateResp.GetConfigNode().Name
		jsonAuthFlowAfterUpdateResp.GetConfigNode().Description = nil
		authCfg = jsonAuthFlowAfterUpdateResp.GetConfigNode().GetAuthFlowConfig()
		authCfg.Source = []byte("some: yaml after update")
		authCfg.SourceFormat = configpb.AuthFlowConfig_FORMAT_BARE_YAML

		yamlAuthFlowConfigResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				Id:          authFlowIDForAppSpace,
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				Name:        "wonka-yaml-auth-flow",
				DisplayName: "Yaml Wonka ChocoFlow",
				Description: wrapperspb.String("Description of the YAML ChocoFlow by Wonka inc."),
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_AuthFlowConfig{
					AuthFlowConfig: &configpb.AuthFlowConfig{
						SourceFormat: configpb.AuthFlowConfig_FORMAT_BARE_YAML,
						Source:       []byte("---\nanother_key: 789987\nkey: For testing this valid Yaml is enough"),
					},
				},
			},
		}

		createBM := "created-auth-flow" + uuid.NewRandom().String()
		createBM2 := "created-auth-flow-2" + uuid.NewRandom().String()
		updateBM := "updated-auth-flow" + uuid.NewRandom().String()
		deleteBM := "deleted-auth-flow" + uuid.NewRandom().String()
		deleteBM2 := "deleted-auth-flow-2" + uuid.NewRandom().String()

		// Create
		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(jsonAuthFlowConfigResp.ConfigNode.Name),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(jsonAuthFlowConfigResp.ConfigNode.DisplayName),
				})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{
					"Value": Equal(jsonAuthFlowConfigResp.ConfigNode.Description.Value),
				})),
				"Location": Equal(tenantID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"AuthFlowConfig": PointTo(MatchFields(IgnoreExtras, Fields{
						"SourceFormat": Equal(configpb.AuthFlowConfig_FORMAT_BARE_JSON),
						"Source": BeEquivalentTo(
							`{ "key": "For testing this valid JSON is enough", "key2": 456 }`),
					}))},
				)),
				"Bookmarks": ConsistOf(mockedBookmark),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         jsonAuthFlowConfigResp.ConfigNode.Id,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
				Bookmark:   createBM,
			}, nil)

		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":     Equal(yamlAuthFlowConfigResp.ConfigNode.Name),
				"Location": Equal(appSpaceID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"AuthFlowConfig": PointTo(MatchFields(IgnoreExtras, Fields{
						"SourceFormat": Equal(configpb.AuthFlowConfig_FORMAT_BARE_YAML),
						"Source":       ContainSubstring("key: For testing this valid Yaml is enough"),
					}))},
				)),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM, deleteBM),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         yamlAuthFlowConfigResp.ConfigNode.Id,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
				Bookmark:   createBM2,
			}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(jsonAuthFlowAfterUpdateResp.ConfigNode.Id),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"AuthFlowConfig": PointTo(MatchFields(IgnoreExtras, Fields{
						"SourceFormat": Equal(configpb.AuthFlowConfig_FORMAT_BARE_YAML),
						"Source":       BeEquivalentTo("some: yaml after update"),
					})),
				})),
				"Bookmarks": ConsistOf(mockedBookmark, createBM),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{
				Id:       jsonAuthFlowAfterUpdateResp.ConfigNode.Id,
				Bookmark: updateBM,
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(jsonAuthFlowConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM),
				})))).
				Times(2).
				Return(jsonAuthFlowConfigResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(authFlowUIResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM),
				})))).
				Times(2).
				Return(authFlowUIResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(jsonAuthFlowAfterUpdateResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM),
				})))).
				Times(3).
				Return(jsonAuthFlowAfterUpdateResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id":        Equal(yamlAuthFlowConfigResp.ConfigNode.Id),
					"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM, deleteBM, createBM2),
				})))).
				Times(2).
				Return(yamlAuthFlowConfigResp, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":        Equal(jsonAuthFlowConfigResp.ConfigNode.Id),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{Bookmark: deleteBM}, nil)

		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":        Equal(yamlAuthFlowConfigResp.ConfigNode.Id),
				"Bookmarks": ConsistOf(mockedBookmark, createBM, updateBM, deleteBM, createBM2),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{Bookmark: deleteBM2}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Error cases should be always first, easier to avoid missing mocks or incomplete plan
				{
					Config: `resource "indykite_auth_flow" "wonka" {
						name = "wonka-auth-flow"
						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "location" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_auth_flow" "wonka" {
  						location = "` + customerID + `"
  						name = "wonka-auth-flow"
					}
					`,
					ExpectError: regexp.MustCompile("one of `json,yaml` must be specified"),
				},
				{
					Config: `resource "indykite_auth_flow" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-auth-flow"
						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "customer_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_auth_flow" "wonka" {
						location = "` + customerID + `"
						name = "wonka-auth-flow"
						json = "{}"
						yaml = "\"some-string\""
					}
					`,
					ExpectError: regexp.MustCompile("\"json\": only one of `json,yaml` can be specified"),
				},
				{
					Config: `resource "indykite_auth_flow" "wonka" {
  						location = "` + customerID + `"
  						name = "wonka-auth-flow"
						json = "{id---"
					}
					`,
					ExpectError: regexp.MustCompile(`"json" contains an invalid JSON: invalid character 'i'`),
				},
				{
					Config: `resource "indykite_auth_flow" "wonka" {
  						location = "` + customerID + `"
  						name = "wonka-auth-flow"
						yaml = "{id]"
					}
					`,
					ExpectError: regexp.MustCompile(`yaml: did not find expected ',' or '}'`),
				},
				{
					Config: `resource "indykite_auth_flow" "wonka" {
  						location = "` + customerID + `"
  						name = "wonka-auth-flow"
						yaml = "{id}"
						has_ui = true
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "has_ui": its value will be decided automatically`),
				},
				{
					// Checking Create and Read (jsonAuthFlowConfigResp)
					Config: `resource "indykite_auth_flow" "wonka" {
						location = "` + tenantID + `"
						name = "wonka-auth-flow"
						display_name = "Wonka ChocoFlow"
						description = "Description of the best ChocoFlow by Wonka inc."
						json = "{ \"key\": \"For testing this valid JSON is enough\", \"key2\": 456 }"
					}
					`,
					Check: resource.ComposeTestCheckFunc(testAuthFlowResourceDataExists(
						resourceName,
						jsonAuthFlowConfigResp,
						string(jsonAuthFlowConfigResp.ConfigNode.GetAuthFlowConfig().Source),
						"",
						false,
					)),
				},
				{
					// Performs 1 read (authFlowUIResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: authFlowUIResp.ConfigNode.Id,
					ImportStateCheck: func(is []*terraform.InstanceState) error {
						if is[0].ID != authFlowUIResp.ConfigNode.Id {
							return errors.New("ID does not match")
						}
						keys := getAuthFlowDataMatcherKeys(authFlowUIResp, "", "", true)
						// keys["timeouts.%"] = Not(BeEmpty()) // added/removed in ImportStateCheck for unknown reason
						return convertOmegaMatcherToError(
							MatchAllKeys(keys),
							is[0].Attributes,
						)
					},
					// Check warning is raised once the issue is resolved and such ability is added
					// https://github.com/hashicorp/terraform-plugin-sdk/issues/864
				},
				{
					// Checking Read(jsonAuthFlowConfigResp), Update and Read(jsonAuthFlowAfterUpdateResp)
					Config: `resource "indykite_auth_flow" "wonka" {
						location = "` + tenantID + `"
						name = "wonka-auth-flow"
						yaml = "some: yaml after update"
					}
					`,
					Check: resource.ComposeTestCheckFunc(testAuthFlowResourceDataExists(
						resourceName,
						jsonAuthFlowAfterUpdateResp,
						"",
						string(jsonAuthFlowAfterUpdateResp.ConfigNode.GetAuthFlowConfig().Source),
						false,
					)),
				},
				{
					// Checking ForceNew on name change
					// Read(afterUpdateAuthFlowResp), Delete, Create and Read(withTenantLocEmailConfigResp)
					Config: `resource "indykite_auth_flow" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-yaml-auth-flow"
						display_name = "Yaml Wonka ChocoFlow"
						description = "Description of the YAML ChocoFlow by Wonka inc."
						yaml = <<-EOT
							another_key: 789987
							key: For testing this valid Yaml is enough
						EOT
					}`,
					Check: resource.ComposeTestCheckFunc(testAuthFlowResourceDataExists(
						resourceName,
						yamlAuthFlowConfigResp,
						"",
						string(yamlAuthFlowConfigResp.ConfigNode.GetAuthFlowConfig().Source),
						false,
					)),
				},
			},
		})
	})
})

func testAuthFlowResourceDataExists(
	n string,
	data *configpb.ReadConfigNodeResponse,
	json, yaml string,
	hasUI bool,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != data.ConfigNode.Id {
			return errors.New("ID does not match")
		}

		return convertOmegaMatcherToError(
			MatchAllKeys(getAuthFlowDataMatcherKeys(data, json, yaml, hasUI)),
			rs.Primary.Attributes,
		)
	}
}

func getAuthFlowDataMatcherKeys(
	data *configpb.ReadConfigNodeResponse,
	json, yaml string,
	hasUI bool,
) Keys {
	return Keys{
		"id": Equal(data.ConfigNode.Id),
		"%":  Not(BeEmpty()), // This is Terraform helper

		"location":     Not(BeEmpty()), // not in response
		"name":         Equal(data.ConfigNode.Name),
		"display_name": Equal(data.ConfigNode.DisplayName),
		"description":  Equal(data.ConfigNode.Description.GetValue()),
		"customer_id":  Equal(data.ConfigNode.CustomerId),
		"app_space_id": Equal(data.ConfigNode.AppSpaceId),
		"tenant_id":    Equal(data.ConfigNode.TenantId),
		"create_time":  Not(BeEmpty()),
		"update_time":  Not(BeEmpty()),

		"json":   Equal(json),
		"yaml":   Equal(yaml),
		"has_ui": Equal(strconv.FormatBool(hasUI)),
	}
}
