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

var _ = Describe("Resource EntityMatchingPipeline", func() {
	const (
		resourceName = "indykite_entity_matching_pipeline.development"
	)
	var (
		mockServer    *httptest.Server
		provider      *schema.Provider
		tfConfigDef   string
		validSettings string
	)

	BeforeEach(func() {
		provider = indykite.Provider()

		tfConfigDef = `resource "indykite_entity_matching_pipeline" "development" {
			location = "%s"
			name = "%s"
			%s
		}`

		validSettings = `
		source_node_filter = ["Person"]
		target_node_filter = ["Person"]
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
							`target_node_filter = ["Person"]
							`),
						ExpectError: regexp.MustCompile(
							`The argument "source_node_filter" is required, but no definition was found.`),
					},
				},
			})
		})
	})

	Describe("Valid configurations", func() {
		It("Test CRUD of EntityMatchingPipeline configuration", func() {
			createTime := time.Now()
			updateTime := time.Now()
			updated := false

			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/entity-matching-pipelines"):
					resp := indykite.EntityMatchingPipelineResponse{
						ID:          sampleID,
						Name:        "my-first-entity-matching-pipeline",
						DisplayName: "Display name of EntityMatchingPipeline configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						NodeFilter: &indykite.EntityMatchingNodeFilter{
							SourceNodeTypes: []string{"Person"},
							TargetNodeTypes: []string{"Person"},
						},
						SimilarityScoreCutoff: 0.7,
						CreateTime:            createTime,
						UpdateTime:            updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)

				case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
					var resp indykite.EntityMatchingPipelineResponse
					if updated {
						// After update: include rerun_interval
						resp = indykite.EntityMatchingPipelineResponse{
							ID:          sampleID,
							Name:        "my-first-entity-matching-pipeline",
							DisplayName: "Display name of EntityMatchingPipeline configuration",
							CustomerID:  customerID,
							AppSpaceID:  appSpaceID,
							NodeFilter: &indykite.EntityMatchingNodeFilter{
								SourceNodeTypes: []string{"Person"},
								TargetNodeTypes: []string{"Person"},
							},
							SimilarityScoreCutoff: 0.7,
							RerunInterval:         "1 day",
							CreateTime:            createTime,
							UpdateTime:            time.Now(),
						}
					} else {
						// Before update: no rerun_interval (same as POST)
						resp = indykite.EntityMatchingPipelineResponse{
							ID:          sampleID,
							Name:        "my-first-entity-matching-pipeline",
							DisplayName: "Display name of EntityMatchingPipeline configuration",
							CustomerID:  customerID,
							AppSpaceID:  appSpaceID,
							NodeFilter: &indykite.EntityMatchingNodeFilter{
								SourceNodeTypes: []string{"Person"},
								TargetNodeTypes: []string{"Person"},
							},
							SimilarityScoreCutoff: 0.7,
							CreateTime:            createTime,
							UpdateTime:            updateTime,
						}
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)

				case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
					updated = true
					resp := indykite.EntityMatchingPipelineResponse{
						ID:          sampleID,
						Name:        "my-first-entity-matching-pipeline",
						DisplayName: "Display name of EntityMatchingPipeline configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						NodeFilter: &indykite.EntityMatchingNodeFilter{
							SourceNodeTypes: []string{"Person"},
							TargetNodeTypes: []string{"Person"},
						},
						SimilarityScoreCutoff: 0.7,
						RerunInterval:         "1 day",
						CreateTime:            createTime,
						UpdateTime:            time.Now(),
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
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-entity-matching-pipeline",
							`display_name = "Display name of EntityMatchingPipeline configuration"
							source_node_filter = ["Person"]
							target_node_filter = ["Person"]
							similarity_score_cutoff =  0.7
							`,
						),
						Check: resource.ComposeTestCheckFunc(
							testResourceEntityMatchingPipelineExists(resourceName, sampleID),
						),
					},
					{
						// Checking Update and Read
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-entity-matching-pipeline",
							`display_name = "Display name of EntityMatchingPipeline configuration"
							source_node_filter = ["Person"]
							target_node_filter = ["Person"]
							similarity_score_cutoff =  0.7
							rerun_interval = "1 day"
							`),
						Check: resource.ComposeTestCheckFunc(
							testResourceEntityMatchingPipelineExists(resourceName, sampleID),
						),
					},
				},
			})
		})

		It("Test import by ID", func() {
			createTime := time.Now()
			updateTime := time.Now()

			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/entity-matching-pipelines"):
					resp := indykite.EntityMatchingPipelineResponse{
						ID:          sampleID,
						Name:        "wonka-pipeline",
						DisplayName: "Wonka Pipeline",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						NodeFilter: &indykite.EntityMatchingNodeFilter{
							SourceNodeTypes: []string{"Person"},
							TargetNodeTypes: []string{"Person"},
						},
						SimilarityScoreCutoff: 0.7,
						CreateTime:            createTime,
						UpdateTime:            updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)

				case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
					resp := indykite.EntityMatchingPipelineResponse{
						ID:          sampleID,
						Name:        "wonka-pipeline",
						DisplayName: "Wonka Pipeline",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						NodeFilter: &indykite.EntityMatchingNodeFilter{
							SourceNodeTypes: []string{"Person"},
							TargetNodeTypes: []string{"Person"},
						},
						SimilarityScoreCutoff: 0.7,
						CreateTime:            createTime,
						UpdateTime:            updateTime,
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
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "wonka-pipeline",
							`display_name = "Wonka Pipeline"
								source_node_filter = ["Person"]
								target_node_filter = ["Person"]
								similarity_score_cutoff = 0.7
								`,
						),
						Check: resource.ComposeTestCheckFunc(
							testResourceEntityMatchingPipelineExists(resourceName, sampleID),
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
			createTime := time.Now()
			updateTime := time.Now()

			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/entity-matching-pipelines"):
					resp := indykite.EntityMatchingPipelineResponse{
						ID:          sampleID,
						Name:        "wonka-pipeline",
						DisplayName: "Wonka Pipeline",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						NodeFilter: &indykite.EntityMatchingNodeFilter{
							SourceNodeTypes: []string{"Person"},
							TargetNodeTypes: []string{"Person"},
						},
						SimilarityScoreCutoff: 0.7,
						CreateTime:            createTime,
						UpdateTime:            updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)

				case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/entity-matching-pipelines/"):
					// Support both ID and name?location=appSpaceID formats
					pathAfterPipelines := strings.TrimPrefix(r.URL.Path, "/configs/v1/entity-matching-pipelines/")
					isNameLookup := strings.Contains(pathAfterPipelines, "wonka-pipeline")
					isIDLookup := strings.Contains(pathAfterPipelines, sampleID)

					if isNameLookup || isIDLookup {
						resp := indykite.EntityMatchingPipelineResponse{
							ID:          sampleID,
							Name:        "wonka-pipeline",
							DisplayName: "Wonka Pipeline",
							CustomerID:  customerID,
							AppSpaceID:  appSpaceID,
							NodeFilter: &indykite.EntityMatchingNodeFilter{
								SourceNodeTypes: []string{"Person"},
								TargetNodeTypes: []string{"Person"},
							},
							SimilarityScoreCutoff: 0.7,
							CreateTime:            createTime,
							UpdateTime:            updateTime,
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
						Config: fmt.Sprintf(tfConfigDef, appSpaceID, "wonka-pipeline",
							`display_name = "Wonka Pipeline"
								source_node_filter = ["Person"]
								target_node_filter = ["Person"]
								similarity_score_cutoff = 0.7
								`,
						),
						Check: resource.ComposeTestCheckFunc(
							testResourceEntityMatchingPipelineExists(resourceName, sampleID),
						),
					},
					{
						ResourceName:  resourceName,
						ImportState:   true,
						ImportStateId: "wonka-pipeline?location=" + appSpaceID,
					},
				},
			})
		})
	})
})

//nolint:unparam // Test helper function designed to be reusable
func testResourceEntityMatchingPipelineExists(
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

			"location":     Equal(appSpaceID),
			"customer_id":  Equal(customerID),
			"app_space_id": Equal(appSpaceID),
			"name":         Not(BeEmpty()),
			"display_name": Not(BeEmpty()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
		}

		// Check source_node_filter count
		if sourceCount := attrs["source_node_filter.#"]; sourceCount != "" {
			count, _ := strconv.Atoi(sourceCount)
			keys["source_node_filter.#"] = Equal(strconv.Itoa(count))
		}

		// Check target_node_filter count
		if targetCount := attrs["target_node_filter.#"]; targetCount != "" {
			count, _ := strconv.Atoi(targetCount)
			keys["target_node_filter.#"] = Equal(strconv.Itoa(count))
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
	}
}
