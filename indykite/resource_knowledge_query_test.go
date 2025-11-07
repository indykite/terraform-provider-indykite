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

var _ = Describe("Resource Knowledge Query config", func() {
	const resourceName = "indykite_knowledge_query.wonka"
	var (
		mockServer            *httptest.Server
		provider              *schema.Provider
		authorizationPolicyID = "gid:AALikeGIDOfAuthZPolicyAA"
	)

	BeforeEach(func() {
		provider = indykite.Provider()
	})

	AfterEach(func() {
		if mockServer != nil {
			mockServer.Close()
		}
	})

	It("Test error cases", func() {
		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						name = "wonka-knowledge-query-config"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "location" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "name" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-knowledge-query-config"

						query = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "status" is required, but no definition was found.`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-knowledge-query-config"

						query = "{}"
						status = "non-existing"
					}
					`,
					ExpectError: regexp.MustCompile(
						`The argument "policy_id" is required, but no definition was found.`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						name = "wonka-knowledge-query-config"

						query = "{}"
						status = "non-existing"
						policy_id = "gid:AALikeGIDOfAuthZPolicyAA"
					}
					`,
					ExpectError: regexp.MustCompile(
						`expected status to be one of \["active" "draft" "inactive"\], got non-existing`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						name = "wonka-knowledge-query-config"

						query = "{}"
						status = "inactive"
						policy_id = "abc"
					}
					`,
					ExpectError: regexp.MustCompile(`expected to have 'gid:' prefix`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						app_space_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-knowledge-query-config"

						query = "{}"
						status = "non-existing"
						policy_id = "gid:AALikeGIDOfAuthZPolicyAA"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "app_space_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						name = "wonka-knowledge-query-config"

						status = "active"
						policy_id = "gid:AALikeGIDOfAuthZPolicyAA"
					}
					`,
					ExpectError: regexp.MustCompile(`argument "query" is required`),
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + customerID + `"
						name = "wonka-knowledge-query-config"

						query = "not valid query"
						status = "active"
						policy_id = "gid:AALikeGIDOfAuthZPolicyAA"
					}
					`,
					ExpectError: regexp.MustCompile(`"query" contains an invalid JSON`),
				},
			},
		})
	})

	It("Test all CRUD", func() {
		createTime := time.Now()
		updateTime := time.Now()

		// Track whether update has been called
		updated := false

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/knowledge-queries"):
				resp := indykite.KnowledgeQueryResponse{
					ID:          sampleID,
					Name:        "wonka-knowledge-query-config",
					DisplayName: "Wonka Query for chocolate receipts",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Query:       `{"something":["like","query"]}`,
					Status:      "active",
					PolicyID:    authorizationPolicyID,
					CreateTime:  createTime,
					UpdateTime:  updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
				var resp indykite.KnowledgeQueryResponse
				if updated {
					// Return updated state
					resp = indykite.KnowledgeQueryResponse{
						ID:          sampleID,
						Name:        "wonka-knowledge-query-config",
						Description: "Description of the best Knowledge Query by Wonka inc.",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Query:       `{"something":["like","another","query"]}`,
						Status:      "draft",
						PolicyID:    authorizationPolicyID,
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				} else {
					// Return initial state (same as POST response)
					resp = indykite.KnowledgeQueryResponse{
						ID:          sampleID,
						Name:        "wonka-knowledge-query-config",
						DisplayName: "Wonka Query for chocolate receipts",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Query:       `{"something":["like","query"]}`,
						Status:      "active",
						PolicyID:    authorizationPolicyID,
						CreateTime:  createTime,
						UpdateTime:  updateTime,
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
				updated = true
				resp := indykite.KnowledgeQueryResponse{
					ID:          sampleID,
					Name:        "wonka-knowledge-query-config",
					Description: "Description of the best Knowledge Query by Wonka inc.",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Query:       `{"something":["like","another","query"]}`,
					Status:      "draft",
					PolicyID:    authorizationPolicyID,
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

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-knowledge-query-config"
						display_name = "Wonka Query for chocolate receipts"

						query = jsonencode({"something":["like", "query"]})
						status = "active"
						policy_id = "` + authorizationPolicyID + `"
					}`,

					Check: resource.ComposeTestCheckFunc(testKnowledgeQueryResourceDataExists(
						resourceName,
						sampleID,
						nil,
					)),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: sampleID,
				},
				{
					Config: `resource "indykite_knowledge_query" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-knowledge-query-config"
						description = "Description of the best Knowledge Query by Wonka inc."

						query = jsonencode({"something":["like", "another", "query"]})
						status = "draft"
						policy_id = "` + authorizationPolicyID + `"
					}
					`,
					Check: resource.ComposeTestCheckFunc(testKnowledgeQueryResourceDataExists(
						resourceName,
						sampleID,
						nil,
					)),
				},
			},
		})
	})

	It("Test import by name with location", func() {
		createTime := time.Now()
		updateTime := time.Now()
		authorizationPolicyID := "gid:AAAABWx1dGhvcml6YXRpb25Qb2xpY3kAAAAA"

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/knowledge-queries"):
				resp := indykite.KnowledgeQueryResponse{
					ID:          sampleID,
					Name:        "wonka-query",
					DisplayName: "Wonka Query",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Query:       `{"something":["like","query"]}`,
					Status:      "active",
					PolicyID:    authorizationPolicyID,
					CreateTime:  createTime,
					UpdateTime:  updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/knowledge-queries/"):
				// Support both ID and name?location=appSpaceID formats
				pathAfterQueries := strings.TrimPrefix(r.URL.Path, "/configs/v1/knowledge-queries/")
				isNameLookup := strings.Contains(pathAfterQueries, "wonka-query")
				isIDLookup := strings.Contains(pathAfterQueries, sampleID)

				if isNameLookup || isIDLookup {
					resp := indykite.KnowledgeQueryResponse{
						ID:          sampleID,
						Name:        "wonka-query",
						DisplayName: "Wonka Query",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Query:       `{"something":["like","query"]}`,
						Status:      "active",
						PolicyID:    authorizationPolicyID,
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
					Config: `resource "indykite_knowledge_query" "wonka" {
							location = "` + appSpaceID + `"
							name = "wonka-query"
							display_name = "Wonka Query"
							query = "{\"something\":[\"like\",\"query\"]}"
							status = "active"
							policy_id = "` + authorizationPolicyID + `"
						}`,
					Check: resource.ComposeTestCheckFunc(
						testKnowledgeQueryResourceDataExists(resourceName, sampleID, nil),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: "wonka-query?location=" + appSpaceID,
				},
			},
		})
	})
})

func testKnowledgeQueryResourceDataExists(
	n string,
	expectedID string,
	extraKeys Keys,
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

			"location":     Not(BeEmpty()),
			"customer_id":  Equal(customerID),
			"app_space_id": Equal(appSpaceID),
			"name":         Not(BeEmpty()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
			"query":        Not(BeEmpty()),
			"status":       Not(BeEmpty()),
			"policy_id":    Not(BeEmpty()),
		}

		for k, v := range extraKeys {
			keys[k] = v
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
	}
}
