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
	"io"
	"maps"
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
		resourceName4 = "indykite_event_sink.development4"
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
		tfConfigDef4 :=
			`resource "indykite_event_sink" "development4" {
				location = "%s"
				name = "%s"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()

		// Track which configuration we're serving
		currentConfig := "kafka"

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Server runs in its own goroutine; GinkgoRecover lets Gomega
			// assertions in the handlers fail the spec instead of panicking silently.
			defer GinkgoRecover()
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
				case "pubsub":
					resp = indykite.EventSinkResponse{
						ID:          sampleID,
						Name:        "my-first-event-sink",
						DisplayName: "Display name of Event Sink configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Config: map[string]any{
							"providers": map[string]any{
								"gcp-pubsub": map[string]any{
									"pubsub": map[string]any{
										"projectId":   "my-gcp-project",
										"topicName":   "my-pubsub-topic",
										"displayName": "pubsub-display-name",
									},
								},
							},
							"routes": []any{
								map[string]any{
									"providerId":     "gcp-pubsub",
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
				case "pubsub":
					resp = indykite.EventSinkResponse{
						ID:          sampleID,
						Name:        "my-first-event-sink",
						DisplayName: "Display name of Event Sink configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Config: map[string]any{
							"providers": map[string]any{
								"gcp-pubsub": map[string]any{
									"pubsub": map[string]any{
										"projectId":   "my-gcp-project",
										"topicName":   "my-pubsub-topic",
										"displayName": "pubsub-display-name",
									},
								},
							},
							"routes": []any{
								map[string]any{
									"providerId":     "gcp-pubsub",
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
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
				// Assert the update payload uses the flat shape the backend expects
				// (top-level providers/routes/include_cdc_events), not the legacy
				// nested "config" envelope that the real API silently ignores.
				var reqBody map[string]any
				Expect(json.NewDecoder(r.Body).Decode(&reqBody)).To(Succeed())
				Expect(reqBody).NotTo(HaveKey("config"))
				Expect(reqBody).To(HaveKey("providers"))
				Expect(reqBody).To(HaveKey("routes"))
				Expect(reqBody).To(HaveKey("include_cdc_events"))

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
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
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
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
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
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
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

		// Reset config for Pub/Sub test
		currentConfig = "pubsub"

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					// Google Cloud Pub/Sub
					Config: fmt.Sprintf(tfConfigDef4, appSpaceID, "my-first-event-sink",
						`display_name = "Display name of Event Sink configuration"
						providers  {
							provider_name = "gcp-pubsub"
							pubsub {
								project_id = "my-gcp-project"
								topic_name = "my-pubsub-topic"
								credentials_json = "{\"type\":\"service_account\"}"
								provider_display_name = "pubsub-display-name"
							}
						}
						routes {
							provider_id = "gcp-pubsub"
                        stop_processing = false
							keys_values_filter {
								event_type = "indykite.audit.config.create"
							}
						}`,
					),
					Check: resource.ComposeTestCheckFunc(
						testEventSinkResourceDataExists(resourceName4),
					),
				},
				{
					ResourceName:  resourceName4,
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

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Server runs in its own goroutine; GinkgoRecover lets Gomega
			// assertions in the handlers fail the spec instead of panicking silently.
			defer GinkgoRecover()
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

		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
			client := indykite.NewTestRestClient(mockServer.URL+"/configs/v1", mockServer.Client())
			ctx = indykite.WithClient(ctx, client)
			return cfgFunc(ctx, data)
		}

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
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

	It("Test include_cdc_events round-trip on create/update", func() {
		tfConfigDef :=
			`resource "indykite_event_sink" "development" {
				location = "%s"
				name = "my-first-event-sink"
				%s
				providers {
					provider_name = "kafka2"
					kafka {
						brokers = ["my.kafka.server.example.com:9092"]
						topic = "my-kafka-topic"
						username = "my-username"
						password = "some-super-secret-password"
					}
				}
				routes {
					provider_id = "kafka-provider"
					stop_processing = false
					keys_values_filter {
						event_type = "indykite.audit.config.create"
					}
				}
			}`

		createTime := time.Now()
		updateTime := time.Now()

		// Track what was received on create/update and what to serve back.
		var lastCreateBody map[string]any
		var lastUpdateBody map[string]any
		currentIncludeCDC := false

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Server runs in its own goroutine; GinkgoRecover lets Gomega
			// assertions in the handlers fail the spec instead of panicking silently.
			defer GinkgoRecover()
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/event-sinks"):
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &lastCreateBody)
				if v, ok := lastCreateBody["include_cdc_events"].(bool); ok {
					currentIncludeCDC = v
				}
				resp := indykite.EventSinkResponse{
					ID:         sampleID,
					Name:       "my-first-event-sink",
					CustomerID: customerID,
					AppSpaceID: appSpaceID,
					Config: map[string]any{
						"providers": map[string]any{
							"kafka2": map[string]any{
								"kafka": map[string]any{
									"brokers":  []string{"my.kafka.server.example.com:9092"},
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
									"eventType": "indykite.audit.config.create",
								},
							},
						},
						"include_cdc_events": currentIncludeCDC,
					},
					CreateTime: createTime,
					UpdateTime: updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &lastUpdateBody)
				// providers/routes/include_cdc_events are sent at the top level on update requests.
				if v, ok := lastUpdateBody["include_cdc_events"].(bool); ok {
					currentIncludeCDC = v
				}
				resp := indykite.EventSinkResponse{
					ID:         sampleID,
					Name:       "my-first-event-sink",
					CustomerID: customerID,
					AppSpaceID: appSpaceID,
					Config: map[string]any{
						"providers": map[string]any{
							"kafka2": map[string]any{
								"kafka": map[string]any{
									"brokers":  []string{"my.kafka.server.example.com:9092"},
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
									"eventType": "indykite.audit.config.create",
								},
							},
						},
						"include_cdc_events": currentIncludeCDC,
					},
					CreateTime: createTime,
					UpdateTime: time.Now(),
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
				resp := indykite.EventSinkResponse{
					ID:         sampleID,
					Name:       "my-first-event-sink",
					CustomerID: customerID,
					AppSpaceID: appSpaceID,
					Config: map[string]any{
						"providers": map[string]any{
							"kafka2": map[string]any{
								"kafka": map[string]any{
									"brokers":  []string{"my.kafka.server.example.com:9092"},
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
									"eventType": "indykite.audit.config.create",
								},
							},
						},
						"include_cdc_events": currentIncludeCDC,
					},
					CreateTime: createTime,
					UpdateTime: updateTime,
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

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					// Create with CDC enabled.
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, `include_cdc_events = true`),
					Check: resource.ComposeTestCheckFunc(
						testEventSinkResourceDataExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "include_cdc_events", "true"),
						func(_ *terraform.State) error {
							if v, _ := lastCreateBody["include_cdc_events"].(bool); !v {
								return errors.New("expected create body to send include_cdc_events=true")
							}
							return nil
						},
					),
				},
				{
					// Toggle CDC off via update.
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, `include_cdc_events = false`),
					Check: resource.ComposeTestCheckFunc(
						testEventSinkResourceDataExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "include_cdc_events", "false"),
						func(_ *terraform.State) error {
							v, present := lastUpdateBody["include_cdc_events"]
							if !present {
								return errors.New("update body missing include_cdc_events")
							}
							if b, _ := v.(bool); b {
								return errors.New("expected update body include_cdc_events=false")
							}
							if _, ok := lastUpdateBody["providers"].(map[string]any); !ok {
								return errors.New("update body missing top-level providers")
							}
							if _, ok := lastUpdateBody["routes"].([]any); !ok {
								return errors.New("update body missing top-level routes")
							}
							return nil
						},
					),
				},
			},
		})
	})

	It("Test include_cdc_events defaults to false when omitted", func() {
		tfConfigDef :=
			`resource "indykite_event_sink" "development" {
				location = "%s"
				name = "my-first-event-sink"
				providers {
					provider_name = "kafka2"
					kafka {
						brokers = ["my.kafka.server.example.com:9092"]
						topic = "my-kafka-topic"
						username = "my-username"
						password = "some-super-secret-password"
					}
				}
				routes {
					provider_id = "kafka-provider"
					stop_processing = false
					keys_values_filter {
						event_type = "indykite.audit.config.create"
					}
				}
			}`

		createTime := time.Now()
		updateTime := time.Now()
		var lastCreateBody map[string]any

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Server runs in its own goroutine; GinkgoRecover lets Gomega
			// assertions in the handlers fail the spec instead of panicking silently.
			defer GinkgoRecover()
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/event-sinks"):
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &lastCreateBody)
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(indykite.EventSinkResponse{
					ID:         sampleID,
					Name:       "my-first-event-sink",
					CustomerID: customerID,
					AppSpaceID: appSpaceID,
					Config: map[string]any{
						"providers": map[string]any{
							"kafka2": map[string]any{
								"kafka": map[string]any{
									"brokers":  []string{"my.kafka.server.example.com:9092"},
									"topic":    "my-kafka-topic",
									"username": "my-username",
								},
							},
						},
						"routes": []any{
							map[string]any{
								"providerId":     "kafka-provider",
								"stopProcessing": false,
								"keysValues":     map[string]any{"eventType": "indykite.audit.config.create"},
							},
						},
					},
					CreateTime: createTime,
					UpdateTime: updateTime,
				})

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(indykite.EventSinkResponse{
					ID:         sampleID,
					Name:       "my-first-event-sink",
					CustomerID: customerID,
					AppSpaceID: appSpaceID,
					Config: map[string]any{
						"providers": map[string]any{
							"kafka2": map[string]any{
								"kafka": map[string]any{
									"brokers":  []string{"my.kafka.server.example.com:9092"},
									"topic":    "my-kafka-topic",
									"username": "my-username",
								},
							},
						},
						"routes": []any{
							map[string]any{
								"providerId":     "kafka-provider",
								"stopProcessing": false,
								"keysValues":     map[string]any{"eventType": "indykite.audit.config.create"},
							},
						},
					},
					CreateTime: createTime,
					UpdateTime: updateTime,
				})

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

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(tfConfigDef, appSpaceID),
					Check: resource.ComposeTestCheckFunc(
						testEventSinkResourceDataExists(resourceName),
						resource.TestCheckResourceAttr(resourceName, "include_cdc_events", "false"),
						func(_ *terraform.State) error {
							v, present := lastCreateBody["include_cdc_events"]
							if !present {
								return errors.New("create body missing include_cdc_events")
							}
							if b, _ := v.(bool); b {
								return errors.New("expected include_cdc_events=false by default")
							}
							return nil
						},
					),
				},
			},
		})
	})

	It("Test route filter validation matches the API", func() {
		tfConfigDef :=
			`resource "indykite_event_sink" "development" {
				project_id = "` + appSpaceID + `"
				name = "filter-sink"
				providers {
					provider_name = "kafka-a"
					kafka {
						brokers = ["kafka-a.example.com:9092"]
						topic = "topic-a"
						username = "user-a"
						password = "password-a"
					}
				}
				routes {
					provider_id = "kafka-a"
					%s
				}
			}`

		// Every case must be rejected at plan time, before any request reaches the API.
		mockServer = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Fail(fmt.Sprintf("unexpected API call %s %s", r.Method, r.URL.Path))
		}))

		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
			client := indykite.NewTestRestClient(mockServer.URL+"/configs/v1", mockServer.Client())
			ctx = indykite.WithClient(ctx, client)
			return cfgFunc(ctx, data)
		}

		filterWith := func(eventType, key, value string) string {
			return fmt.Sprintf(`keys_values_filter {
				event_type = %q
				key_value_pairs {
					key = %q
					value = %q
				}
			}`, eventType, key, value)
		}
		invalidEventType := regexp.MustCompile(`must contain only letters, numbers, underscores, asterisks and dots`)
		emptyPairField := func(field string) *regexp.Regexp {
			return regexp.MustCompile(`expected "routes.0.keys_values_filter.0.key_value_pairs.0.` + field +
				`" to not be an empty string`)
		}

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					Config:      fmt.Sprintf(tfConfigDef, `stop_processing = true`),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`Insufficient keys_values_filter blocks`),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, filterWith(`indykite.audit\config.*`, "captureLabel", "Car")),
					PlanOnly:    true,
					ExpectError: invalidEventType,
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, filterWith("indykite.audit config", "captureLabel", "Car")),
					PlanOnly:    true,
					ExpectError: invalidEventType,
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, filterWith("", "captureLabel", "Car")),
					PlanOnly:    true,
					ExpectError: invalidEventType,
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, filterWith("indykite.audit.capture.*", "", "Car")),
					PlanOnly:    true,
					ExpectError: emptyPairField("key"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, filterWith("indykite.audit.capture.*", "captureLabel", "")),
					PlanOnly:    true,
					ExpectError: emptyPairField("value"),
				},
			},
		})
	})

	It("Test API request and response shape with multiple providers", func() {
		// providers are listed kafka-b first, so a read ordering them by name instead of by state
		// would produce a diff. Both are Kafka, so secrets must be matched by provider name.
		tfConfigDef :=
			`resource "indykite_event_sink" "development" {
				project_id = "` + appSpaceID + `"
				name = "shape-sink"
				providers {
					provider_name = "kafka-b"
					kafka {
						brokers = ["kafka-b.example.com:9092"]
						topic = "topic-b"
						username = "user-b"
						password = "password-b"
					}
				}
				providers {
					provider_name = "kafka-a"
					kafka {
						brokers = ["kafka-a.example.com:9092"]
						topic = "%s"
						username = "user-a"
						password = "password-a"
						provider_display_name = "Kafka A"
					}
				}
				routes {
					provider_id = "kafka-a"
					route_display_name = "route-a"
					route_id = "legacy-route-id"
					keys_values_filter {
						event_type = "indykite.audit.capture.*"
						key_value_pairs {
							key = "captureLabel"
							value = "Car"
						}
					}
				}
				routes {
					provider_id = "kafka-b"
					stop_processing = true
					keys_values_filter {
						event_type = "indykite.audit.config.*"
					}
				}
			}`

		var createBody, lastBody map[string]any

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Server runs in its own goroutine; GinkgoRecover lets Gomega
			// assertions in the handlers fail the spec instead of panicking silently.
			defer GinkgoRecover()
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/event-sinks"):
				var body map[string]any
				Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
				createBody, lastBody = body, body
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": sampleID})

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
				var body map[string]any
				Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
				lastBody = body
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": sampleID})

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(eventSinkAPIReadResponse(createBody, lastBody))

			case r.Method == http.MethodDelete:
				w.WriteHeader(http.StatusOK)

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

		// The routes sent to the API must use context_key_value and carry no route id.
		expectRoutesBody := func(body *map[string]any) resource.TestCheckFunc {
			return func(_ *terraform.State) error {
				return convertOmegaMatcherToError(HaveKeyWithValue("routes", ConsistOf(
					And(
						Not(HaveKey("id")),
						HaveKeyWithValue("provider_id", "kafka-a"),
						HaveKeyWithValue("event_type_key_values_filter", And(
							Not(HaveKey("key_value_pairs")),
							HaveKeyWithValue("context_key_value", ConsistOf(
								map[string]any{"key": "captureLabel", "value": "Car"})),
						)),
					),
					And(
						Not(HaveKey("id")),
						HaveKeyWithValue("provider_id", "kafka-b"),
						HaveKeyWithValue("event_type_key_values_filter", And(
							Not(HaveKey("key_value_pairs")),
							Not(HaveKey("context_key_value")),
						)),
					),
				)), *body)
			}
		}

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					// Create; the implicit plan after apply fails on any drift.
					Config: fmt.Sprintf(tfConfigDef, "topic-a"),
					Check: resource.ComposeTestCheckFunc(
						expectRoutesBody(&createBody),
						resource.TestCheckResourceAttr(resourceName, "providers.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "providers.0.provider_name", "kafka-b"),
						resource.TestCheckResourceAttr(resourceName, "providers.0.kafka.0.password", "password-b"),
						resource.TestCheckResourceAttr(resourceName, "providers.1.provider_name", "kafka-a"),
						resource.TestCheckResourceAttr(resourceName, "providers.1.kafka.0.password", "password-a"),
						resource.TestCheckResourceAttr(resourceName,
							"providers.1.kafka.0.provider_display_name", "Kafka A"),
						resource.TestCheckResourceAttr(resourceName, "routes.#", "2"),
						resource.TestCheckResourceAttr(resourceName, "routes.0.route_id", "legacy-route-id"),
						resource.TestCheckResourceAttr(resourceName, "routes.0.route_display_name", "route-a"),
						resource.TestCheckResourceAttr(resourceName, "routes.0.stop_processing", "false"),
						resource.TestCheckResourceAttr(resourceName,
							"routes.0.keys_values_filter.0.key_value_pairs.0.key", "captureLabel"),
						resource.TestCheckResourceAttr(resourceName,
							"routes.0.keys_values_filter.0.key_value_pairs.0.value", "Car"),
						resource.TestCheckResourceAttr(resourceName, "routes.1.stop_processing", "true"),
						resource.TestCheckResourceAttr(resourceName,
							"routes.1.keys_values_filter.0.key_value_pairs.#", "0"),
						resource.TestCheckResourceAttr(resourceName, "include_cdc_events", "false"),
					),
				},
				{
					Config:   fmt.Sprintf(tfConfigDef, "topic-a"),
					PlanOnly: true,
				},
				{
					// Update keeps the request shape, ordering and per-provider secrets.
					Config: fmt.Sprintf(tfConfigDef, "topic-a-updated"),
					Check: resource.ComposeTestCheckFunc(
						expectRoutesBody(&lastBody),
						resource.TestCheckResourceAttr(resourceName, "providers.0.provider_name", "kafka-b"),
						resource.TestCheckResourceAttr(resourceName, "providers.0.kafka.0.password", "password-b"),
						resource.TestCheckResourceAttr(resourceName, "providers.1.kafka.0.topic", "topic-a-updated"),
						resource.TestCheckResourceAttr(resourceName, "providers.1.kafka.0.password", "password-a"),
					),
				},
				{
					// Import has no prior state: providers come back sorted by name and secrets empty.
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: sampleID,
					ImportStateCheck: func(states []*terraform.InstanceState) error {
						if len(states) != 1 {
							return fmt.Errorf("expected 1 imported state, got %d", len(states))
						}
						return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, Keys{
							"name":                         Equal("shape-sink"),
							"providers.#":                  Equal("2"),
							"providers.0.provider_name":    Equal("kafka-a"),
							"providers.0.kafka.0.topic":    Equal("topic-a-updated"),
							"providers.0.kafka.0.password": BeEmpty(),
							"providers.1.provider_name":    Equal("kafka-b"),
							"routes.#":                     Equal("2"),
							"routes.0.route_id":            BeEmpty(),
							"routes.0.keys_values_filter.0.key_value_pairs.0.value": Equal("Car"),
							"routes.1.stop_processing":                              Equal("true"),
						}), states[0].Attributes)
					},
				},
			},
		})
	})

	It("Test provider last_error is read from the API", func() {
		fake := newFakeConfigAPI("/event-sinks", sampleID, nil)
		mockServer = fake.Server

		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
			client := indykite.NewTestRestClient(mockServer.URL+"/configs/v1", mockServer.Client())
			ctx = indykite.WithClient(ctx, client)
			return cfgFunc(ctx, data)
		}

		tfConfig := `resource "indykite_event_sink" "development" {
			project_id = "` + appSpaceID + `"
			name = "error-sink"
			providers {
				provider_name = "kafka-a"
				kafka {
					brokers = ["kafka-a.example.com:9092"]
					topic = "topic-a"
					username = "user-a"
					password = "password-a"
				}
			}
			routes {
				provider_id = "kafka-a"
				keys_values_filter {
					event_type = "indykite.audit.config.*"
				}
			}
		}`

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					Config: tfConfig,
					Check:  resource.TestCheckResourceAttr(resourceName, "providers.0.kafka.0.last_error", ""),
				},
				{
					// A delivery fails; the next refresh reads the error without any diff.
					PreConfig: func() {
						fake.update(func(stored map[string]any) {
							providers, _ := stored["providers"].(map[string]any)
							provider, _ := providers["kafka-a"].(map[string]any)
							kafka, _ := provider["kafka"].(map[string]any)
							kafka["last_error"] = "dial tcp: lookup kafka-a.example.com: no such host"
						})
					},
					Config: tfConfig,
					Check: resource.TestCheckResourceAttr(resourceName, "providers.0.kafka.0.last_error",
						"dial tcp: lookup kafka-a.example.com: no such host"),
				},
			},
		})
	})

	It("Test Kafka provider without credentials", func() {
		fake := newFakeConfigAPI("/event-sinks", sampleID, nil)
		mockServer = fake.Server

		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
			client := indykite.NewTestRestClient(mockServer.URL+"/configs/v1", mockServer.Client())
			ctx = indykite.WithClient(ctx, client)
			return cfgFunc(ctx, data)
		}

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{{
				// Brokers without authentication need no username or password.
				Config: `resource "indykite_event_sink" "development" {
					project_id = "` + appSpaceID + `"
					name = "no-auth-sink"
					providers {
						provider_name = "kafka-a"
						kafka {
							brokers = ["kafka-a.example.com:9092"]
							topic = "topic-a"
						}
					}
					routes {
						provider_id = "kafka-a"
						keys_values_filter {
							event_type = "indykite.audit.config.*"
						}
					}
				}`,
				Check: func(_ *terraform.State) error {
					var kafka map[string]any
					fake.update(func(stored map[string]any) {
						providers, _ := stored["providers"].(map[string]any)
						provider, _ := providers["kafka-a"].(map[string]any)
						kafka, _ = provider["kafka"].(map[string]any)
					})
					return convertOmegaMatcherToError(And(
						Not(HaveKey("username")), Not(HaveKey("password")),
						HaveKeyWithValue("topic", "topic-a"),
					), kafka)
				},
			}},
		})
	})
})

// eventSinkAPIReadResponse builds an event sink read response the way the Config API does:
// metadata comes from the create body and configuration from the latest create or update body.
// providers, routes and include_cdc_events are top-level, secrets are returned empty, unknown
// route fields (a route id, key_value_pairs) are dropped and a false stop_processing is omitted.
func eventSinkAPIReadResponse(created, current map[string]any) map[string]any {
	providers := map[string]any{}
	currentProviders, _ := current["providers"].(map[string]any)
	for name, p := range currentProviders {
		typed := map[string]any{}
		for providerType, cfg := range p.(map[string]any) {
			providerCfg := maps.Clone(cfg.(map[string]any))
			for _, secret := range []string{"password", "access_key", "connection_string", "credentials_json"} {
				if _, ok := providerCfg[secret]; ok {
					providerCfg[secret] = ""
				}
			}
			typed[providerType] = providerCfg
		}
		providers[name] = typed
	}

	currentRoutes, _ := current["routes"].([]any)
	routes := make([]any, 0, len(currentRoutes))
	for _, r := range currentRoutes {
		route := r.(map[string]any)
		filter, _ := route["event_type_key_values_filter"].(map[string]any)
		pairs, ok := filter["context_key_value"].([]any)
		if !ok {
			pairs = []any{}
		}
		out := map[string]any{
			"provider_id": route["provider_id"],
			"event_type_key_values_filter": map[string]any{
				"event_type":        filter["event_type"],
				"context_key_value": pairs,
			},
		}
		if v, _ := route["display_name"].(string); v != "" {
			out["display_name"] = v
		}
		if v, _ := route["stop_processing"].(bool); v {
			out["stop_processing"] = true
		}
		routes = append(routes, out)
	}

	return map[string]any{
		"id":                 sampleID,
		"name":               created["name"],
		"display_name":       created["name"],
		"description":        "",
		"organization_id":    customerID,
		"project_id":         created["project_id"],
		"create_time":        "2026-10-13T16:58:06Z",
		"update_time":        "2026-10-13T16:58:06Z",
		"providers":          providers,
		"routes":             routes,
		"include_cdc_events": current["include_cdc_events"],
	}
}

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
