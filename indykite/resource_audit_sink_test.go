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
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource AuditSink", func() {
	const (
		resourceName = "indykite_audit_sink.development"
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

	It("Test CRUD of Audit Sink configuration with Kafka provider", func() {
		tfConfigDef :=
			`resource "indykite_audit_sink" "development" {
				location = "%s"
				name = "%s"
				%s
			}`
		expectedResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				Id:          sampleID,
				Name:        "my-first-audit-sink",
				DisplayName: "Display name of Audit Sink configuration",
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				CreateTime:  timestamppb.Now(),
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_AuditSinkConfig{
					AuditSinkConfig: &configpb.AuditSinkConfig{
						Provider: &configpb.AuditSinkConfig_Kafka{
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
			},
		}
		expectedUpdatedResp := &configpb.ReadConfigNodeResponse{
			ConfigNode: &configpb.ConfigNode{
				Id:          sampleID,
				Name:        "my-first-audit-sink",
				Description: wrapperspb.String("sink for IK audit logs for internal monitoring"),
				CustomerId:  customerID,
				AppSpaceId:  appSpaceID,
				CreateTime:  expectedResp.ConfigNode.CreateTime,
				UpdateTime:  timestamppb.Now(),
				Config: &configpb.ConfigNode_AuditSinkConfig{
					AuditSinkConfig: &configpb.AuditSinkConfig{
						Provider: &configpb.AuditSinkConfig_Kafka{
							Kafka: &configpb.KafkaSinkConfig{
								Brokers:  []string{"my.kafka.server.example.com", "another.kafka.server.example.com"},
								Topic:    "audit-topic",
								Username: "my-username",
								// This doesn't make sense, but still valid testing scenario
								DisableTls:    true,
								TlsSkipVerify: true,
							},
						},
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
					"AuditSinkConfig": test.EqualProto(&configpb.AuditSinkConfig{
						Provider: &configpb.AuditSinkConfig_Kafka{Kafka: &configpb.KafkaSinkConfig{
							Brokers:  []string{"my.kafka.server.example.com"},
							Topic:    "my-kafka-topic",
							Username: "my-username",
							Password: "some-super-secret-password",
						}},
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
					expectedUpdatedResp.ConfigNode.Description.GetValue(),
				)})),
				"Config": PointTo(MatchFields(IgnoreExtras, Fields{
					"AuditSinkConfig": test.EqualProto(&configpb.AuditSinkConfig{
						Provider: &configpb.AuditSinkConfig_Kafka{Kafka: &configpb.KafkaSinkConfig{
							Brokers:  []string{"my.kafka.server.example.com", "another.kafka.server.example.com"},
							Topic:    "audit-topic",
							Username: "my-username",
							Password: "changed-password",
							// This doesn't make sense, but still valid testing scenario
							DisableTls:    true,
							TlsSkipVerify: true,
						}},
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
				if rs.Primary.ID != data.ConfigNode.Id {
					return errors.New("ID does not match")
				}
				attrs := rs.Primary.Attributes

				kafkaCfg := data.ConfigNode.GetAuditSinkConfig().GetKafka()

				keys := Keys{
					"id": Equal(data.ConfigNode.Id),
					"%":  Not(BeEmpty()), // This is Terraform helper

					"location":     Equal(data.ConfigNode.AppSpaceId), // Audit Sink is always on AppSpace level
					"customer_id":  Equal(data.ConfigNode.CustomerId),
					"app_space_id": Equal(data.ConfigNode.AppSpaceId),
					"name":         Equal(data.ConfigNode.Name),
					"display_name": Equal(data.ConfigNode.DisplayName),
					"description":  Equal(data.ConfigNode.GetDescription().GetValue()),
					"create_time":  Not(BeEmpty()),
					"update_time":  Not(BeEmpty()),

					"kafka.#":                 Equal("1"),     // Terraform helper - always max 1 kafka element
					"kafka.0.%":               Not(BeEmpty()), // This is Terraform helper
					"kafka.0.topic":           Equal(kafkaCfg.Topic),
					"kafka.0.username":        Equal(kafkaCfg.Username),
					"kafka.0.password":        Equal(password),
					"kafka.0.disable_tls":     Equal(strconv.FormatBool(kafkaCfg.DisableTls)),
					"kafka.0.tls_skip_verify": Equal(strconv.FormatBool(kafkaCfg.TlsSkipVerify)),
				}

				addStringArrayToKeys(keys, "kafka.0.brokers", kafkaCfg.Brokers)

				return convertOmegaMatcherToError(MatchAllKeys(keys), attrs)
			}
		}

		validKafkaBlock := `kafka {
			brokers = ["my.kafka.server.example.com"]
			topic = "my-kafka-topic"
			username = "my-username"
			password = "some-super-secret-password"
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
					Config:      fmt.Sprintf(tfConfigDef, "ccc", "name", ``),
					ExpectError: regexp.MustCompile(`At least 1 "kafka" blocks are required`),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, "ccc", "name", validKafkaBlock+"\n"+validKafkaBlock),
					ExpectError: regexp.MustCompile(`No more than 1 "kafka" blocks are allowed`),
				},
				{
					// Checking Create and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-audit-sink",
						`display_name = "Display name of Audit Sink configuration"
						kafka {
							brokers = ["my.kafka.server.example.com"]
							topic = "my-kafka-topic"
							username = "my-username"
							password = "some-super-secret-password"
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
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-audit-sink",
						`description = "sink for IK audit logs for internal monitoring"
						kafka {
							brokers = ["my.kafka.server.example.com", "another.kafka.server.example.com"]
							topic = "audit-topic"
							username = "my-username"
							password = "changed-password"
							disable_tls = true
							tls_skip_verify = true
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
