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
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource project_id", func() {
	// The Config API takes project_id in the create body of these resources and
	// deprecates the location query parameter, so location is a deprecated alias.
	DescribeTable("replaces the deprecated location",
		func(resourceType string) {
			res := indykite.Provider().ResourcesMap[resourceType]
			Expect(res).NotTo(BeNil())

			projectID := res.Schema["project_id"]
			Expect(projectID).NotTo(BeNil())
			Expect(projectID.Optional).To(BeTrue())
			Expect(projectID.Computed).To(BeTrue())
			Expect(projectID.ExactlyOneOf).To(ConsistOf("location", "project_id"))

			location := res.Schema["location"]
			Expect(location).NotTo(BeNil())
			Expect(location.Required).To(BeFalse())
			Expect(location.Deprecated).NotTo(BeEmpty())
			Expect(location.ExactlyOneOf).To(ConsistOf("location", "project_id"))
		},
		Entry(nil, "indykite_audit_signing"),
		Entry(nil, "indykite_authorization_policy"),
		Entry(nil, "indykite_entity_matching_pipeline"),
		Entry(nil, "indykite_event_sink"),
		Entry(nil, "indykite_external_data_resolver"),
		Entry(nil, "indykite_knowledge_query"),
		Entry(nil, "indykite_mcp_server"),
		Entry(nil, "indykite_token_introspect"),
		Entry(nil, "indykite_trust_score_profile"),
	)

	DescribeTable("imports by name with project_id",
		func(resourceType string) {
			res := indykite.Provider().ResourcesMap[resourceType]
			Expect(res).NotTo(BeNil())
			importID := func(id string) (string, error) {
				data := res.TestResourceData()
				data.SetId(id)
				imported, err := res.Importer.StateContext(context.Background(), data, nil)
				if err != nil {
					return "", err
				}
				return imported[0].Id(), nil
			}

			Expect(importID("wonka?project_id=" + appSpaceID)).To(Equal("wonka?location=" + appSpaceID))
			Expect(importID("wonka?location=" + appSpaceID)).To(Equal("wonka?location=" + appSpaceID))
			Expect(importID(sampleID)).To(Equal(sampleID))

			_, err := importID("wonka?organization_id=" + customerID)
			Expect(err).To(MatchError(
				"Unimplemented id format: wonka?organization_id=" + customerID +
					". Expected 'gid:xxx', 'resource-name?project_id=gid:xxx', " +
					"or deprecated 'resource-name?location=gid:xxx'"))
		},
		Entry(nil, "indykite_audit_signing"),
		Entry(nil, "indykite_authorization_policy"),
		Entry(nil, "indykite_entity_matching_pipeline"),
		Entry(nil, "indykite_event_sink"),
		Entry(nil, "indykite_external_data_resolver"),
		Entry(nil, "indykite_knowledge_query"),
		Entry(nil, "indykite_mcp_server"),
		Entry(nil, "indykite_token_introspect"),
		Entry(nil, "indykite_trust_score_profile"),
	)

	It("Test CRUD of Audit Signing configuration with project_id", func() {
		var (
			mu            sync.Mutex
			stored        indykite.AuditSigningResponse
			nameLookupPID []string
		)
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/audit-signings"):
				var req indykite.CreateAuditSigningRequest
				_ = json.NewDecoder(r.Body).Decode(&req)
				if req.ProjectID != appSpaceID {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				stored = indykite.AuditSigningResponse{
					ID:          sampleID,
					Name:        req.Name,
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Provider:    req.Provider,
					KeyResource: new(""),
					Kid:         new(""),
					CreateTime:  time.Now(),
					UpdateTime:  time.Now(),
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]string{"id": sampleID})
			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(stored)
			case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/audit-signings/wonka-audit"):
				// Import by name: the Config API resolves the name within the project_id.
				nameLookupPID = append(nameLookupPID, r.URL.Query().Get("project_id"))
				if r.URL.Query().Get("project_id") != appSpaceID || r.URL.Query().Has("location") {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(stored)
			case r.Method == http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer mockServer.Close()

		provider := indykite.Provider()
		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
			client := indykite.NewTestRestClient(mockServer.URL+"/configs/v1", mockServer.Client())
			ctx = indykite.WithClient(ctx, client)
			return cfgFunc(ctx, data)
		}

		tfConfigDef := `resource "indykite_audit_signing" "development" {
				%s
				name         = "wonka-audit"
				key_provider = "PLATFORM_MANAGED"
			}`

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				// Errors case must always come first
				{
					Config:      fmt.Sprintf(tfConfigDef, ""),
					ExpectError: regexp.MustCompile("one of `location,project_id` must be specified"),
				},
				{
					Config: fmt.Sprintf(tfConfigDef,
						`project_id = "`+appSpaceID+`"
						location   = "`+appSpaceID+`"`),
					ExpectError: regexp.MustCompile("only one of `location,project_id` can be specified"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, `project_id = "ccc"`),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, `project_id = "`+appSpaceID+`"`),
					Check: resource.ComposeTestCheckFunc(
						testAuditSigningResourceDataExists("PLATFORM_MANAGED", "", "", nil),
						resource.TestCheckResourceAttr(auditSigningResourceName, "project_id", appSpaceID),
					),
				},
				{
					// Import by name within the project, the state must match the created resource.
					ResourceName:      auditSigningResourceName,
					ImportState:       true,
					ImportStateId:     "wonka-audit?project_id=" + appSpaceID,
					ImportStateVerify: true,
				},
				{
					// Moving back to the deprecated location with the same value is not a change.
					Config:   fmt.Sprintf(tfConfigDef, `location = "`+appSpaceID+`"`),
					PlanOnly: true,
				},
			},
		})

		mu.Lock()
		defer mu.Unlock()
		Expect(nameLookupPID).To(ConsistOf(appSpaceID))
	})
})
