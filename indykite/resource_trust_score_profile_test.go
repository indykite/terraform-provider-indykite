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
	"strconv"
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

var _ = Describe("Resource TrustScoreProfile", func() {
	const (
		resourceName = "indykite_trust_score_profile.development"
	)
	var (
		mockServer  *httptest.Server
		provider    *schema.Provider
		tfConfigDef string
	)

	BeforeEach(func() {
		provider = indykite.Provider()

		tfConfigDef = `resource "indykite_trust_score_profile" "development" {
			location = "%s"
			name = "%s"
			%s
		}`
	})

	AfterEach(func() {
		if mockServer != nil {
			mockServer.Close()
		}
	})

	Describe("Error cases", func() {
		It("should handle invalid configurations", func() {
			resource.Test(GinkgoT(), resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"indykite": func() (*schema.Provider, error) { return provider, nil },
				},
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(tfConfigDef, "ccc", "name",
							`node_classification = "Person"
							schedule = "UPDATE_FREQUENCY_DAILY"
							dimension {
								name = "NAME_FRESHNESS"
								weight = 0.6
							}`),
						ExpectError: regexp.MustCompile("Invalid ID value"),
					},
					{
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name",
							`node_classification = "Person"
							`),
						ExpectError: regexp.MustCompile(
							`The argument "schedule" is required, but no definition was found.`),
					},
				},
			})
		})

		DescribeTable("schedule is validated at plan time",
			func(schedule string, valid bool) {
				// Plan only: any request reaching the API fails the test.
				mockServer = httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()
					Fail(fmt.Sprintf("unexpected API call %s %s", r.Method, r.URL.Path))
				}))
				cfgFunc := provider.ConfigureContextFunc
				provider.ConfigureContextFunc = func(
					ctx context.Context, data *schema.ResourceData,
				) (any, diag.Diagnostics) {
					client := indykite.NewTestRestClient(mockServer.URL+"/configs/v1", mockServer.Client())
					ctx = indykite.WithClient(ctx, client)
					return cfgFunc(ctx, data)
				}

				step := resource.TestStep{
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name", fmt.Sprintf(
						`node_classification = "Person"
						schedule = %q
						dimension {
							name = "NAME_FRESHNESS"
							weight = 0.6
						}`, schedule)),
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				}
				if !valid {
					step.ExpectError = regexp.MustCompile(`expected schedule to be one of`)
				}
				resource.Test(GinkgoT(), resource.TestCase{
					ProviderFactories: map[string]func() (*schema.Provider, error){
						"indykite": func() (*schema.Provider, error) { return provider, nil },
					},
					Steps: []resource.TestStep{step},
				})
			},
			Entry("three hours", "UPDATE_FREQUENCY_THREE_HOURS", true),
			Entry("six hours", "UPDATE_FREQUENCY_SIX_HOURS", true),
			Entry("twelve hours", "UPDATE_FREQUENCY_TWELVE_HOURS", true),
			Entry("daily", "UPDATE_FREQUENCY_DAILY", true),
			Entry("invalid, rejected by the API", "UPDATE_FREQUENCY_INVALID", false),
			Entry("hourly, not supported", "UPDATE_FREQUENCY_HOURLY", false),
			Entry("API value instead of Terraform value", "DAILY", false),
			Entry("empty", "", false),
		)
	})

	Describe("Valid configurations", func() {
		It("Test CRUD of TrustScoreProfile configuration", func() {
			createTime := time.Now()
			updateTime := time.Now()

			// Track whether update has been called
			updated := false

			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/trust-score-profiles"):
					resp := indykite.TrustScoreProfileResponse{
						ID:                 sampleID,
						Name:               "my-first-trust-score-profile",
						DisplayName:        "Display name of TrustScoreProfile configuration",
						CustomerID:         customerID,
						AppSpaceID:         appSpaceID,
						NodeClassification: "Person",
						Dimensions: []*indykite.TrustScoreDimension{
							{Name: "NAME_FRESHNESS", Weight: 0.7},
							{Name: "NAME_ORIGIN", Weight: 0.7},
						},
						Schedule:   "UPDATE_FREQUENCY_DAILY",
						CreateTime: createTime,
						UpdateTime: updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)

				case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
					var resp indykite.TrustScoreProfileResponse
					if updated {
						// Return updated state
						resp = indykite.TrustScoreProfileResponse{
							ID:                 sampleID,
							Name:               "my-first-trust-score-profile",
							DisplayName:        "Display name of TrustScoreProfile configuration",
							CustomerID:         customerID,
							AppSpaceID:         appSpaceID,
							NodeClassification: "Person",
							Dimensions: []*indykite.TrustScoreDimension{
								{Name: "NAME_COMPLETENESS", Weight: 0.9},
								{Name: "NAME_ORIGIN", Weight: 0.9},
							},
							Schedule:   "UPDATE_FREQUENCY_SIX_HOURS",
							CreateTime: createTime,
							UpdateTime: time.Now(),
						}
					} else {
						// Return initial state
						resp = indykite.TrustScoreProfileResponse{
							ID:                 sampleID,
							Name:               "my-first-trust-score-profile",
							DisplayName:        "Display name of TrustScoreProfile configuration",
							CustomerID:         customerID,
							AppSpaceID:         appSpaceID,
							NodeClassification: "Person",
							Dimensions: []*indykite.TrustScoreDimension{
								{Name: "NAME_FRESHNESS", Weight: 0.7},
								{Name: "NAME_ORIGIN", Weight: 0.7},
							},
							Schedule:   "UPDATE_FREQUENCY_DAILY",
							CreateTime: createTime,
							UpdateTime: updateTime,
						}
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)

				case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
					updated = true
					resp := indykite.TrustScoreProfileResponse{
						ID:                 sampleID,
						Name:               "my-first-trust-score-profile",
						DisplayName:        "Display name of TrustScoreProfile configuration",
						CustomerID:         customerID,
						AppSpaceID:         appSpaceID,
						NodeClassification: "Person",
						Dimensions: []*indykite.TrustScoreDimension{
							{Name: "NAME_COMPLETENESS", Weight: 0.9},
							{Name: "NAME_ORIGIN", Weight: 0.9},
						},
						Schedule:   "UPDATE_FREQUENCY_SIX_HOURS",
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
			provider.ConfigureContextFunc = func(
				ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
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
						// Checking Create and Read
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-trust-score-profile",
							`display_name = "Display name of TrustScoreProfile configuration"
							node_classification = "Person"
							dimension {
								name   = "NAME_FRESHNESS"
								weight = 0.7
							}
							dimension {
								name   = "NAME_ORIGIN"
								weight = 0.7
							}
							schedule = "UPDATE_FREQUENCY_DAILY"
							`,
						),
						Check: resource.ComposeTestCheckFunc(
							testResourceTrustScoreProfileDataExists(resourceName, sampleID),
						),
					},
					{
						// Checking Update and Read
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-trust-score-profile",
							`display_name = "Display name of TrustScoreProfile configuration"
							node_classification = "Person"
							dimension {
							name   = "NAME_COMPLETENESS"
							weight = 0.9
							}
							dimension {
								name   = "NAME_ORIGIN"
								weight = 0.9
							}
							schedule = "UPDATE_FREQUENCY_SIX_HOURS"
							`),
						Check: resource.ComposeTestCheckFunc(
							testResourceTrustScoreProfileDataExists(resourceName, sampleID),
						),
					},
				},
			})
		})

		It("Test import by ID", func() {
			tfConfigDef := `resource "indykite_trust_score_profile" "development" {
				location = "%s"
				name = "%s"
				%s
			}`

			createTime := time.Now()
			updateTime := time.Now()

			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/trust-score-profiles"):
					resp := indykite.TrustScoreProfileResponse{
						ID:                 sampleID,
						Name:               "wonka-trust-score",
						DisplayName:        "Wonka Trust Score",
						CustomerID:         customerID,
						AppSpaceID:         appSpaceID,
						NodeClassification: "Person",
						Dimensions: []*indykite.TrustScoreDimension{
							{Name: "NAME_FRESHNESS", Weight: 0.7},
							{Name: "NAME_ORIGIN", Weight: 0.7},
						},
						Schedule:   "UPDATE_FREQUENCY_DAILY",
						CreateTime: createTime,
						UpdateTime: updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)

				case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
					resp := indykite.TrustScoreProfileResponse{
						ID:                 sampleID,
						Name:               "wonka-trust-score",
						DisplayName:        "Wonka Trust Score",
						CustomerID:         customerID,
						AppSpaceID:         appSpaceID,
						NodeClassification: "Person",
						Dimensions: []*indykite.TrustScoreDimension{
							{Name: "NAME_FRESHNESS", Weight: 0.7},
							{Name: "NAME_ORIGIN", Weight: 0.7},
						},
						Schedule:   "UPDATE_FREQUENCY_DAILY",
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
			provider.ConfigureContextFunc = func(
				ctx context.Context,
				data *schema.ResourceData,
			) (any, diag.Diagnostics) {
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
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "wonka-trust-score",
							`display_name = "Wonka Trust Score"
							node_classification = "Person"
							dimension {
								name   = "NAME_FRESHNESS"
								weight = 0.7
							}
							dimension {
								name   = "NAME_ORIGIN"
								weight = 0.7
							}
							schedule = "UPDATE_FREQUENCY_DAILY"
							`,
						),
						Check: resource.ComposeTestCheckFunc(
							testResourceTrustScoreProfileDataExists(resourceName, sampleID),
						),
					},
					{
						ResourceName:  resourceName,
						ImportState:   true,
						ImportStateId: sampleID,
					},
				},
			})
		})

		It("Test import by name with location", func() {
			tfConfigDef := `resource "indykite_trust_score_profile" "development" {
				location = "%s"
				name = "%s"
				%s
			}`

			createTime := time.Now()
			updateTime := time.Now()

			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/trust-score-profiles"):
					resp := indykite.TrustScoreProfileResponse{
						ID:                 sampleID,
						Name:               "wonka-trust-score",
						DisplayName:        "Wonka Trust Score",
						CustomerID:         customerID,
						AppSpaceID:         appSpaceID,
						NodeClassification: "Person",
						Dimensions: []*indykite.TrustScoreDimension{
							{Name: "NAME_FRESHNESS", Weight: 0.7},
							{Name: "NAME_ORIGIN", Weight: 0.7},
						},
						Schedule:   "UPDATE_FREQUENCY_DAILY",
						CreateTime: createTime,
						UpdateTime: updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)

				case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/trust-score-profiles/"):
					// Support both ID and name?location=appSpaceID formats
					pathAfterProfiles := strings.TrimPrefix(r.URL.Path, "/configs/v1/trust-score-profiles/")
					isNameLookup := strings.Contains(pathAfterProfiles, "wonka-trust-score")
					isIDLookup := strings.Contains(pathAfterProfiles, sampleID)

					if isNameLookup || isIDLookup {
						resp := indykite.TrustScoreProfileResponse{
							ID:                 sampleID,
							Name:               "wonka-trust-score",
							DisplayName:        "Wonka Trust Score",
							CustomerID:         customerID,
							AppSpaceID:         appSpaceID,
							NodeClassification: "Person",
							Dimensions: []*indykite.TrustScoreDimension{
								{Name: "NAME_FRESHNESS", Weight: 0.7},
								{Name: "NAME_ORIGIN", Weight: 0.7},
							},
							Schedule:   "UPDATE_FREQUENCY_DAILY",
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
			provider.ConfigureContextFunc = func(
				ctx context.Context,
				data *schema.ResourceData,
			) (any, diag.Diagnostics) {
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
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "wonka-trust-score",
							`display_name = "Wonka Trust Score"
							node_classification = "Person"
							dimension {
								name   = "NAME_FRESHNESS"
								weight = 0.7
							}
							dimension {
								name   = "NAME_ORIGIN"
								weight = 0.7
							}
							schedule = "UPDATE_FREQUENCY_DAILY"
							`,
						),
						Check: resource.ComposeTestCheckFunc(
							testResourceTrustScoreProfileDataExists(resourceName, sampleID),
						),
					},
					{
						ResourceName:  resourceName,
						ImportState:   true,
						ImportStateId: "wonka-trust-score?location=" + appSpaceID,
					},
				},
			})
		})
	})

	It("Test last run fields are read from the API", func() {
		fake := newFakeConfigAPI("/trust-score-profiles", sampleID, nil)
		mockServer = fake.Server

		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
			client := indykite.NewTestRestClient(mockServer.URL+"/configs/v1", mockServer.Client())
			ctx = indykite.WithClient(ctx, client)
			return cfgFunc(ctx, data)
		}

		tfConfig := `resource "indykite_trust_score_profile" "development" {
			project_id = "` + appSpaceID + `"
			name = "run-profile"
			node_classification = "Person"
			schedule = "UPDATE_FREQUENCY_DAILY"
			dimension {
				name = "NAME_FRESHNESS"
				weight = 0.5
			}
		}`

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					Config: tfConfig,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "last_run_id", ""),
						resource.TestCheckResourceAttr(resourceName, "dimensions_execution_times.#", "0"),
					),
				},
				{
					// The API runs the profile; the next refresh reads the run without any diff. The
					// windows come back as a map, so they are sorted and use the Terraform dimension names.
					PreConfig: func() {
						fake.update(func(stored map[string]any) {
							stored["last_run_id"] = "run-1"
							stored["last_run_start_time"] = "2026-10-15T08:00:00Z"
							stored["last_run_end_time"] = "2026-10-15T08:05:00Z"
							stored["dimensions_execution_times"] = map[string]any{
								"VERIFICATION": map[string]any{
									"execution_start_time": "2026-10-15T08:03:00Z",
									"execution_end_time":   "2026-10-15T08:05:00Z",
								},
								"FRESHNESS": map[string]any{
									"execution_start_time": "2026-10-15T08:00:00Z",
									"execution_end_time":   "2026-10-15T08:02:00Z",
								},
								"ORIGIN": nil,
							}
						})
					},
					Config: tfConfig,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "last_run_id", "run-1"),
						resource.TestCheckResourceAttr(resourceName, "last_run_start_time", "2026-10-15T08:00:00Z"),
						resource.TestCheckResourceAttr(resourceName, "last_run_end_time", "2026-10-15T08:05:00Z"),
						resource.TestCheckResourceAttr(resourceName, "dimensions_execution_times.#", "3"),
						resource.TestCheckResourceAttr(resourceName,
							"dimensions_execution_times.0.dimension", "NAME_FRESHNESS"),
						resource.TestCheckResourceAttr(resourceName,
							"dimensions_execution_times.0.execution_end_time", "2026-10-15T08:02:00Z"),
						resource.TestCheckResourceAttr(resourceName,
							"dimensions_execution_times.1.dimension", "NAME_ORIGIN"),
						resource.TestCheckResourceAttr(resourceName,
							"dimensions_execution_times.1.execution_start_time", ""),
						resource.TestCheckResourceAttr(resourceName,
							"dimensions_execution_times.2.dimension", "NAME_VERIFICATION"),
						resource.TestCheckResourceAttr(resourceName,
							"dimensions_execution_times.2.execution_start_time", "2026-10-15T08:03:00Z"),
					),
				},
			},
		})
	})
})

//nolint:unparam // Test helper function designed to be reusable
func testResourceTrustScoreProfileDataExists(
	n string,
	expectedID string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != expectedID {
			return errors.New("ID does not match")
		}
		attrs := rs.Primary.Attributes

		keys := Keys{
			"id": Equal(expectedID),
			"%":  Not(BeEmpty()),

			"location":            Equal(appSpaceID),
			"customer_id":         Equal(customerID),
			"app_space_id":        Equal(appSpaceID),
			"name":                Not(BeEmpty()),
			"display_name":        Not(BeEmpty()),
			"create_time":         Not(BeEmpty()),
			"update_time":         Not(BeEmpty()),
			"node_classification": Not(BeEmpty()),
			"schedule":            Not(BeEmpty()),
		}

		// Check dimension count
		dimensionCount := attrs["dimension.#"]
		if dimensionCount != "" {
			count, _ := strconv.Atoi(dimensionCount)
			keys["dimension.#"] = Equal(strconv.Itoa(count))
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
	}
}
