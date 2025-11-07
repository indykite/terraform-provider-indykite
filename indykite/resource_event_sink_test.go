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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource EventSink", func() {
	const (
		resourceName  = "indykite_event_sink.development"
		resourceName2 = "indykite_event_sink.development2"
		resourceName3 = "indykite_event_sink.development3"
	)
	var (
		mockServer *httptest.Server
		provider   *schema.Provider
	)

	BeforeEach(func() {
		provider = indykite.Provider()
	})

	AfterEach(func() {
		if mockServer != nil {
			mockServer.Close()
		}
	})

	It("Test CRUD of Event Sink configuration with Kafka provider", func() {
		tfConfigDef :=
			`resource "indykite_event_sink" "development" {
				location = "%s"
				name = "%s"
				%s
			}`
		tfConfigDef2 :=
			`resource "indykite_event_sink" "development2" {
				location = "%s"
				name = "%s"
				%s
			}`
		tfConfigDef3 :=
			`resource "indykite_event_sink" "development3" {
				location = "%s"
				name = "%s"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()

		// Track which configuration we're serving
		currentConfig := "kafka"

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/event-sinks"):
				// Create response based on current config
				var resp indykite.EventSinkResponse
				switch currentConfig {
				case "kafka":
					resp = indykite.EventSinkResponse{
						ID:          sampleID,
						Name:        "my-first-event-sink",
						DisplayName: "Display name of Event Sink configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Config: map[string]any{
							"providers": map[string]any{
								"kafka2": map[string]any{
									"kafka": map[string]any{
										"brokers":     []string{"my.kafka.server.example.com:9092"},
										"topic":       "my-kafka-topic",
										"username":    "my-username",
										"displayName": "provider-display-name",
									},
								},
							},
							"routes": []any{
								map[string]any{
									"providerId":     "kafka-provider",
									"stopProcessing": false,
									"keysValues": map[string]any{
										"eventType": "indykite.audit.config.create",
									},
									"displayName": "route-display-name",
									"id":          "route-id",
								},
							},
						},
						CreateTime: createTime,
						UpdateTime: updateTime,
					}
				case "azuregrid":
					resp = indykite.EventSinkResponse{
						ID:          sampleID,
						Name:        "my-first-event-sink",
						DisplayName: "Display name of Event Sink configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Config: map[string]any{
							"providers": map[string]any{
								"azuregrid": map[string]any{
									"azureEventGrid": map[string]any{
										"topicEndpoint": "https://ik-test.eventgrid.azure.net/api/events",
									},
								},
							},
							"routes": []any{
								map[string]any{
									"providerId":     "azuregrid",
									"stopProcessing": false,
									"keysValues": map[string]any{
										"eventType": "indykite.audit.config.create",
									},
								},
							},
						},
						CreateTime: createTime,
						UpdateTime: updateTime,
					}
				case "azurebus":
					resp = indykite.EventSinkResponse{
						ID:          sampleID,
						Name:        "my-first-event-sink",
						DisplayName: "Display name of Event Sink configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Config: map[string]any{
							"providers": map[string]any{
								"azurebus": map[string]any{
									"azureServiceBus": map[string]any{
										"queueOrTopicName": "your-queue",
									},
								},
							},
							"routes": []any{
								map[string]any{
									"providerId":     "azurebus",
									"stopProcessing": false,
									"keysValues": map[string]any{
										"eventType": "indykite.audit.capture.*",
										"keyValuePairs": []any{
											map[string]any{
												"key":   "captureLabel",
												"value": "HAS",
											},
										},
									},
								},
							},
						},
						CreateTime: createTime,
						UpdateTime: updateTime,
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
				// Read response
				var resp indykite.EventSinkResponse
				switch currentConfig {
				case "kafka":
					resp = indykite.EventSinkResponse{
						ID:          sampleID,
						Name:        "my-first-event-sink",
						DisplayName: "Display name of Event Sink configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Config: map[string]any{
							"providers": map[string]any{
								"kafka2": map[string]any{
									"kafka": map[string]any{
										"brokers":     []string{"my.kafka.server.example.com:9092"},
										"topic":       "my-kafka-topic",
										"username":    "my-username",
										"displayName": "provider-display-name",
									},
								},
							},
							"routes": []any{
								map[string]any{
									"providerId":     "kafka-provider",
									"stopProcessing": false,
									"keysValues": map[string]any{
										"eventType": "indykite.audit.config.create",
									},
									"displayName": "route-display-name",
									"id":          "route-id",
								},
							},
						},
						CreateTime: createTime,
						UpdateTime: updateTime,
					}
				case "kafka-updated":
					resp = indykite.EventSinkResponse{
						ID:          sampleID,
						Name:        "my-first-event-sink",
						Description: "sink for IK event logs for internal monitoring",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Config: map[string]any{
							"providers": map[string]any{
								"kafka2": map[string]any{
									"kafka": map[string]any{
										"brokers":       []string{"my.kafka.server.example.com:9092"},
										"topic":         "event-topic",
										"username":      "my-username",
										"disableTls":    true,
										"tlsSkipVerify": true,
										"displayName":   "provider-display-name",
									},
								},
							},
							"routes": []any{
								map[string]any{
									"providerId":     "kafka-provider",
									"stopProcessing": false,
									"keysValues": map[string]any{
										"eventType": "indykite.audit.config.update",
									},
									"displayName": "route-display-name-upd",
									"id":          "route-id",
								},
							},
						},
						CreateTime: createTime,
						UpdateTime: time.Now(),
					}
				case "azuregrid":
					resp = indykite.EventSinkResponse{
						ID:          sampleID,
						Name:        "my-first-event-sink",
						DisplayName: "Display name of Event Sink configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Config: map[string]any{
							"providers": map[string]any{
								"azuregrid": map[string]any{
									"azureEventGrid": map[string]any{
										"topicEndpoint": "https://ik-test.eventgrid.azure.net/api/events",
									},
								},
							},
							"routes": []any{
								map[string]any{
									"providerId":     "azuregrid",
									"stopProcessing": false,
									"keysValues": map[string]any{
										"eventType": "indykite.audit.config.create",
									},
								},
							},
						},
						CreateTime: createTime,
						UpdateTime: updateTime,
					}
				case "azurebus":
					resp = indykite.EventSinkResponse{
						ID:          sampleID,
						Name:        "my-first-event-sink",
						DisplayName: "Display name of Event Sink configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Config: map[string]any{
							"providers": map[string]any{
								"azurebus": map[string]any{
									"azureServiceBus": map[string]any{
										"queueOrTopicName": "your-queue",
									},
								},
							},
							"routes": []any{
								map[string]any{
									"providerId":     "azurebus",
									"stopProcessing": false,
									"keysValues": map[string]any{
										"eventType": "indykite.audit.capture.*",
										"keyValuePairs": []any{
											map[string]any{
												"key":   "captureLabel",
												"value": "HAS",
											},
										},
									},
								},
							},
						},
						CreateTime: createTime,
						UpdateTime: updateTime,
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
				// Update - switch to kafka-updated
				currentConfig = "kafka-updated"
				resp := indykite.EventSinkResponse{
					ID:          sampleID,
					Name:        "my-first-event-sink",
					Description: "sink for IK event logs for internal monitoring",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Config: map[string]any{
						"providers": map[string]any{
							"kafka2": map[string]any{
								"kafka": map[string]any{
									"brokers":       []string{"my.kafka.server.example.com:9092"},
									"topic":         "event-topic",
									"username":      "my-username",
									"disableTls":    true,
									"tlsSkipVerify": true,
									"displayName":   "provider-display-name",
								},
							},
						},
						"routes": []any{
							map[string]any{
								"providerId":     "kafka-provider",
								"stopProcessing": false,
								"keysValues": map[string]any{
									"eventType": "indykite.audit.config.update",
								},
								"displayName": "route-display-name-upd",
								"id":          "route-id",
							},
						},
					},
					CreateTime: createTime,
					UpdateTime: time.Now(),
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)

			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))

		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
			client := indykite.NewTestRestClient(mockServer.URL+"/configs/v1", mockServer.Client())
			ctx = indykite.WithClient(ctx, client)
			return cfgFunc(ctx, data)
		}

		validKafkaBlock := `
		providers  {
			provider_name = "kafka2"
			kafka {
				brokers = ["my.kafka.server.example.com:9092"]
				topic = "my-kafka-topic"
				username = "my-username"
				password = "some-super-secret-password"
				provider_display_name = "provider-display-name"
			}
		}
		routes {
			provider_id = "kafka-provider"
			stop_processing = false
			keys_values_filter {
				event_type = "indykite.audit.config.create"
			}
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
							brokers = ["my.kafka.server.example.com:9092"]
							topic = "my-kafka-topic"
							username = "my-username"
							password = "some-super-secret-password"
							provider_display_name = "provider-display-name"
							}
					}
					routes {
                        stop_processing = false
						keys_values_filter {
							event_type = "indykite.audit.config.create"
						}
					}
					`),
					ExpectError: regexp.MustCompile(
						`The argument "provider_id" is required, but no definition was found.`),
				},
				{
					// Checking Create and Read - Kafka
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-event-sink",
						`display_name = "Display name of Event Sink configuration"
						providers  {
							provider_name = "kafka2"
							kafka {
								brokers = ["my.kafka.server.example.com:9092"]
								topic = "my-kafka-topic"
								username = "my-username"
								password = "some-super-secret-password"
								provider_display_name = "provider-display-name"
							}
						}
						routes {
							provider_id = "kafka-provider"
                        stop_processing = false
							keys_values_filter {
								event_type = "indykite.audit.config.create"
							}
							route_display_name = "route-display-name"
							route_id = "route-id"
						}
						`,
					),
					Check: resource.ComposeTestCheckFunc(
						testEventSinkResourceDataExists(resourceName),
					),
				},
				{
					// Import test
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: sampleID,
				},
				{
					// Update Kafka configuration
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-event-sink",
						`description = "sink for IK event logs for internal monitoring"
						providers  {
							provider_name = "kafka2"
							kafka {
								brokers = ["my.kafka.server.example.com:9092"]
								topic = "event-topic"
								username = "my-username"
								password = "changed-password"
								disable_tls = true
								tls_skip_verify = true
								provider_display_name = "provider-display-name"
							}
						}
						routes {
							provider_id = "kafka-provider"
                        stop_processing = false
							keys_values_filter {
								event_type = "indykite.audit.config.update"
							}
							route_display_name = "route-display-name-upd"
							route_id = "route-id"
						}
					`,
					),
					Check: resource.ComposeTestCheckFunc(
						testEventSinkResourceDataExists(resourceName),
					),
				},
			},
		})

		// Reset config for next test
		currentConfig = "azuregrid"

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				{
					// Azure Event Grid
					Config: fmt.Sprintf(tfConfigDef2, appSpaceID, "my-first-event-sink",
						`display_name = "Display name of Event Sink configuration"
						providers  {
							provider_name = "azuregrid"
							azure_event_grid {
								topic_endpoint = "https://ik-test.eventgrid.azure.net/api/events"
								access_key = "secret-access-key"
							}
						}
						routes {
							provider_id = "azuregrid"
                        stop_processing = false
							keys_values_filter {
								event_type = "indykite.audit.config.create"
							}
						}
						`,
					),
					Check: resource.ComposeTestCheckFunc(
						testEventSinkResourceDataExists(resourceName2),
					),
				},
				{
					ResourceName:  resourceName2,
					ImportState:   true,
					ImportStateId: sampleID,
				},
			},
		})

		// Reset config for Azure Service Bus test
		currentConfig = "azurebus"

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				{
					// Azure Service Bus
					Config: fmt.Sprintf(tfConfigDef3, appSpaceID, "my-first-event-sink",
						`display_name = "Display name of Event Sink configuration"
						providers  {
							provider_name = "azurebus"
							azure_service_bus {
								connection_string = "personal-connection-info"
								queue_or_topic_name = "your-queue"
							}
						}
						routes {
							provider_id = "azurebus"
                        stop_processing = false
							keys_values_filter {
								key_value_pairs{
									key        = "captureLabel"
									value      = "HAS"
								}
								event_type  = "indykite.audit.capture.*"
							}
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testEventSinkResourceDataExists(resourceName3),
					),
				},
				{
					ResourceName:  resourceName3,
					ImportState:   true,
					ImportStateId: sampleID,
				},
			},
		})
	})

	It("Test import by name with location", func() {
		tfConfigDef :=
			`resource "indykite_event_sink" "development" {
					location = "` + appSpaceID + `"
					name = "wonka-sink"
					display_name = "Wonka Event Sink"

					providers {
						provider_name = "kafka2"
						kafka {
							brokers = ["localhost:9092"]
							topic = "my-kafka-topic"
							username = "my-username"
							password = "some-super-secret-password"
						}
					}

					routes {
						provider_id = "kafka-provider"
						stop_processing = false
						keys_values_filter {
							event_type = "indykite.audit.*"
						}
					}
				}`

		createTime := time.Now()
		updateTime := time.Now()

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/event-sinks"):
				resp := indykite.EventSinkResponse{
					ID:          sampleID,
					Name:        "wonka-sink",
					DisplayName: "Wonka Event Sink",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Config: map[string]any{
						"providers": map[string]any{
							"kafka2": map[string]any{
								"kafka": map[string]any{
									"brokers":  []string{"localhost:9092"},
									"topic":    "my-kafka-topic",
									"username": "my-username",
								},
							},
						},
						"routes": []any{
							map[string]any{
								"providerId":     "kafka-provider",
								"stopProcessing": false,
								"keysValues": map[string]any{
									"eventType": "indykite.audit.*",
								},
							},
						},
					},
					CreateTime: createTime,
					UpdateTime: updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/event-sinks/"):
				// Support both ID and name?location=appSpaceID formats
				pathAfterSinks := strings.TrimPrefix(r.URL.Path, "/configs/v1/event-sinks/")
				isNameLookup := strings.Contains(pathAfterSinks, "wonka-sink")
				isIDLookup := strings.Contains(pathAfterSinks, sampleID)

				if isNameLookup || isIDLookup {
					resp := indykite.EventSinkResponse{
						ID:          sampleID,
						Name:        "wonka-sink",
						DisplayName: "Wonka Event Sink",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Config: map[string]any{
							"providers": map[string]any{
								"kafka2": map[string]any{
									"kafka": map[string]any{
										"brokers":  []string{"localhost:9092"},
										"topic":    "my-kafka-topic",
										"username": "my-username",
									},
								},
							},
							"routes": []any{
								map[string]any{
									"providerId":     "kafka-provider",
									"stopProcessing": false,
									"keysValues": map[string]any{
										"eventType": "indykite.audit.*",
									},
								},
							},
						},
						CreateTime: createTime,
						UpdateTime: updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)
				} else {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
				}

			case r.Method == http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)

			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer mockServer.Close()

		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
			client := indykite.NewTestRestClient(mockServer.URL+"/configs/v1", mockServer.Client())
			ctx = indykite.WithClient(ctx, client)
			return cfgFunc(ctx, data)
		}

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				{
					Config: tfConfigDef,
					Check: resource.ComposeTestCheckFunc(
						testEventSinkResourceDataExists(resourceName),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: "wonka-sink?location=" + appSpaceID,
				},
			},
		})
	})
})

func testEventSinkResourceDataExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != sampleID {
			return errors.New("ID does not match")
		}
		attrs := rs.Primary.Attributes

		keys := Keys{
			"id": Equal(sampleID),
			"%":  Not(BeEmpty()),

			"location":     Equal(appSpaceID),
			"customer_id":  Equal(customerID),
			"app_space_id": Equal(appSpaceID),
			"name":         Not(BeEmpty()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
		}

		// Verify providers and routes exist
		if attrs["providers.#"] != "" {
			keys["providers.#"] = Not(BeEmpty())
		}
		if attrs["routes.#"] != "" {
			keys["routes.#"] = Not(BeEmpty())
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
	}
}
