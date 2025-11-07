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

var _ = Describe("Resource ExternalDataResolver", func() {
	const (
		resourceName = "indykite_external_data_resolver.development"
	)
	var (
		mockServer    *httptest.Server
		provider      *schema.Provider
		tfConfigDef   string
		validSettings string
	)

	BeforeEach(func() {
		provider = indykite.Provider()

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

	AfterEach(func() {
		if mockServer != nil {
			mockServer.Close()
		}
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
			createTime := time.Now()
			updateTime := time.Now()
			currentState := "initial"

			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/external-data-resolvers"):
					resp := indykite.ExternalDataResolverResponse{
						ID:          sampleID,
						Name:        "my-first-external-data-resolver1",
						DisplayName: "Display name of ExternalDataResolver configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						URL:         "https://example.com/source2",
						Method:      "GET",
						Headers: map[string]any{
							"Authorization": []string{"Bearer edolkUTY"},
							"Content-Type":  []string{"application/json"},
						},
						RequestType:      "json",
						RequestPayload:   `{"key": "value"}`,
						ResponseType:     "json",
						ResponseSelector: ".",
						CreateTime:       createTime,
						UpdateTime:       updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)

				case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
					var resp indykite.ExternalDataResolverResponse

					switch currentState {
					case "initial", "after_create":
						resp = indykite.ExternalDataResolverResponse{
							ID:          sampleID,
							Name:        "my-first-external-data-resolver1",
							DisplayName: "Display name of ExternalDataResolver configuration",
							CustomerID:  customerID,
							AppSpaceID:  appSpaceID,
							URL:         "https://example.com/source2",
							Method:      "GET",
							Headers: map[string]any{
								"Authorization": []string{"Bearer edolkUTY"},
								"Content-Type":  []string{"application/json"},
							},
							RequestType:      "json",
							RequestPayload:   `{"key": "value"}`,
							ResponseType:     "json",
							ResponseSelector: ".",
							CreateTime:       createTime,
							UpdateTime:       updateTime,
						}
					case "after_update1":
						resp = indykite.ExternalDataResolverResponse{
							ID:          sampleID,
							Name:        "my-first-external-data-resolver1",
							DisplayName: "Display name of ExternalDataResolver configuration",
							CustomerID:  customerID,
							AppSpaceID:  appSpaceID,
							URL:         "https://example.com/source2",
							Method:      "GET",
							Headers: map[string]any{
								"Authorization": []string{"Bearer edolkUTY"},
								"Content-Type":  []string{"application/json"},
							},
							RequestType:      "json",
							RequestPayload:   `{"key2": "value2"}`,
							ResponseType:     "json",
							ResponseSelector: ".",
							CreateTime:       createTime,
							UpdateTime:       time.Now(),
						}
					case "after_update2":
						resp = indykite.ExternalDataResolverResponse{
							ID:         sampleID,
							Name:       "my-first-external-data-resolver1",
							CustomerID: customerID,
							AppSpaceID: appSpaceID,
							URL:        "https://example.com/source2",
							Method:     "GET",
							Headers: map[string]any{
								"Authorization": []string{"Bearer edolkUTY"},
								"Content-Type":  []string{"application/json"},
							},
							RequestType:      "json",
							ResponseType:     "json",
							ResponseSelector: ".",
							CreateTime:       createTime,
							UpdateTime:       time.Now(),
						}
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)

				case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
					// Parse request to determine state
					var reqBody map[string]any
					_ = json.NewDecoder(r.Body).Decode(&reqBody)

					if reqBody["requestPayload"] != nil &&
						strings.Contains(reqBody["requestPayload"].(string), "key2") {
						currentState = "after_update1"
					} else {
						currentState = "after_update2"
					}

					var resp indykite.ExternalDataResolverResponse
					if currentState == "after_update1" {
						resp = indykite.ExternalDataResolverResponse{
							ID:          sampleID,
							Name:        "my-first-external-data-resolver1",
							DisplayName: "Display name of ExternalDataResolver configuration",
							CustomerID:  customerID,
							AppSpaceID:  appSpaceID,
							URL:         "https://example.com/source2",
							Method:      "GET",
							Headers: map[string]any{
								"Authorization": []string{"Bearer edolkUTY"},
								"Content-Type":  []string{"application/json"},
							},
							RequestType:      "json",
							RequestPayload:   `{"key2": "value2"}`,
							ResponseType:     "json",
							ResponseSelector: ".",
							CreateTime:       createTime,
							UpdateTime:       time.Now(),
						}
					} else {
						resp = indykite.ExternalDataResolverResponse{
							ID:         sampleID,
							Name:       "my-first-external-data-resolver1",
							CustomerID: customerID,
							AppSpaceID: appSpaceID,
							URL:        "https://example.com/source2",
							Method:     "GET",
							Headers: map[string]any{
								"Authorization": []string{"Bearer edolkUTY"},
								"Content-Type":  []string{"application/json"},
							},
							RequestType:      "json",
							ResponseType:     "json",
							ResponseSelector: ".",
							CreateTime:       createTime,
							UpdateTime:       time.Now(),
						}
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
			provider.ConfigureContextFunc = func(
				ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
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
							request_payload  = "{\"key\": \"value\"}"
							response_type = "json"
							response_selector = "."
							`,
						),
						Check: resource.ComposeTestCheckFunc(
							testExternalDataResolverResourceDataExists(resourceName, sampleID),
						),
					},
					{
						// Import test
						ResourceName:  resourceName,
						ImportState:   true,
						ImportStateId: sampleID,
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
							request_payload  = "{\"key2\": \"value2\"}"
							response_type = "json"
							response_selector = "."
							`,
						),
						Check: resource.ComposeTestCheckFunc(
							testExternalDataResolverResourceDataExists(resourceName, sampleID),
						),
					},
					{
						// Checking Update and Read - remove optional fields
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-external-data-resolver1",
							`url = "https://example.com/source2"
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
							testExternalDataResolverResourceDataExists(resourceName, sampleID),
						),
					},
				},
			})
		})

		It("Test import by name with location", func() {
			tfConfigDef := `resource "indykite_external_data_resolver" "development" {
					location = "%s"
					name = "%s"
					%s
				}`

			createTime := time.Now()
			updateTime := time.Now()

			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/external-data-resolvers"):
					resp := indykite.ExternalDataResolverResponse{
						ID:               sampleID,
						Name:             "wonka-resolver",
						DisplayName:      "Wonka Resolver",
						CustomerID:       customerID,
						AppSpaceID:       appSpaceID,
						URL:              "https://example.com/source",
						Method:           "POST",
						Headers:          map[string]any{"Authorization": []any{"Bearer token"}},
						RequestType:      "JSON",
						RequestPayload:   `{"key": "value"}`,
						ResponseType:     "JSON",
						ResponseSelector: ".",
						CreateTime:       createTime,
						UpdateTime:       updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)

				case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/external-data-resolvers/"):
					// Support both ID and name?location=appSpaceID formats
					pathAfterResolvers := strings.TrimPrefix(r.URL.Path, "/configs/v1/external-data-resolvers/")
					isNameLookup := strings.Contains(pathAfterResolvers, "wonka-resolver")
					isIDLookup := strings.Contains(pathAfterResolvers, sampleID)

					if isNameLookup || isIDLookup {
						resp := indykite.ExternalDataResolverResponse{
							ID:               sampleID,
							Name:             "wonka-resolver",
							DisplayName:      "Wonka Resolver",
							CustomerID:       customerID,
							AppSpaceID:       appSpaceID,
							URL:              "https://example.com/source",
							Method:           "POST",
							Headers:          map[string]any{"Authorization": []any{"Bearer token"}},
							RequestType:      "JSON",
							RequestPayload:   `{"key": "value"}`,
							ResponseType:     "JSON",
							ResponseSelector: ".",
							CreateTime:       createTime,
							UpdateTime:       updateTime,
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
			provider.ConfigureContextFunc = func(
				ctx context.Context,
				data *schema.ResourceData,
			) (any, diag.Diagnostics) {
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
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "wonka-resolver",
							`display_name = "Wonka Resolver"
								url = "https://example.com/source"
								method = "POST"
								headers {
									name = "Authorization"
									values = ["Bearer token"]
								}
								request_type = "json"
								request_payload = "{\"key\": \"value\"}"
								response_type = "json"
								response_selector = "."
								`,
						),
						Check: resource.ComposeTestCheckFunc(
							testExternalDataResolverResourceDataExists(resourceName, sampleID),
						),
					},
					{
						ResourceName:  resourceName,
						ImportState:   true,
						ImportStateId: "wonka-resolver?location=" + appSpaceID,
					},
				},
			})
		})
	})
})

//nolint:unparam // Test helper function designed to be reusable
func testExternalDataResolverResourceDataExists(n, expectedID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID != expectedID {
			return errors.New("ID does not match")
		}

		keys := Keys{
			"id": Equal(expectedID),
			"%":  Not(BeEmpty()),

			"customer_id":  Equal(customerID),
			"app_space_id": Equal(appSpaceID),
			"name":         Not(BeEmpty()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), rs.Primary.Attributes)
	}
}
