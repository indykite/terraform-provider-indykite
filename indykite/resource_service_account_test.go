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

var _ = Describe("Resource Service Account", func() {
	const resourceName = "indykite_service_account.development"
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
			`resource "indykite_service_account" "development" {
				customer_id = "` + customerID + `"
				name = "my-service-account"
				display_name = "%s"
				description = "%s"
				role = "all_editor"
				%s
			}`

		serviceAccountID := "gid:AAAABWx1dGhvcml6YXRpb25Qb2xpY3kAAAAA"
		createTime := time.Now()
		updateTime := time.Now()
		currentState := "initial"

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/service-accounts"):
				resp := indykite.ServiceAccountResponse{
					ID:             serviceAccountID,
					OrganizationID: customerID,
					Name:           "my-service-account",
					DisplayName:    "My Service Account",
					Description:    "Just some service account description",
					Role:           "all_editor",
					CreateTime:     createTime,
					UpdateTime:     updateTime,
				}
				currentState = "after_create"
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, serviceAccountID):
				var resp indykite.ServiceAccountResponse
				switch currentState {
				case "initial", "after_create":
					resp = indykite.ServiceAccountResponse{
						ID:             serviceAccountID,
						OrganizationID: customerID,
						Name:           "my-service-account",
						DisplayName:    "My Service Account",
						Description:    "Just some service account description",
						Role:           "all_editor",
						CreateTime:     createTime,
						UpdateTime:     updateTime,
					}
				case "after_update1":
					resp = indykite.ServiceAccountResponse{
						ID:             serviceAccountID,
						OrganizationID: customerID,
						Name:           "my-service-account",
						DisplayName:    "My Service Account",
						Description:    "Another service account description",
						Role:           "all_editor",
						CreateTime:     createTime,
						UpdateTime:     time.Now(),
					}
				case "after_update2":
					resp = indykite.ServiceAccountResponse{
						ID:             serviceAccountID,
						OrganizationID: customerID,
						Name:           "my-service-account",
						DisplayName:    "Some new display name",
						Role:           "all_editor",
						CreateTime:     createTime,
						UpdateTime:     time.Now(),
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, serviceAccountID):
				var reqBody map[string]any
				_ = json.NewDecoder(r.Body).Decode(&reqBody)

				if reqBody["description"] != nil && strings.Contains(reqBody["description"].(string), "Another") {
					currentState = "after_update1"
				} else {
					currentState = "after_update2"
				}

				var resp indykite.ServiceAccountResponse
				if currentState == "after_update1" {
					resp = indykite.ServiceAccountResponse{
						ID:             serviceAccountID,
						OrganizationID: customerID,
						Name:           "my-service-account",
						DisplayName:    "My Service Account",
						Description:    "Another service account description",
						Role:           "all_editor",
						CreateTime:     createTime,
						UpdateTime:     time.Now(),
					}
				} else {
					resp = indykite.ServiceAccountResponse{
						ID:             serviceAccountID,
						OrganizationID: customerID,
						Name:           "my-service-account",
						DisplayName:    "Some new display name",
						Role:           "all_editor",
						CreateTime:     createTime,
						UpdateTime:     time.Now(),
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, serviceAccountID):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{})

			default:
				w.WriteHeader(http.StatusNotImplemented)
			}
		}))

		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
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
					Config: fmt.Sprintf(tfConfigDef, "My Service Account", "Just some service account description", ""),
					Check: resource.ComposeTestCheckFunc(
						testServiceAccountResourceDataExists(resourceName),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: serviceAccountID,
				},
				{
					Config: fmt.Sprintf(
						tfConfigDef,
						"My Service Account",
						"Another service account description",
						"",
					),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							resourceName,
							"description",
							"Another service account description",
						),
					),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "Some new display name", "", ""),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "display_name", "Some new display name"),
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
})

func testServiceAccountResourceDataExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("ID is not set")
		}
		attrs := rs.Primary.Attributes

		keys := Keys{
			"id": Not(BeEmpty()),
			"%":  Not(BeEmpty()),

			"customer_id": Equal(customerID),
			"name":        Not(BeEmpty()),
			"role":        Not(BeEmpty()),
			"create_time": Not(BeEmpty()),
			"update_time": Not(BeEmpty()),
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
	}
}

var _ = Describe("Resource ServiceAccount Import by Name", func() {
	const resourceName = "indykite_service_account.development"
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

	It("Test import by name with location", func() {
		tfConfigDef :=
			`resource "indykite_service_account" "development" {
				customer_id = "` + customerID + `"
				name = "my-service-account"
				display_name = "My Service Account"
				description = "Just some service account description"
				role = "all_editor"
				deletion_protection = false
			}`

		serviceAccountID := "gid:AAAABWx1dGhvcml6YXRpb25Qb2xpY3kAAAAA"
		createTime := time.Now()
		updateTime := time.Now()

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/service-accounts"):
				resp := indykite.ServiceAccountResponse{
					ID:             serviceAccountID,
					OrganizationID: customerID,
					Name:           "my-service-account",
					DisplayName:    "My Service Account",
					Description:    "Just some service account description",
					Role:           "all_editor",
					CreateTime:     createTime,
					UpdateTime:     updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/service-accounts/"):
				// Support both ID and name?location=organizationID formats
				// Check if it's a name-based lookup or ID-based lookup
				pathAfterServiceAccounts := strings.TrimPrefix(r.URL.Path, "/configs/v1/service-accounts/")
				isNameLookup := strings.Contains(pathAfterServiceAccounts, "my-service-account")
				isIDLookup := strings.Contains(pathAfterServiceAccounts, serviceAccountID)

				var resp indykite.ServiceAccountResponse
				if isNameLookup || isIDLookup {
					resp = indykite.ServiceAccountResponse{
						ID:             serviceAccountID,
						OrganizationID: customerID,
						Name:           "my-service-account",
						DisplayName:    "My Service Account",
						Description:    "Just some service account description",
						Role:           "all_editor",
						CreateTime:     createTime,
						UpdateTime:     updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)
				} else {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
				}

			case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, serviceAccountID):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{})

			default:
				w.WriteHeader(http.StatusNotImplemented)
			}
		}))

		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
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
					Config: tfConfigDef,
					Check: resource.ComposeTestCheckFunc(
						testServiceAccountResourceDataExists(resourceName),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: "my-service-account?location=" + customerID,
				},
			},
		})
	})

	It("Test update clears fields and skips deletion_protection only changes", func() {
		tfConfigDef :=
			`resource "indykite_service_account" "development" {
				customer_id = "` + customerID + `"
				name = "update-service-account"
				role = "all_viewer"
				%s
			}`

		// Like the Config API: an update without changes is rejected, an empty string clears
		// a field and an empty display_name falls back to the name.
		var (
			account    map[string]any
			putBodies  []map[string]any
			putRejects int
		)
		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/service-accounts"):
				var req map[string]any
				Expect(json.NewDecoder(r.Body).Decode(&req)).To(Succeed())
				account = map[string]any{
					"id":              serviceAccountID,
					"name":            req["name"],
					"display_name":    req["name"],
					"description":     "",
					"organization_id": req["organization_id"],
					"create_time":     "2026-10-14T08:00:00Z",
					"update_time":     "2026-10-14T08:00:00Z",
				}
				for _, key := range []string{"display_name", "description"} {
					if v, ok := req[key].(string); ok && v != "" {
						account[key] = v
					}
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": serviceAccountID})

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, serviceAccountID):
				var req map[string]any
				Expect(json.NewDecoder(r.Body).Decode(&req)).To(Succeed())
				putBodies = append(putBodies, req)
				if len(req) == 0 {
					putRejects++
					w.WriteHeader(http.StatusUnprocessableEntity)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"message": "Unprocessable Entity", "details": []string{"nothing to change"}})
					return
				}
				if v, ok := req["description"].(string); ok {
					account["description"] = v
				}
				if v, ok := req["display_name"].(string); ok {
					if v == "" {
						v, _ = account["name"].(string)
					}
					account["display_name"] = v
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": serviceAccountID})

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, serviceAccountID) && account != nil:
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(account)

			case r.Method == http.MethodDelete:
				w.WriteHeader(http.StatusOK)

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

		expectPuts := func(count int, lastBody OmegaMatcher) resource.TestCheckFunc {
			return func(_ *terraform.State) error {
				if err := convertOmegaMatcherToError(HaveLen(count), putBodies); err != nil {
					return fmt.Errorf("unexpected number of updates: %w", err)
				}
				if putRejects != 0 {
					return fmt.Errorf("%d update(s) rejected with nothing to change", putRejects)
				}
				if lastBody == nil {
					return nil
				}
				return convertOmegaMatcherToError(lastBody, putBodies[count-1])
			}
		}

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(tfConfigDef, `
						display_name = "Update Account"
						description = "Some description"
						deletion_protection = false`),
					Check: resource.ComposeTestCheckFunc(
						expectPuts(0, nil),
						resource.TestCheckResourceAttr(resourceName, "description", "Some description"),
					),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, `deletion_protection = false`),
					Check: resource.ComposeTestCheckFunc(
						expectPuts(1, Equal(map[string]any{"display_name": "", "description": ""})),
						resource.TestCheckResourceAttr(resourceName, "description", ""),
					),
				},
				{
					// Only deletion_protection changes: nothing is sent to the API.
					Config: fmt.Sprintf(tfConfigDef, `deletion_protection = true`),
					Check: resource.ComposeTestCheckFunc(
						expectPuts(1, nil),
						resource.TestCheckResourceAttr(resourceName, "deletion_protection", "true"),
					),
				},
				{
					// Turn protection off again so the test can destroy the service account.
					Config: fmt.Sprintf(tfConfigDef, `deletion_protection = false`),
					Check:  expectPuts(1, nil),
				},
			},
		})
	})

	It("Test rename fails at apply for a service account", func() {
		tfConfigDef := `resource "indykite_service_account" "development" {
			customer_id = "` + customerID + `"
			name = "%s"
			role = "all_viewer"
			deletion_protection = %t
		}`

		fake := newFakeConfigAPI("/service-accounts", serviceAccountID, nil)
		mockServer = fake.Server

		cfgFunc := provider.ConfigureContextFunc
		provider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (any, diag.Diagnostics) {
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
					Config: fmt.Sprintf(tfConfigDef, "rename-account", false),
				},
				{
					// A rename, here together with a deletion_protection change, fails before any request.
					Config:      fmt.Sprintf(tfConfigDef, "rename-account-v2", true),
					ExpectError: regexp.MustCompile(`name cannot be changed after creation`),
				},
				{
					Config:   fmt.Sprintf(tfConfigDef, "rename-account", false),
					PlanOnly: true,
					Check: func(_ *terraform.State) error {
						return convertOmegaMatcherToError(BeEmpty(), fake.putBodies)
					},
				},
			},
		})
	})

	It("Test role is kept when the API does not return it", func() {
		tfConfigDef :=
			`resource "indykite_service_account" "development" {
				customer_id = "` + customerID + `"
				name = "role-service-account"
				role = "%s"
				deletion_protection = false
			}`

		// Like the Config API, the fake never returns the role and has no role on update.
		var (
			created    map[string]any
			posts      []string // role of each create
			putCount   int
			expectPost = func(roles ...string) resource.TestCheckFunc {
				return func(_ *terraform.State) error {
					if err := convertOmegaMatcherToError(Equal(roles), posts); err != nil {
						return fmt.Errorf("unexpected creates: %w", err)
					}
					return convertOmegaMatcherToError(BeZero(), putCount)
				}
			}
		)
		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/service-accounts"):
				var req map[string]any
				Expect(json.NewDecoder(r.Body).Decode(&req)).To(Succeed())
				role, _ := req["role"].(string)
				posts = append(posts, role)
				created = map[string]any{
					"id":              serviceAccountID,
					"name":            req["name"],
					"display_name":    req["name"],
					"description":     "",
					"organization_id": req["organization_id"],
					"create_time":     "2026-10-14T06:49:00Z",
					"update_time":     "2026-10-14T06:49:00Z",
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": serviceAccountID})

			case r.Method == http.MethodPut:
				putCount++
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": serviceAccountID})

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, serviceAccountID) && created != nil:
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(created)

			case r.Method == http.MethodDelete:
				w.WriteHeader(http.StatusOK)

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
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					// Create; the implicit plan after apply fails if the missing role forces replacement.
					Config: fmt.Sprintf(tfConfigDef, "all_viewer"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "role", "all_viewer"),
						expectPost("all_viewer"),
					),
				},
				{
					// A role change replaces the service account.
					Config: fmt.Sprintf(tfConfigDef, "all_editor"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "role", "all_editor"),
						expectPost("all_viewer", "all_editor"),
					),
				},
			},
		})

		// Import into an empty state; the fake still serves the service account created above.
		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					// Import cannot recover the role.
					Config:             fmt.Sprintf(tfConfigDef, "all_editor"),
					ResourceName:       resourceName,
					ImportState:        true,
					ImportStateId:      serviceAccountID,
					ImportStatePersist: true,
					ImportStateCheck: func(states []*terraform.InstanceState) error {
						if len(states) != 1 {
							return fmt.Errorf("expected 1 imported state, got %d", len(states))
						}
						return convertOmegaMatcherToError(
							Not(HaveKeyWithValue("role", Not(BeEmpty()))), states[0].Attributes)
					},
				},
				{
					// The first apply after import records the configured role in place, without any request.
					Config: fmt.Sprintf(tfConfigDef, "all_editor"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "role", "all_editor"),
						expectPost("all_viewer", "all_editor"),
					),
				},
				{
					// Once recorded, a role change replaces the service account.
					Config: fmt.Sprintf(tfConfigDef, "all_viewer"),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "role", "all_viewer"),
						expectPost("all_viewer", "all_editor", "all_viewer"),
					),
				},
			},
		})
	})
})
