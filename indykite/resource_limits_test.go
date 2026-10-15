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
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plan-time limits", func() {
	const (
		resolverConfig = `resource "indykite_external_data_resolver" "resolver" {
			project_id = "` + appSpaceID + `"
			name = "resolver"
			url = "https://example.com/source"
			request_type = "json"
			response_type = "json"
			response_selector = "."
			%s
		}`
		trustScoreConfig = `resource "indykite_trust_score_profile" "score" {
			project_id = "` + appSpaceID + `"
			name = "score"
			node_classification = "Person"
			schedule = "UPDATE_FREQUENCY_DAILY"
			dimension {
				weight = 0.5
				%s
			}
		}`
		eventSinkConfig = `resource "indykite_event_sink" "sink" {
			project_id = "` + appSpaceID + `"
			name = "sink"
			providers {
				provider_name = "kafka-a"
				kafka {
					topic = "topic"
					username = "user"
					password = "password"
					%s
				}
			}
			routes {
				provider_id = "kafka-a"
				keys_values_filter {
					event_type = "indykite.audit.config.*"
				}
			}
		}`
		policyBodyConfig = `resource "indykite_authorization_policy" "wonka" {
			project_id = "` + appSpaceID + `"
			name = "wonka"
			status = "active"
			%s
		}`
		knowledgeQueryConfig = `resource "indykite_knowledge_query" "query" {
			project_id = "` + appSpaceID + `"
			name = "query"
			status = "active"
			policy_id = "` + sampleID + `"
			%s
		}`
		introspectConfig = `resource "indykite_token_introspect" "introspect" {
			project_id = "` + appSpaceID + `"
			name = "introspect"
			ikg_node_type = "MyUser"
			online_validation {
				cache_ttl = 600
			}
			%s
		}`
		policyConfig = `resource "indykite_authorization_policy" "wonka" {
			project_id = "` + appSpaceID + `"
			json = "{}"
			status = "active"
			%s
		}`
		saCredentialConfig = `resource "indykite_service_account_credential" "cred" {
			service_account_id = "` + serviceAccountID + `"
			%s
		}`
		agentCredentialConfig = `resource "indykite_application_agent_credential" "cred" {
			app_agent_id = "` + appAgentID + `"
			%s
		}`
	)

	DescribeTable("are validated at plan time",
		func(configDef, attributes, expectedError string) {
			provider := indykite.Provider()
			// Plan only: any request reaching the API fails the test.
			mockServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()
				Fail(fmt.Sprintf("unexpected API call %s %s", r.Method, r.URL.Path))
			}))
			DeferCleanup(mockServer.Close)
			cfgFunc := provider.ConfigureContextFunc
			provider.ConfigureContextFunc = func(
				ctx context.Context, data *schema.ResourceData,
			) (any, diag.Diagnostics) {
				client := indykite.NewTestRestClient(mockServer.URL+"/configs/v1", mockServer.Client())
				return cfgFunc(indykite.WithClient(ctx, client), data)
			}

			step := resource.TestStep{
				Config:             fmt.Sprintf(configDef, attributes),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			}
			if expectedError != "" {
				step.ExpectError = regexp.MustCompile(expectedError)
			}
			resource.Test(GinkgoT(), resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"indykite": func() (*schema.Provider, error) { return provider, nil },
				},
				Steps: []resource.TestStep{step},
			})
		},
		Entry("valid limits", policyConfig, `name = "`+strings.Repeat("a", 63)+`"
			display_name = "`+strings.Repeat("d", 254)+`"
			description = "`+strings.Repeat("é", 32501)+`"`, ""),
		Entry("name of 64 characters", policyConfig, `name = "`+strings.Repeat("a", 64)+`"`,
			`between 2 and 63 characters but received 64`),
		Entry("display_name of 1 character", policyConfig, `name = "wonka"
			display_name = "x"`, `expected display_name to be empty or between 2 and 254 characters, got 1`),
		Entry("display_name of 255 characters", policyConfig, `name = "wonka"
			display_name = "`+strings.Repeat("d", 255)+`"`,
			`expected display_name to be empty or between 2 and 254 characters, got 255`),
		Entry("description of 1 character", policyConfig, `name = "wonka"
			description = "x"`, `expected description to be empty or between 2 and 65000 characters, got 1`),
		Entry("service account credential display_name", saCredentialConfig, `display_name = "x"`,
			`expected display_name to be empty or between 2 and 254 characters, got 1`),
		Entry("application agent credential display_name", agentCredentialConfig, `display_name = "x"`,
			`expected display_name to be empty or between 2 and 254 characters, got 1`),
		Entry("resolver method in upper case", resolverConfig, `method = "PATCH"`, ""),
		Entry("resolver method in lower case", resolverConfig, `method = "get"`,
			`expected method to be one of \["GET" "POST" "PUT" "PATCH"\], got get`),
		Entry("introspect limits", introspectConfig, `
			jwt_matcher {
				issuer = "https://example.com"
				audience = "`+strings.Repeat("a", 150)+`"
			}
			claims_mapping = { "email" = "`+strings.Repeat("c", 200)+`" }
			sub_claim = "`+strings.Repeat("s", 200)+`"`, ""),
		Entry("introspect audience of 151 characters", introspectConfig, `
			jwt_matcher {
				issuer = "https://example.com"
				audience = "`+strings.Repeat("a", 151)+`"
			}`, `expected jwt_matcher.0.audience to be between 1 and 150 characters, got 151`),
		Entry("introspect claim selector of 201 characters", introspectConfig, `
			jwt_matcher {
				issuer = "https://example.com"
				audience = "aud"
			}
			claims_mapping = { "email" = "`+strings.Repeat("c", 201)+`" }`,
			`expected value of "email" to be between 1 and 200 characters, got 201`),
		Entry("introspect sub_claim of 201 characters", introspectConfig, `
			jwt_matcher {
				issuer = "https://example.com"
				audience = "aud"
			}
			sub_claim = "`+strings.Repeat("s", 201)+`"`,
			`expected sub_claim to be between 1 and 200 characters, got 201`),
		Entry("trust score dimension name", trustScoreConfig, `name = "NAME_FRESHNESS"`, ""),
		Entry("trust score dimension name in API format", trustScoreConfig, `name = "FRESHNESS"`,
			`expected dimension.0.name to be one of`),
		Entry("kafka broker of 8 characters", eventSinkConfig, `brokers = ["a.bc:909"]`, ""),
		Entry("kafka broker of 7 characters", eventSinkConfig, `brokers = ["a.b:909"]`,
			`expected providers.0.kafka.0.brokers.0 to be at least 8 characters, got 7`),
		Entry("no kafka broker", eventSinkConfig, `brokers = []`, `Not enough list items`),
		Entry("policy tags and size", policyBodyConfig, `tags = ["`+strings.Repeat("t", 20)+`", "Tag2"]
			json = "{\"p\":\"`+strings.Repeat("x", 512000-8)+`\"}"`, ""),
		Entry("policy tag of 21 characters", policyConfig, `name = "wonka"
			tags = ["`+strings.Repeat("t", 21)+`"]`, `expected tags.0 to be between 1 and 20 characters, got 21`),
		Entry("policy tag with a hyphen", policyConfig, `name = "wonka"
			tags = ["a-b"]`, `must contain only letters and digits`),
		Entry("repeated policy tags", policyConfig, `name = "wonka"
			tags = ["dup", "dup"]`, `tags must be unique, "dup" is repeated`),
		Entry("policy of 512001 characters", policyBodyConfig, `json = "{\"p\":\"`+strings.Repeat("x", 512001-8)+`\"}"`,
			`expected json to be between 1 and 512000 characters, got 512001`),
		Entry("knowledge query of 512001 characters", knowledgeQueryConfig,
			`query = "{\"p\":\"`+strings.Repeat("x", 512001-8)+`\"}"`,
			`expected query to be between 1 and 512000 characters, got 512001`),
	)
})

var _ = Describe("Provider schema", func() {
	It("passes the SDK internal validation", func() {
		Expect(indykite.Provider().InternalValidate()).To(Succeed())
	})
})
