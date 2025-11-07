// Copyright (c) 2022 IndyKite
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

var _ = Describe("Resource ApplicationAgent", func() {
	const resourceName = "indykite_application_agent.development"
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

	It("Test all CRUD", func() {
		turnOffDelProtection := "deletion_protection=false"
		tfConfigDef :=
			`resource "indykite_application_agent" "development" {
				application_id = "` + applicationID + `"
				name = "acme"
				display_name = "%s"
				description = "%s"
				api_permissions = ["Authorization","Capture"]
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()
		currentState := "initial"

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/application-agents"):
				resp := indykite.ApplicationAgentResponse{
					ID:             appAgentID,
					CustomerID:     customerID,
					AppSpaceID:     appSpaceID,
					ApplicationID:  applicationID,
					Name:           "acme",
					DisplayName:    "acme",
					Description:    "Just some App description",
					APIPermissions: []string{"Authorization", "Capture"},
					CreateTime:     createTime,
					UpdateTime:     updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, appAgentID):
				var resp indykite.ApplicationAgentResponse
				switch currentState {
				case "initial", "after_create":
					resp = indykite.ApplicationAgentResponse{
						ID:             appAgentID,
						CustomerID:     customerID,
						AppSpaceID:     appSpaceID,
						ApplicationID:  applicationID,
						Name:           "acme",
						DisplayName:    "acme",
						Description:    "Just some App description",
						APIPermissions: []string{"Authorization", "Capture"},
						CreateTime:     createTime,
						UpdateTime:     updateTime,
					}
				case "after_update1":
					resp = indykite.ApplicationAgentResponse{
						ID:             appAgentID,
						CustomerID:     customerID,
						AppSpaceID:     appSpaceID,
						ApplicationID:  applicationID,
						Name:           "acme",
						DisplayName:    "acme",
						Description:    "Another App description",
						APIPermissions: []string{"Authorization", "Capture"},
						CreateTime:     createTime,
						UpdateTime:     time.Now(),
					}
				case "after_update2":
					resp = indykite.ApplicationAgentResponse{
						ID:             appAgentID,
						CustomerID:     customerID,
						AppSpaceID:     appSpaceID,
						ApplicationID:  applicationID,
						Name:           "acme",
						DisplayName:    "Some new display name",
						APIPermissions: []string{"Authorization", "Capture"},
						CreateTime:     createTime,
						UpdateTime:     time.Now(),
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, appAgentID):
				var reqBody map[string]any
				_ = json.NewDecoder(r.Body).Decode(&reqBody)

				if reqBody["description"] != nil && strings.Contains(reqBody["description"].(string), "Another") {
					currentState = "after_update1"
				} else {
					currentState = "after_update2"
				}

				var resp indykite.ApplicationAgentResponse
				if currentState == "after_update1" {
					resp = indykite.ApplicationAgentResponse{
						ID:             appAgentID,
						CustomerID:     customerID,
						AppSpaceID:     appSpaceID,
						ApplicationID:  applicationID,
						Name:           "acme",
						DisplayName:    "acme",
						Description:    "Another App description",
						APIPermissions: []string{"Authorization", "Capture"},
						CreateTime:     createTime,
						UpdateTime:     time.Now(),
					}
				} else {
					resp = indykite.ApplicationAgentResponse{
						ID:             appAgentID,
						CustomerID:     customerID,
						AppSpaceID:     appSpaceID,
						ApplicationID:  applicationID,
						Name:           "acme",
						DisplayName:    "Some new display name",
						APIPermissions: []string{"Authorization", "Capture"},
						CreateTime:     createTime,
						UpdateTime:     time.Now(),
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, appAgentID):
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
				// Errors cases must be always first
				{
					Config:      fmt.Sprintf(tfConfigDef, "", "", `customer_id = "`+customerID+`"`),
					ExpectError: regexp.MustCompile("Value for unconfigurable attribute"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, "", "", `app_space_id = "`+appSpaceID+`"`),
					ExpectError: regexp.MustCompile("Value for unconfigurable attribute"),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "", "Just some App description", ""),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentResourceDataExists(resourceName, appAgentID),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: appAgentID,
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "", "Another App description", ""),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentResourceDataExists(resourceName, appAgentID),
					),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "Some new display name", "", ""),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentResourceDataExists(resourceName, appAgentID),
					),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, "Some new display name", "", ""),
					Destroy:     true,
					ExpectError: regexp.MustCompile("Cannot destroy instance"),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "Some new display name", "", turnOffDelProtection),
				},
			},
		})
	})

	It("Test import by name with location", func() {
		tfConfigDef :=
			`resource "indykite_application_agent" "development" {
				application_id = "` + applicationID + `"
				name = "wonka-agent"
				display_name = "Wonka Agent"
				description = "Just some Agent description"
				api_permissions = ["Authorization","Capture"]
				deletion_protection = false
			}`

		createTime := time.Now()
		updateTime := time.Now()

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/application-agents"):
				resp := indykite.ApplicationAgentResponse{
					ID:             appAgentID,
					CustomerID:     customerID,
					AppSpaceID:     appSpaceID,
					ApplicationID:  applicationID,
					Name:           "wonka-agent",
					DisplayName:    "Wonka Agent",
					Description:    "Just some Agent description",
					APIPermissions: []string{"Authorization", "Capture"},
					CreateTime:     createTime,
					UpdateTime:     updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/application-agents/"):
				// Support both ID and name?location=applicationID formats
				// Check if it's a name-based lookup or ID-based lookup
				pathAfterAgents := strings.TrimPrefix(r.URL.Path, "/configs/v1/application-agents/")
				isNameLookup := strings.Contains(pathAfterAgents, "wonka-agent")
				isIDLookup := strings.Contains(pathAfterAgents, appAgentID)

				var resp indykite.ApplicationAgentResponse
				if isNameLookup || isIDLookup {
					resp = indykite.ApplicationAgentResponse{
						ID:             appAgentID,
						CustomerID:     customerID,
						AppSpaceID:     appSpaceID,
						ApplicationID:  applicationID,
						Name:           "wonka-agent",
						DisplayName:    "Wonka Agent",
						Description:    "Just some Agent description",
						APIPermissions: []string{"Authorization", "Capture"},
						CreateTime:     createTime,
						UpdateTime:     updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)
				} else {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
				}

			case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, appAgentID):
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
					Config: tfConfigDef,
					Check: resource.ComposeTestCheckFunc(
						testAppAgentResourceDataExists(resourceName, appAgentID),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: "wonka-agent?location=" + applicationID,
				},
			},
		})
	})
})

//nolint:unparam // Test helper function designed to be reusable
func testAppAgentResourceDataExists(n, expectedID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID != expectedID {
			return errors.New("ID does not match")
		}

		keys := Keys{
			"id":                Equal(expectedID),
			"%":                 Not(BeEmpty()),
			"api_permissions.#": Equal("2"),
			"api_permissions.0": Equal("Authorization"),
			"api_permissions.1": Equal("Capture"),

			"customer_id":         Equal(customerID),
			"app_space_id":        Equal(appSpaceID),
			"application_id":      Equal(applicationID),
			"name":                Not(BeEmpty()),
			"create_time":         Not(BeEmpty()),
			"update_time":         Not(BeEmpty()),
			"deletion_protection": Not(BeEmpty()),
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), rs.Primary.Attributes)
	}
}
