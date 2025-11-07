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

var _ = Describe("Resource Application", func() {
	const resourceName = "indykite_application.development"
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
			`resource "indykite_application" "development" {
				app_space_id = "` + appSpaceID + `"
				name = "acme"
				display_name = "%s"
				description = "%s"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()
		currentState := "initial"

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/applications"):
				resp := indykite.ApplicationResponse{
					ID:          applicationID,
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Name:        "acme",
					DisplayName: "acme",
					Description: "Just some App description",
					CreateTime:  createTime,
					UpdateTime:  updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, applicationID):
				var resp indykite.ApplicationResponse
				switch currentState {
				case "initial", "after_create":
					resp = indykite.ApplicationResponse{
						ID:          applicationID,
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Name:        "acme",
						DisplayName: "acme",
						Description: "Just some App description",
						CreateTime:  createTime,
						UpdateTime:  updateTime,
					}
				case "after_update1":
					resp = indykite.ApplicationResponse{
						ID:          applicationID,
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Name:        "acme",
						DisplayName: "acme",
						Description: "Another App description",
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				case "after_update2":
					resp = indykite.ApplicationResponse{
						ID:          applicationID,
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Name:        "acme",
						DisplayName: "Some new display name",
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, applicationID):
				var reqBody map[string]any
				_ = json.NewDecoder(r.Body).Decode(&reqBody)

				if reqBody["description"] != nil && strings.Contains(reqBody["description"].(string), "Another") {
					currentState = "after_update1"
				} else {
					currentState = "after_update2"
				}

				var resp indykite.ApplicationResponse
				if currentState == "after_update1" {
					resp = indykite.ApplicationResponse{
						ID:          applicationID,
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Name:        "acme",
						DisplayName: "acme",
						Description: "Another App description",
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				} else {
					resp = indykite.ApplicationResponse{
						ID:          applicationID,
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Name:        "acme",
						DisplayName: "Some new display name",
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, applicationID):
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
					Config: fmt.Sprintf(tfConfigDef, "", "Just some App description", ""),
					Check: resource.ComposeTestCheckFunc(
						testApplicationResourceDataExists(resourceName, applicationID),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: applicationID,
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "", "Another App description", ""),
					Check: resource.ComposeTestCheckFunc(
						testApplicationResourceDataExists(resourceName, applicationID),
					),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "Some new display name", "", ""),
					Check: resource.ComposeTestCheckFunc(
						testApplicationResourceDataExists(resourceName, applicationID),
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

//nolint:unparam // Test helper function designed to be reusable
func testApplicationResourceDataExists(
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

			"customer_id":  Equal(customerID),
			"app_space_id": Equal(appSpaceID),
			"name":         Not(BeEmpty()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
	}
}

var _ = Describe("Resource Application Import by Name", func() {
	const resourceName = "indykite_application.wonka-app"
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
			`resource "indykite_application" "wonka-app" {
				app_space_id = "` + appSpaceID + `"
				name = "wonka-app"
				display_name = "Wonka Application"
				description = "Just some Application description"
				deletion_protection = false
			}`

		createTime := time.Now()
		updateTime := time.Now()

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/applications"):
				resp := indykite.ApplicationResponse{
					ID:          applicationID,
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Name:        "wonka-app",
					DisplayName: "Wonka Application",
					Description: "Just some Application description",
					CreateTime:  createTime,
					UpdateTime:  updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/applications/"):
				// Support both ID and name?location=appSpaceID formats
				// Check if it's a name-based lookup or ID-based lookup
				pathAfterApplications := strings.TrimPrefix(r.URL.Path, "/configs/v1/applications/")
				isNameLookup := strings.Contains(pathAfterApplications, "wonka-app")
				isIDLookup := strings.Contains(pathAfterApplications, applicationID)

				var resp indykite.ApplicationResponse
				if isNameLookup || isIDLookup {
					resp = indykite.ApplicationResponse{
						ID:          applicationID,
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Name:        "wonka-app",
						DisplayName: "Wonka Application",
						Description: "Just some Application description",
						CreateTime:  createTime,
						UpdateTime:  updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)
				} else {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
				}

			case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, applicationID):
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
						testApplicationResourceDataExists(resourceName, applicationID),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: "wonka-app?location=" + appSpaceID,
				},
			},
		})
	})
})
