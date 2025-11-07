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
			Providers: map[string]*schema.Provider{
				"indykite": provider,
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
			Providers: map[string]*schema.Provider{
				"indykite": provider,
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
})
