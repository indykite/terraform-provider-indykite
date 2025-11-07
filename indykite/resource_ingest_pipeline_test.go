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

var _ = Describe("Resource IngestPipeline", func() {
	//nolint:gosec,lll // there are no secrets
	const (
		resourceName  = "indykite_ingest_pipeline.development"
		appAgentToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaWQ6QUFBQUJXbHVaSGxyYVVSbGdBQUZEd0FBQUFBIiwic3ViIjoiZ2lkOkFBQUFCV2x1WkhscmFVUmxnQUFGRHdBQUFBQSIsImV4cCI6MjUzNDAyMjYxMTk5LCJpYXQiOjE1MTYyMzkwMjJ9.39Kc7pL8Vjf1S4qA6NHBGMP06TahR5Y9JOGSWKOo5Rw" // checkov:skip=CKV_SECRET_9:acceptance test // gitleaks:allow //nolint:lll
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

	It("Test CRUD of Ingest Pipeline configuration", func() {
		tfConfigDef :=
			`resource "indykite_ingest_pipeline" "development" {
				location = "%s"
				name = "%s"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()
		updated := false

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/ingest-pipelines"):
				resp := indykite.IngestPipelineResponse{
					ID:          sampleID,
					Name:        "my-first-ingest-pipeline",
					DisplayName: "Display name of Ingest Pipeline configuration",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Sources:     []string{"source1", "source2"},
					CreateTime:  createTime,
					UpdateTime:  updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
				// Return different data based on whether we've been updated
				var resp indykite.IngestPipelineResponse
				if updated {
					// After update: return description and 3 sources
					resp = indykite.IngestPipelineResponse{
						ID:          sampleID,
						Name:        "my-first-ingest-pipeline",
						Description: "ingest pipeline description",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Sources:     []string{"source1", "source2", "source3"},
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				} else {
					// Before update: return display_name and 2 sources (same as POST)
					resp = indykite.IngestPipelineResponse{
						ID:          sampleID,
						Name:        "my-first-ingest-pipeline",
						DisplayName: "Display name of Ingest Pipeline configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Sources:     []string{"source1", "source2"},
						CreateTime:  createTime,
						UpdateTime:  updateTime,
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
				updated = true
				resp := indykite.IngestPipelineResponse{
					ID:          sampleID,
					Name:        "my-first-ingest-pipeline",
					Description: "ingest pipeline description",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Sources:     []string{"source1", "source2", "source3"},
					CreateTime:  createTime,
					UpdateTime:  time.Now(),
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

		validSettings := `
		sources = ["source1", "source2"]
		app_agent_token = "` + appAgentToken + `"
		`

		validSettingsUpdate := `
		sources = ["source1", "source2", "source3"]
		app_agent_token = "` + appAgentToken + `"
		`

		// Invalid token
		invalidSettings1 := `
		sources = ["source1", "source2"]
		app_agent_token = "invalid-token"
		` // checkov:skip=CKV_SECRET_6:acceptance test

		// Missing required argument
		invalidSettings2 := `
		app_agent_token = "` + appAgentToken + `"
		`

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors case must be always first
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", invalidSettings1),
					ExpectError: regexp.MustCompile("invalid value for app_agent_token"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, "ccc", "name", validSettings),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", invalidSettings2),
					ExpectError: regexp.MustCompile("Missing required argument"),
				},

				{
					// Checking Create and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-ingest-pipeline",
						`display_name = "Display name of Ingest Pipeline configuration"
						`+validSettings+``,
					),
					Check: resource.ComposeTestCheckFunc(
						testIngestPipelineResourceDataExists(resourceName, sampleID, appAgentToken),
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
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-ingest-pipeline",
						`description = "ingest pipeline description"
						`+validSettingsUpdate+``,
					),
					Check: resource.ComposeTestCheckFunc(
						testIngestPipelineResourceDataExists(resourceName, sampleID, appAgentToken),
					),
				},
			},
		})
	})

	It("Test import by name with location", func() {
		tfConfigDef := `resource "indykite_ingest_pipeline" "development" {
				location = "%s"
				name = "%s"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/ingest-pipelines"):
				resp := indykite.IngestPipelineResponse{
					ID:          sampleID,
					Name:        "wonka-ingest",
					DisplayName: "Wonka Ingest",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Sources:     []string{"source1", "source2"},
					CreateTime:  createTime,
					UpdateTime:  updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/ingest-pipelines/"):
				// Support both ID and name?location=appSpaceID formats
				pathAfterPipelines := strings.TrimPrefix(r.URL.Path, "/configs/v1/ingest-pipelines/")
				isNameLookup := strings.Contains(pathAfterPipelines, "wonka-ingest")
				isIDLookup := strings.Contains(pathAfterPipelines, sampleID)

				if isNameLookup || isIDLookup {
					resp := indykite.IngestPipelineResponse{
						ID:          sampleID,
						Name:        "wonka-ingest",
						DisplayName: "Wonka Ingest",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Sources:     []string{"source1", "source2"},
						CreateTime:  createTime,
						UpdateTime:  updateTime,
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
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "wonka-ingest",
						`display_name = "Wonka Ingest"
							sources = ["source1", "source2"]
							app_agent_token = "`+appAgentToken+`"
							`,
					),
					Check: resource.ComposeTestCheckFunc(
						testIngestPipelineResourceDataExists(resourceName, sampleID, appAgentToken),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: "wonka-ingest?location=" + appSpaceID,
				},
			},
		})
	})
})

func testIngestPipelineResourceDataExists(n, expectedID, expectedToken string) resource.TestCheckFunc {
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

			"customer_id":     Equal(customerID),
			"app_space_id":    Equal(appSpaceID),
			"name":            Not(BeEmpty()),
			"app_agent_token": Equal(expectedToken),
			"create_time":     Not(BeEmpty()),
			"update_time":     Not(BeEmpty()),
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), rs.Primary.Attributes)
	}
}
