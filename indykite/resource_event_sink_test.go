// Copyright (c) 2025 IndyKite
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

var _ = Describe("Resource EventSink", func() {
	const (
		resourceName = "indykite_event_sink.development"
	)
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

	It("Test CRUD of Event Sink configuration with Kafka provider", func() {
		tfConfigDef :=
			`resource "indykite_event_sink" "development" {
				location = "%s"
				name = "%s"
				%s
			}`
		expectedResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				Id:          sampleID,
				Name:        "my-first-event-sink",
				DisplayName: "Display name of Event Sink configuration",
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_EventSinkConfig{
					EventSinkConfig: &configpb.EventSinkConfig{
						Providers: map[string]*configpb.EventSinkConfig_Provider{
							"kafka2": {
								Provider: &configpb.EventSinkConfig_Provider_Kafka{
									Kafka: &configpb.KafkaSinkConfig{
										Brokers:  []string{"my.kafka.server.example.com"},
										Topic:    "my-kafka-topic",
										Username: "my-username",
										// Password is never returned as it is sensitive value
										// Password: "",
									},
								},
							},
						},
						Routes: []*configpb.EventSinkConfig_Route{
							{
								ProviderId:     "kafka-provider",
								StopProcessing: false,
								Filter: &configpb.EventSinkConfig_Route_EventType{
									EventType: "indykite.eventsink.config.create",
								},
							},
						},
					},
				},
			},
		}
		expectedUpdatedResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				Id:          sampleID,
				Name:        "my-first-event-sink",
				Description: wrapperspb.String("sink for IK event logs for internal monitoring"),
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				CreateTime:  expectedResp.GetConfigNode().GetCreateTime(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_EventSinkConfig{
					EventSinkConfig: &configpb.EventSinkConfig{
						Providers: map[string]*configpb.EventSinkConfig_Provider{
							"kafka2": {
								Provider: &configpb.EventSinkConfig_Provider_Kafka{
									Kafka: &configpb.KafkaSinkConfig{
										Brokers:  []string{"my.kafka.server.example.com"},
										Topic:    "event-topic",
										Username: "my-username",
										// This doesn't make sense, but still valid testing scenario
										DisableTls:    true,
										TlsSkipVerify: true,
									},
								},
							},
						},
						Routes: []*configpb.EventSinkConfig_Route{
							{
								ProviderId:     "kafka-provider",
								StopProcessing: false,
								Filter: &configpb.EventSinkConfig_Route_EventType{
									EventType: "indykite.eventsink.config.update",
								},
							},
						},
					},
				},
			},
		}

		// Create
		mockConfigClient.EXPECT().
			CreateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Name": Equal(expectedResp.GetConfigNode().GetName()),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					expectedResp.GetConfigNode().GetDisplayName(),
				)})),
				"Description": BeNil(),
				"Location":    Equal(appSpaceID),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"EventSinkConfig": test.EqualProto(&configpb.EventSinkConfig{
						Providers: map[string]*configpb.EventSinkConfig_Provider{
							"kafka2": {
								Provider: &configpb.EventSinkConfig_Provider_Kafka{
									Kafka: &configpb.KafkaSinkConfig{
										Brokers:  []string{"my.kafka.server.example.com"},
										Topic:    "my-kafka-topic",
										Username: "my-username",
										Password: "some-super-secret-password",
									},
								},
							},
						},
						Routes: []*configpb.EventSinkConfig_Route{
							{
								ProviderId:     "kafka-provider",
								StopProcessing: false,
								Filter: &configpb.EventSinkConfig_Route_EventType{
									EventType: "indykite.eventsink.config.create",
								},
							},
						},
					}),
				})),
			})))).
			Return(&configpb.CreateConfigNodeResponse{
				Id:         sampleID,
				CreateTime: timestamppb.Now(),
				UpdateTime: timestamppb.Now(),
			}, nil)

		// update
		mockConfigClient.EXPECT().
			UpdateConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id":          Equal(sampleID),
				"DisplayName": PointTo(MatchFields(IgnoreExtras, Fields{"Value": BeEmpty()})),
				"Description": PointTo(MatchFields(IgnoreExtras, Fields{"Value": Equal(
					expectedUpdatedResp.GetConfigNode().GetDescription().GetValue(),
				)})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"EventSinkConfig": test.EqualProto(&configpb.EventSinkConfig{
						Providers: map[string]*configpb.EventSinkConfig_Provider{
							"kafka2": {
								Provider: &configpb.EventSinkConfig_Provider_Kafka{
									Kafka: &configpb.KafkaSinkConfig{
										Brokers:  []string{"my.kafka.server.example.com"},
										Topic:    "event-topic",
										Username: "my-username",
										Password: "changed-password",
										// This doesn't make sense, but still valid testing scenario
										DisableTls:    true,
										TlsSkipVerify: true,
									},
								},
							},
						},
						Routes: []*configpb.EventSinkConfig_Route{
							{
								ProviderId:     "kafka-provider",
								StopProcessing: false,
								Filter: &configpb.EventSinkConfig_Route_EventType{
									EventType: "indykite.eventsink.config.update",
								},
							},
						},
					}),
				})),
			})))).
			Return(&configpb.UpdateConfigNodeResponse{
				Id: sampleID,
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(sampleID),
				})))).
				Times(4).
				Return(expectedResp, nil),

			mockConfigClient.EXPECT().
				ReadConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(sampleID),
				})))).
				Times(2).
				Return(expectedUpdatedResp, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteConfigNode(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(sampleID),
			})))).
			Return(&configpb.DeleteConfigNodeResponse{}, nil)

		testResourceDataExists := func(
			n string,
			data *configpb.ReadConfigNodeResponse,
			password string,
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
				keys := Keys{
					"id": Equal(data.GetConfigNode().GetId()),
					"%":  Not(BeEmpty()), // This is Terraform helper

					"location":     Equal(data.GetConfigNode().GetAppSpaceId()), // Event Sink is always on AppSpace
					"customer_id":  Equal(data.GetConfigNode().GetCustomerId()),
					"app_space_id": Equal(data.GetConfigNode().GetAppSpaceId()),
					"name":         Equal(data.GetConfigNode().GetName()),
					"display_name": Equal(data.GetConfigNode().GetDisplayName()),
					"description":  Equal(data.GetConfigNode().GetDescription().GetValue()),
					"create_time":  Not(BeEmpty()),
					"update_time":  Not(BeEmpty()),

					"providers.#":                         Equal("1"),
					"providers.0.%":                       Equal("2"),
					"providers.0.kafka.#":                 Equal("1"),
					"providers.0.kafka.0.%":               Equal("6"),
					"providers.0.kafka.0.brokers.#":       Equal("1"),
					"providers.0.kafka.0.brokers.0":       Not(BeEmpty()),
					"providers.0.kafka.0.disable_tls":     Not(BeEmpty()),
					"providers.0.kafka.0.password":        Equal(password),
					"providers.0.kafka.0.tls_skip_verify": Not(BeEmpty()),
					"providers.0.kafka.0.topic":           Not(BeEmpty()),
					"providers.0.kafka.0.username":        Not(BeEmpty()),
					"providers.0.provider_name":           Not(BeEmpty()),
					"routes.#":                            Equal("1"),
					"routes.0.context_key_value_filter.#": Equal("0"),
					"routes.0.%":                          Equal("4"),
					"routes.0.provider_id":                Not(BeEmpty()),
					"routes.0.event_type_filter":          Not(BeEmpty()),
					"routes.0.stop_processing":            Not(BeEmpty()),
				}

				return convertOmegaMatcherToError(MatchAllKeys(keys), attrs)
			}
		}

		validKafkaBlock := `
		providers  {
			provider_name = "kafka2"
			kafka {
				brokers = ["my.kafka.server.example.com"]
				topic = "my-kafka-topic"
				username = "my-username"
				password = "some-super-secret-password"
			}
		}
		routes {
			provider_id = "kafka-provider"
			stop_processing = false
			event_type_filter = "indykite.eventsink.config.create"
		}`

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors case must be always first
				{
					Config:      fmt.Sprintf(tfConfigDef, "ccc", "name", validKafkaBlock),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", ``),
					ExpectError: regexp.MustCompile(`At least 1 "providers" blocks are required.`),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name", `
					display_name = "Display name of Event Sink configuration"
					providers  {
						provider_name = "kafka2"
						kafka {
							brokers = ["my.kafka.server.example.com"]
							topic = "my-kafka-topic"
							username = "my-username"
							password = "some-super-secret-password"
							}
					}
					routes {
	                        stop_processing = false
							event_type_filter = "indykite.eventsink.config.create"
						}
					`),
					ExpectError: regexp.MustCompile(
						`The argument "provider_id" is required, but no definition was found.`),
				},
				{
					// Checking Create and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-event-sink",
						`display_name = "Display name of Event Sink configuration"
						providers  {
							provider_name = "kafka2"
							kafka {
								brokers = ["my.kafka.server.example.com"]
								topic = "my-kafka-topic"
								username = "my-username"
								password = "some-super-secret-password"
							}
						}
						routes {
							provider_id = "kafka-provider"
	                        stop_processing = false
							event_type_filter = "indykite.eventsink.config.create"
						}
						`,
					),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(resourceName, expectedResp, "some-super-secret-password"),
					),
				},
				{
					// Performs 1 read
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: sampleID,
				},
				{
					// Checking Create and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-event-sink",
						`description = "sink for IK event logs for internal monitoring"
						providers  {
							provider_name = "kafka2"
							kafka {
								brokers = ["my.kafka.server.example.com"]
								topic = "event-topic"
								username = "my-username"
								password = "changed-password"
								disable_tls = true
								tls_skip_verify = true
							}
						}
						routes {
							provider_id = "kafka-provider"
	                        stop_processing = false
							event_type_filter = "indykite.eventsink.config.update"
						}
					`,
					),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(resourceName, expectedUpdatedResp, "changed-password"),
					),
				},
			},
		})
	})
})
