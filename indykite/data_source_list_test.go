// Copyright (c) 2026 IndyKite
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
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
)

// listDataSourceTestCase describes one config-collection list data source test:
// the mock returns respBody when path and query parameters match, the Terraform
// config lists with (or without) a filter, and entry 0 is verified.
type listDataSourceTestCase struct {
	respBody       any
	entryChecks    map[string]string
	dataSourceType string
	listAttr       string
	scopeAttr      string
	scopeID        string
	scopeParam     string
	path           string
	entryCount     string
	entrySetChecks []string
	fullFetch      bool
	withFilter     bool
}

var _ = Describe("DataSource config collection lists", func() {
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

	configureTestClient := func() {
		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
			client := indykite.NewTestRestClient(mockServer.URL+"/configs/v1", mockServer.Client())
			ctx = indykite.WithClient(ctx, client)
			return cfgFunc(ctx, data)
		}
	}

	createTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	expireTime := time.Date(2027, 6, 1, 12, 0, 0, 0, time.UTC)

	testCases := []listDataSourceTestCase{
		{
			dataSourceType: "indykite_authorization_policies",
			listAttr:       "authorization_policies",
			scopeAttr:      "app_space_id",
			scopeID:        appSpaceID,
			scopeParam:     "project_id",
			path:           "/authorization-policies",
			fullFetch:      true,
			withFilter:     true,
			respBody: indykite.ListResponse[indykite.AuthorizationPolicyResponse]{
				Data: []indykite.AuthorizationPolicyResponse{
					{
						ID: sampleID, CustomerID: customerID, AppSpaceID: appSpaceID,
						Name: "name-one", DisplayName: "Policy One", Description: "First policy",
						Policy: `{"meta":{"policyVersion":"1.0-indykite"}}`, Status: "active",
						Tags: []string{"tag1", "tag2"},
					},
					{ID: applicationID, CustomerID: customerID, AppSpaceID: appSpaceID, Name: "name-two"},
				},
			},
			entryCount: "1",
			entryChecks: map[string]string{
				"id":           sampleID,
				"customer_id":  customerID,
				"app_space_id": appSpaceID,
				"name":         "name-one",
				"display_name": "Policy One",
				"description":  "First policy",
				"json":         `{"meta":{"policyVersion":"1.0-indykite"}}`,
				"status":       "active",
				"tags.#":       "2",
				"tags.0":       "tag1",
			},
		},
		{
			dataSourceType: "indykite_token_introspects",
			listAttr:       "token_introspects",
			scopeAttr:      "app_space_id",
			scopeID:        appSpaceID,
			scopeParam:     "project_id",
			path:           "/token-introspects",
			fullFetch:      true,
			withFilter:     true,
			respBody: indykite.ListResponse[indykite.TokenIntrospectResponse]{
				Data: []indykite.TokenIntrospectResponse{
					{
						ID: sampleID, CustomerID: customerID, AppSpaceID: appSpaceID,
						Name: "name-one", DisplayName: "Introspect One",
						IKGNodeType: "Person", PerformUpsert: true,
					},
					{ID: applicationID, CustomerID: customerID, AppSpaceID: appSpaceID, Name: "name-two"},
				},
			},
			entryCount: "1",
			entryChecks: map[string]string{
				"id":             sampleID,
				"name":           "name-one",
				"ikg_node_type":  "Person",
				"perform_upsert": "true",
			},
		},
		{
			dataSourceType: "indykite_entity_matching_pipelines",
			listAttr:       "entity_matching_pipelines",
			scopeAttr:      "app_space_id",
			scopeID:        appSpaceID,
			scopeParam:     "project_id",
			path:           "/entity-matching-pipelines",
			fullFetch:      true,
			withFilter:     true,
			respBody: indykite.ListResponse[indykite.EntityMatchingPipelineResponse]{
				Data: []indykite.EntityMatchingPipelineResponse{
					{
						ID: sampleID, CustomerID: customerID, AppSpaceID: appSpaceID,
						Name: "name-one", SimilarityScoreCutoff: 0.5, RerunInterval: "1 day",
					},
					{ID: applicationID, CustomerID: customerID, AppSpaceID: appSpaceID, Name: "name-two"},
				},
			},
			entryCount: "1",
			entryChecks: map[string]string{
				"id":                      sampleID,
				"name":                    "name-one",
				"similarity_score_cutoff": "0.5",
				"rerun_interval":          "1 day",
			},
		},
		{
			dataSourceType: "indykite_event_sinks",
			listAttr:       "event_sinks",
			scopeAttr:      "app_space_id",
			scopeID:        appSpaceID,
			scopeParam:     "project_id",
			path:           "/event-sinks",
			withFilter:     true,
			respBody: indykite.ListResponse[indykite.EventSinkResponse]{
				Data: []indykite.EventSinkResponse{
					{
						ID: sampleID, CustomerID: customerID, AppSpaceID: appSpaceID,
						Name: "name-one", DisplayName: "Sink One", Description: "First sink",
					},
					{ID: applicationID, CustomerID: customerID, AppSpaceID: appSpaceID, Name: "name-two"},
				},
			},
			entryCount: "1",
			entryChecks: map[string]string{
				"id":           sampleID,
				"name":         "name-one",
				"display_name": "Sink One",
				"description":  "First sink",
			},
		},
		{
			dataSourceType: "indykite_external_data_resolvers",
			listAttr:       "external_data_resolvers",
			scopeAttr:      "app_space_id",
			scopeID:        appSpaceID,
			scopeParam:     "project_id",
			path:           "/external-data-resolvers",
			fullFetch:      true,
			withFilter:     true,
			respBody: indykite.ListResponse[indykite.ExternalDataResolverResponse]{
				Data: []indykite.ExternalDataResolverResponse{
					{
						ID: sampleID, CustomerID: customerID, AppSpaceID: appSpaceID,
						Name: "name-one", URL: "https://example.com/resolver", Method: "GET",
						RequestType: "json", ResponseType: "json", ResponseSelector: ".",
					},
					{ID: applicationID, CustomerID: customerID, AppSpaceID: appSpaceID, Name: "name-two"},
				},
			},
			entryCount: "1",
			entryChecks: map[string]string{
				"id":                sampleID,
				"name":              "name-one",
				"url":               "https://example.com/resolver",
				"method":            "GET",
				"request_type":      "json",
				"response_type":     "json",
				"response_selector": ".",
			},
		},
		{
			dataSourceType: "indykite_knowledge_queries",
			listAttr:       "knowledge_queries",
			scopeAttr:      "app_space_id",
			scopeID:        appSpaceID,
			scopeParam:     "project_id",
			path:           "/knowledge-queries",
			fullFetch:      true,
			withFilter:     true,
			respBody: indykite.ListResponse[indykite.KnowledgeQueryResponse]{
				Data: []indykite.KnowledgeQueryResponse{
					{
						ID: sampleID, CustomerID: customerID, AppSpaceID: appSpaceID,
						Name: "name-one", Query: `{"nodes":["n.external_id"]}`,
						Status: "active", PolicyID: applicationID,
					},
					{ID: applicationID, CustomerID: customerID, AppSpaceID: appSpaceID, Name: "name-two"},
				},
			},
			entryCount: "1",
			entryChecks: map[string]string{
				"id":        sampleID,
				"name":      "name-one",
				"query":     `{"nodes":["n.external_id"]}`,
				"status":    "active",
				"policy_id": applicationID,
			},
		},
		{
			dataSourceType: "indykite_trust_score_profiles",
			listAttr:       "trust_score_profiles",
			scopeAttr:      "app_space_id",
			scopeID:        appSpaceID,
			scopeParam:     "project_id",
			path:           "/trust-score-profiles",
			fullFetch:      true,
			withFilter:     true,
			respBody: indykite.ListResponse[indykite.TrustScoreProfileResponse]{
				Data: []indykite.TrustScoreProfileResponse{
					{
						ID: sampleID, CustomerID: customerID, AppSpaceID: appSpaceID,
						Name: "name-one", NodeClassification: "Person", Schedule: "UPDATE_FREQUENCY_DAILY",
					},
					{ID: applicationID, CustomerID: customerID, AppSpaceID: appSpaceID, Name: "name-two"},
				},
			},
			entryCount: "1",
			entryChecks: map[string]string{
				"id":                  sampleID,
				"name":                "name-one",
				"node_classification": "Person",
				"schedule":            "UPDATE_FREQUENCY_DAILY",
			},
		},
		{
			dataSourceType: "indykite_mcp_servers",
			listAttr:       "mcp_servers",
			scopeAttr:      "app_space_id",
			scopeID:        appSpaceID,
			scopeParam:     "project_id",
			path:           "/mcp-servers",
			fullFetch:      true,
			withFilter:     true,
			respBody: indykite.ListResponse[indykite.MCPServerResponse]{
				Data: []indykite.MCPServerResponse{
					{
						ID: sampleID, CustomerID: customerID, AppSpaceID: appSpaceID,
						Name: "name-one", AppAgentID: appAgentID, TokenIntrospectID: applicationID,
						ScopesSupported: []string{"read", "write"}, Enabled: true,
					},
					{ID: applicationID, CustomerID: customerID, AppSpaceID: appSpaceID, Name: "name-two"},
				},
			},
			entryCount: "1",
			entryChecks: map[string]string{
				"id":                  sampleID,
				"name":                "name-one",
				"app_agent_id":        appAgentID,
				"token_introspect_id": applicationID,
				"scopes_supported.#":  "2",
				"scopes_supported.0":  "read",
				"enabled":             "true",
			},
		},
		{
			dataSourceType: "indykite_service_accounts",
			listAttr:       "service_accounts",
			scopeAttr:      "customer_id",
			scopeID:        customerID,
			scopeParam:     "organization_id",
			path:           "/service-accounts",
			withFilter:     true,
			respBody: indykite.ListResponse[indykite.ServiceAccountResponse]{
				Data: []indykite.ServiceAccountResponse{
					{
						ID: serviceAccountID, OrganizationID: customerID,
						Name: "name-one", DisplayName: "Account One", Description: "First account",
					},
					{ID: sampleID, OrganizationID: customerID, Name: "name-two"},
				},
			},
			entryCount: "1",
			entryChecks: map[string]string{
				"id":           serviceAccountID,
				"customer_id":  customerID,
				"name":         "name-one",
				"display_name": "Account One",
				"description":  "First account",
			},
		},
		{
			dataSourceType: "indykite_application_agent_credentials",
			listAttr:       "app_agent_credentials",
			scopeAttr:      "app_space_id",
			scopeID:        appSpaceID,
			scopeParam:     "project_id",
			path:           "/application-agent-credentials",
			withFilter:     false,
			respBody: indykite.ListResponse[indykite.ApplicationAgentCredentialResponse]{
				Data: []indykite.ApplicationAgentCredentialResponse{
					{
						ID: appAgentCredID, CustomerID: customerID, AppSpaceID: appSpaceID,
						ApplicationID: applicationID, ApplicationAgentID: appAgentID,
						DisplayName: "Credential One", Kid: "kid-1",
						CreateTime: createTime, ExpireTime: expireTime,
					},
					{
						ID: appAgentCredID2, CustomerID: customerID, AppSpaceID: appSpaceID,
						ApplicationID: applicationID, ApplicationAgentID: appAgentID,
						DisplayName: "Credential Two", Kid: "kid-2", CreateTime: createTime,
					},
				},
			},
			entryCount: "2",
			entryChecks: map[string]string{
				"id":             appAgentCredID,
				"customer_id":    customerID,
				"app_space_id":   appSpaceID,
				"application_id": applicationID,
				"app_agent_id":   appAgentID,
				"display_name":   "Credential One",
				"kid":            "kid-1",
			},
			entrySetChecks: []string{"create_time", "expire_time"},
		},
		{
			dataSourceType: "indykite_service_account_credentials",
			listAttr:       "service_account_credentials",
			scopeAttr:      "customer_id",
			scopeID:        customerID,
			scopeParam:     "organization_id",
			path:           "/service-account-credentials",
			withFilter:     false,
			respBody: indykite.ListResponse[indykite.ServiceAccountCredentialResponse]{
				Data: []indykite.ServiceAccountCredentialResponse{
					{
						ID: serviceAccountCredID, OrganizationID: customerID,
						ServiceAccountID: serviceAccountID, DisplayName: "Credential One",
						Kid: "kid-1", CreateTime: createTime, ExpireTime: "2027-06-01T12:00:00Z",
					},
					{
						ID: sampleID, OrganizationID: customerID,
						ServiceAccountID: serviceAccountID, DisplayName: "Credential Two",
						Kid: "kid-2", CreateTime: createTime,
					},
				},
			},
			entryCount: "2",
			entryChecks: map[string]string{
				"id":                 serviceAccountCredID,
				"customer_id":        customerID,
				"service_account_id": serviceAccountID,
				"display_name":       "Credential One",
				"kid":                "kid-1",
				"expire_time":        "2027-06-01T12:00:00Z",
			},
			entrySetChecks: []string{"create_time"},
		},
	}

	for _, testCase := range testCases {
		It("Test list "+testCase.dataSourceType, func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				wantFullFetch := ""
				if testCase.fullFetch {
					wantFullFetch = "true"
				}
				if r.Method == http.MethodGet &&
					strings.HasSuffix(r.URL.Path, testCase.path) &&
					r.URL.Query().Get(testCase.scopeParam) == testCase.scopeID &&
					r.URL.Query().Get("full_fetch") == wantFullFetch {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(testCase.respBody)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			configureTestClient()

			resName := "data." + testCase.dataSourceType + ".test"
			config := `data "` + testCase.dataSourceType + `" "test" {
				` + testCase.scopeAttr + ` = "` + testCase.scopeID + `"
			`
			expectedID := testCase.scopeID + "/" + testCase.listAttr
			if testCase.withFilter {
				config += `filter = ["name-one", "not-existing-name"]
				`
				expectedID += "/name-one,not-existing-name"
			}
			config += "}"

			checks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr(resName, "id", expectedID),
				resource.TestCheckResourceAttr(resName, testCase.listAttr+".#", testCase.entryCount),
			}
			for attr, expected := range testCase.entryChecks {
				checks = append(checks,
					resource.TestCheckResourceAttr(resName, testCase.listAttr+".0."+attr, expected))
			}
			for _, attr := range testCase.entrySetChecks {
				checks = append(checks,
					resource.TestCheckResourceAttrSet(resName, testCase.listAttr+".0."+attr))
			}

			resource.Test(GinkgoT(), resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"indykite": func() (*schema.Provider, error) { return provider, nil },
				},
				Steps: []resource.TestStep{
					{
						Config: config,
						Check:  resource.ComposeTestCheckFunc(checks...),
					},
				},
			})
		})
	}

	It("Test filter and scope validation", func() {
		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		configureTestClient()

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					Config: `data "indykite_authorization_policies" "test" {
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile(`The argument "app_space_id" is required`),
				},
				{
					Config: `data "indykite_authorization_policies" "test" {
						app_space_id = "` + appSpaceID + `"
					}`,
					ExpectError: regexp.MustCompile(`The argument "filter" is required`),
				},
				{
					Config: `data "indykite_authorization_policies" "test" {
						app_space_id = "` + appSpaceID + `"
						filter = []
					}`,
					ExpectError: regexp.MustCompile("Attribute filter requires 1 item minimum, but config has only 0"),
				},
				{
					Config: `data "indykite_authorization_policies" "test" {
						app_space_id = "` + appSpaceID + `"
						filter = [123]
					}`,
					ExpectError: regexp.MustCompile("Value can have lowercase letters, digits, or hyphens."),
				},
				{
					Config: `data "indykite_service_accounts" "test" {
						customer_id = "abc"
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile("expected to have 'gid:' prefix"),
				},
				{
					Config: `data "indykite_application_agent_credentials" "test" {
						app_space_id = "` + appSpaceID + `"
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile(`An argument named "filter" is not expected here`),
				},
				{
					// Valid configuration, but the mock returns 500 for everything:
					// covers the API error path of the shared list readContext.
					Config: `data "indykite_authorization_policies" "test" {
						app_space_id = "` + appSpaceID + `"
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile("Communication with IndyKite failed"),
				},
			},
		})
	})
})
