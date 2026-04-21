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

const (
	mcpServerTokenIntrospectID = "gid:AAAABnRva2VuSW50cm9zcGVjdAAA" // #nosec G101
)

var _ = Describe("Resource MCPServer", func() {
	const (
		resourceName = "indykite_mcp_server.development"
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

	It("Test CRUD of MCP Server configuration", func() {
		tfConfigDef :=
			`resource "indykite_mcp_server" "development" {
				location = "%s"
				name = "%s"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()
		updated := false

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/mcp-servers"):
				var req indykite.CreateMCPServerRequest
				_ = json.NewDecoder(r.Body).Decode(&req)

				// Basic sanity check on request shape.
				if req.Name == "" || req.AppAgentID == "" || req.TokenIntrospectID == "" ||
					len(req.ScopesSupported) == 0 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				resp := indykite.MCPServerResponse{
					ID:                sampleID,
					Name:              "my-first-mcp-server",
					DisplayName:       "Display name of MCP Server",
					CustomerID:        customerID,
					AppSpaceID:        appSpaceID,
					AppAgentID:        appAgentID,
					TokenIntrospectID: mcpServerTokenIntrospectID,
					ScopesSupported:   []string{"read", "write"},
					Enabled:           true,
					CreateTime:        createTime,
					UpdateTime:        updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
				var resp indykite.MCPServerResponse
				if updated {
					resp = indykite.MCPServerResponse{
						ID:                sampleID,
						Name:              "my-first-mcp-server",
						Description:       "mcp server description",
						CustomerID:        customerID,
						AppSpaceID:        appSpaceID,
						AppAgentID:        appAgentID,
						TokenIntrospectID: mcpServerTokenIntrospectID,
						ScopesSupported:   []string{"read", "write", "admin"},
						Enabled:           false,
						CreateTime:        createTime,
						UpdateTime:        time.Now(),
					}
				} else {
					resp = indykite.MCPServerResponse{
						ID:                sampleID,
						Name:              "my-first-mcp-server",
						DisplayName:       "Display name of MCP Server",
						CustomerID:        customerID,
						AppSpaceID:        appSpaceID,
						AppAgentID:        appAgentID,
						TokenIntrospectID: mcpServerTokenIntrospectID,
						ScopesSupported:   []string{"read", "write"},
						Enabled:           true,
						CreateTime:        createTime,
						UpdateTime:        updateTime,
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
				updated = true
				resp := indykite.MCPServerResponse{
					ID:                sampleID,
					Name:              "my-first-mcp-server",
					Description:       "mcp server description",
					CustomerID:        customerID,
					AppSpaceID:        appSpaceID,
					AppAgentID:        appAgentID,
					TokenIntrospectID: mcpServerTokenIntrospectID,
					ScopesSupported:   []string{"read", "write", "admin"},
					Enabled:           false,
					CreateTime:        createTime,
					UpdateTime:        time.Now(),
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

		validSettings := fmt.Sprintf(`
		app_agent_id        = "%s"
		token_introspect_id = "%s"
		scopes_supported    = ["read", "write"]
		enabled             = true
		`, appAgentID, mcpServerTokenIntrospectID)

		validSettingsUpdate := fmt.Sprintf(`
		app_agent_id        = "%s"
		token_introspect_id = "%s"
		scopes_supported    = ["read", "write", "admin"]
		enabled             = false
		`, appAgentID, mcpServerTokenIntrospectID)

		// Invalid app_agent_id (not a GID)
		invalidAppAgent := fmt.Sprintf(`
		app_agent_id        = "not-a-gid"
		token_introspect_id = "%s"
		scopes_supported    = ["read"]
		enabled             = true
		`, mcpServerTokenIntrospectID)

		// Missing scopes_supported
		missingScopes := fmt.Sprintf(`
		app_agent_id        = "%s"
		token_introspect_id = "%s"
		enabled             = true
		`, appAgentID, mcpServerTokenIntrospectID)

		// Empty scopes_supported
		emptyScopes := fmt.Sprintf(`
		app_agent_id        = "%s"
		token_introspect_id = "%s"
		scopes_supported    = []
		enabled             = true
		`, appAgentID, mcpServerTokenIntrospectID)

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors case must always come first
				{
					Config:      fmt.Sprintf(tfConfigDef, "ccc", "name", validSettings),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", invalidAppAgent),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", missingScopes),
					ExpectError: regexp.MustCompile("Missing required argument"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", emptyScopes),
					ExpectError: regexp.MustCompile("Attribute scopes_supported requires 1 item minimum"),
				},

				{
					// Create and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-mcp-server",
						`display_name = "Display name of MCP Server"
						`+validSettings+``,
					),
					Check: resource.ComposeTestCheckFunc(
						testMCPServerResourceDataExists(resourceName, sampleID, true, []string{"read", "write"}),
					),
				},
				{
					// Import by ID
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: sampleID,
				},
				{
					// Update and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-mcp-server",
						`description = "mcp server description"
						`+validSettingsUpdate+``,
					),
					Check: resource.ComposeTestCheckFunc(
						testMCPServerResourceDataExists(
							resourceName, sampleID, false, []string{"read", "write", "admin"}),
					),
				},
			},
		})
	})

	It("Test import by name with location", func() {
		tfConfigDef := `resource "indykite_mcp_server" "development" {
				location = "%s"
				name = "%s"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/mcp-servers"):
				resp := indykite.MCPServerResponse{
					ID:                sampleID,
					Name:              "wonka-mcp",
					DisplayName:       "Wonka MCP",
					CustomerID:        customerID,
					AppSpaceID:        appSpaceID,
					AppAgentID:        appAgentID,
					TokenIntrospectID: mcpServerTokenIntrospectID,
					ScopesSupported:   []string{"read"},
					Enabled:           true,
					CreateTime:        createTime,
					UpdateTime:        updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/mcp-servers/"):
				pathAfter := strings.TrimPrefix(r.URL.Path, "/configs/v1/mcp-servers/")
				isNameLookup := strings.Contains(pathAfter, "wonka-mcp")
				isIDLookup := strings.Contains(pathAfter, sampleID)

				if isNameLookup || isIDLookup {
					resp := indykite.MCPServerResponse{
						ID:                sampleID,
						Name:              "wonka-mcp",
						DisplayName:       "Wonka MCP",
						CustomerID:        customerID,
						AppSpaceID:        appSpaceID,
						AppAgentID:        appAgentID,
						TokenIntrospectID: mcpServerTokenIntrospectID,
						ScopesSupported:   []string{"read"},
						Enabled:           true,
						CreateTime:        createTime,
						UpdateTime:        updateTime,
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
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "wonka-mcp",
						fmt.Sprintf(`display_name        = "Wonka MCP"
							app_agent_id        = "%s"
							token_introspect_id = "%s"
							scopes_supported    = ["read"]
							enabled             = true
							`, appAgentID, mcpServerTokenIntrospectID),
					),
					Check: resource.ComposeTestCheckFunc(
						testMCPServerResourceDataExists(resourceName, sampleID, true, []string{"read"}),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: "wonka-mcp?location=" + appSpaceID,
				},
			},
		})
	})
})

func testMCPServerResourceDataExists(
	n, expectedID string, expectedEnabled bool, expectedScopes []string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID != expectedID {
			return errors.New("ID does not match")
		}

		enabledStr := "false"
		if expectedEnabled {
			enabledStr = "true"
		}

		keys := Keys{
			"id": Equal(expectedID),
			"%":  Not(BeEmpty()),

			"customer_id":         Equal(customerID),
			"app_space_id":        Equal(appSpaceID),
			"name":                Not(BeEmpty()),
			"app_agent_id":        Equal(appAgentID),
			"token_introspect_id": Equal(mcpServerTokenIntrospectID),
			"enabled":             Equal(enabledStr),
			"scopes_supported.#":  Equal(strconv.Itoa(len(expectedScopes))),
			"create_time":         Not(BeEmpty()),
			"update_time":         Not(BeEmpty()),
		}
		for i, scope := range expectedScopes {
			keys[fmt.Sprintf("scopes_supported.%d", i)] = Equal(scope)
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), rs.Primary.Attributes)
	}
}
