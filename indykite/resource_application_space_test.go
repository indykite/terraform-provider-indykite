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

var _ = Describe("Resource Application Space", func() {
	const resourceName = "indykite_application_space.development"
	const resourceNameSimple = "indykite_application_space.developmentSimple"
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

	It("Test all CRUD Simple", func() {
		turnOffDelProtection := "deletion_protection=false"
		tfConfigDefSimple :=
			`resource "indykite_application_space" "developmentSimple" {
				customer_id = "` + customerID + `"
				name = "acme0"
				display_name = "%s"
				description = "%s"
				region = "europe-west1"
				ikg_size = "4GB"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()
		currentState := "initial"

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects"):
				resp := indykite.ApplicationSpaceResponse{
					ID:          appSpaceID,
					CustomerID:  customerID,
					Name:        "acme0",
					DisplayName: "acme0",
					Description: "Just some AppSpace description",
					Region:      "europe-west1",
					IKGSize:     "4GB",
					IKGStatus:   "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
					CreateTime:  createTime,
					UpdateTime:  updateTime,
				}
				currentState = "after_create"
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, appSpaceID):
				var resp indykite.ApplicationSpaceResponse
				switch currentState {
				case "initial", "after_create":
					resp = indykite.ApplicationSpaceResponse{
						ID:          appSpaceID,
						CustomerID:  customerID,
						Name:        "acme0",
						DisplayName: "acme0",
						Description: "Just some AppSpace description",
						Region:      "europe-west1",
						IKGSize:     "4GB",
						IKGStatus:   "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
						CreateTime:  createTime,
						UpdateTime:  updateTime,
					}
				case "after_update1":
					resp = indykite.ApplicationSpaceResponse{
						ID:          appSpaceID,
						CustomerID:  customerID,
						Name:        "acme0",
						DisplayName: "acme0",
						Description: "Another AppSpace description",
						Region:      "europe-west1",
						IKGSize:     "4GB",
						IKGStatus:   "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				case "after_update2":
					resp = indykite.ApplicationSpaceResponse{
						ID:          appSpaceID,
						CustomerID:  customerID,
						Name:        "acme0",
						DisplayName: "Some new display name",
						Region:      "europe-west1",
						IKGSize:     "4GB",
						IKGStatus:   "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, appSpaceID):
				var reqBody map[string]any
				_ = json.NewDecoder(r.Body).Decode(&reqBody)

				if reqBody["description"] != nil && strings.Contains(reqBody["description"].(string), "Another") {
					currentState = "after_update1"
				} else {
					currentState = "after_update2"
				}

				var resp indykite.ApplicationSpaceResponse
				if currentState == "after_update1" {
					resp = indykite.ApplicationSpaceResponse{
						ID:          appSpaceID,
						CustomerID:  customerID,
						Name:        "acme0",
						DisplayName: "acme0",
						Description: "Another AppSpace description",
						Region:      "europe-west1",
						IKGSize:     "4GB",
						IKGStatus:   "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				} else {
					resp = indykite.ApplicationSpaceResponse{
						ID:          appSpaceID,
						CustomerID:  customerID,
						Name:        "acme0",
						DisplayName: "Some new display name",
						Region:      "europe-west1",
						IKGSize:     "4GB",
						IKGStatus:   "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, appSpaceID):
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
					Config: fmt.Sprintf(tfConfigDefSimple, "", "Just some AppSpace description", ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceNameSimple),
					),
				},
				{
					ResourceName:  resourceNameSimple,
					ImportState:   true,
					ImportStateId: appSpaceID,
				},
				{
					Config: fmt.Sprintf(tfConfigDefSimple, "", "Another AppSpace description", ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceNameSimple),
					),
				},
				{
					Config: fmt.Sprintf(tfConfigDefSimple, "Some new display name", "", ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceNameSimple),
					),
				},
				{
					Config:      fmt.Sprintf(tfConfigDefSimple, "Some new display name", "", ""),
					Destroy:     true,
					ExpectError: regexp.MustCompile("Cannot destroy instance"),
				},
				{
					Config: fmt.Sprintf(tfConfigDefSimple, "Some new display name", "", turnOffDelProtection),
				},
			},
		})
	})

	It("Test all CRUD with DB Connection", func() {
		turnOffDelProtection := "deletion_protection=false"
		tfConfigDef :=
			`resource "indykite_application_space" "development" {
					customer_id = "` + customerID + `"
					name = "acme"
					display_name = "%s"
					description = "%s"
					region = "us-east1"
					ikg_size = "4GB"
					replica_region = "us-west1"
					%s
					%s
				}`

		dbConnConfig := `db_connection {
			url = "postgresql://localhost:5432/test"
			username = "testuser"
			password = "testpass"
			name = "testdb"
		}`

		createTime := time.Now()
		updateTime := time.Now()
		currentState := "initial"

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects"):
				resp := indykite.ApplicationSpaceResponse{
					ID:            appSpaceID,
					CustomerID:    customerID,
					Name:          "acme",
					DisplayName:   "acme",
					Description:   "Just some AppSpace description",
					Region:        "us-east1",
					IKGSize:       "4GB",
					ReplicaRegion: "us-west1",
					DBConnection: &indykite.DBConnection{
						URL:      "postgresql://localhost:5432/test",
						Username: "testuser",
						Password: "testpass",
						Name:     "testdb",
					},
					IKGStatus:  "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
					CreateTime: createTime,
					UpdateTime: updateTime,
				}
				currentState = "after_create"
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, appSpaceID):
				var resp indykite.ApplicationSpaceResponse
				switch currentState {
				case "initial", "after_create":
					resp = indykite.ApplicationSpaceResponse{
						ID:            appSpaceID,
						CustomerID:    customerID,
						Name:          "acme",
						DisplayName:   "acme",
						Description:   "Just some AppSpace description",
						Region:        "us-east1",
						IKGSize:       "4GB",
						ReplicaRegion: "us-west1",
						DBConnection: &indykite.DBConnection{
							URL:      "postgresql://localhost:5432/test",
							Username: "testuser",
							Password: "testpass",
							Name:     "testdb",
						},
						IKGStatus:  "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
						CreateTime: createTime,
						UpdateTime: updateTime,
					}
				case "after_update1":
					resp = indykite.ApplicationSpaceResponse{
						ID:            appSpaceID,
						CustomerID:    customerID,
						Name:          "acme",
						DisplayName:   "acme",
						Description:   "Another AppSpace description",
						Region:        "us-east1",
						IKGSize:       "4GB",
						ReplicaRegion: "us-west1",
						DBConnection: &indykite.DBConnection{
							URL:      "postgresql://localhost:5432/test",
							Username: "testuser",
							Password: "testpass",
							Name:     "testdb",
						},
						IKGStatus:  "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
						CreateTime: createTime,
						UpdateTime: time.Now(),
					}
				case "after_update2":
					resp = indykite.ApplicationSpaceResponse{
						ID:            appSpaceID,
						CustomerID:    customerID,
						Name:          "acme",
						DisplayName:   "Some new display name",
						Region:        "us-east1",
						IKGSize:       "4GB",
						ReplicaRegion: "us-west1",
						DBConnection: &indykite.DBConnection{
							URL:      "postgresql://localhost:5432/test",
							Username: "testuser",
							Password: "testpass",
							Name:     "testdb",
						},
						IKGStatus:  "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
						CreateTime: createTime,
						UpdateTime: time.Now(),
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, appSpaceID):
				var reqBody map[string]any
				_ = json.NewDecoder(r.Body).Decode(&reqBody)

				if reqBody["description"] != nil && strings.Contains(reqBody["description"].(string), "Another") {
					currentState = "after_update1"
				} else {
					currentState = "after_update2"
				}

				var resp indykite.ApplicationSpaceResponse
				if currentState == "after_update1" {
					resp = indykite.ApplicationSpaceResponse{
						ID:            appSpaceID,
						CustomerID:    customerID,
						Name:          "acme",
						DisplayName:   "acme",
						Description:   "Another AppSpace description",
						Region:        "us-east1",
						IKGSize:       "4GB",
						ReplicaRegion: "us-west1",
						DBConnection: &indykite.DBConnection{
							URL:      "postgresql://localhost:5432/test",
							Username: "testuser",
							Password: "testpass",
							Name:     "testdb",
						},
						IKGStatus:  "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
						CreateTime: createTime,
						UpdateTime: time.Now(),
					}
				} else {
					resp = indykite.ApplicationSpaceResponse{
						ID:            appSpaceID,
						CustomerID:    customerID,
						Name:          "acme",
						DisplayName:   "Some new display name",
						Region:        "us-east1",
						IKGSize:       "4GB",
						ReplicaRegion: "us-west1",
						DBConnection: &indykite.DBConnection{
							URL:      "postgresql://localhost:5432/test",
							Username: "testuser",
							Password: "testpass",
							Name:     "testdb",
						},
						IKGStatus:  "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
						CreateTime: createTime,
						UpdateTime: time.Now(),
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, appSpaceID):
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
					Config: fmt.Sprintf(tfConfigDef, "", "Just some AppSpace description", dbConnConfig, ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceName),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: appSpaceID,
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "", "Another AppSpace description", dbConnConfig, ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceName),
					),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "Some new display name", "", dbConnConfig, ""),
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceResourceDataExists(resourceName),
					),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, "Some new display name", "", dbConnConfig, ""),
					Destroy:     true,
					ExpectError: regexp.MustCompile("Cannot destroy instance"),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "Some new display name", "", dbConnConfig, turnOffDelProtection),
				},
			},
		})
	})
})

func testAppSpaceResourceDataExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != appSpaceID {
			return errors.New("ID does not match")
		}
		attrs := rs.Primary.Attributes

		keys := Keys{
			"id": Equal(appSpaceID),
			"%":  Not(BeEmpty()),

			"customer_id": Equal(customerID),
			"name":        Not(BeEmpty()),
			"region":      Not(BeEmpty()),
			"ikg_size":    Not(BeEmpty()),
			"create_time": Not(BeEmpty()),
			"update_time": Not(BeEmpty()),
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
	}
}

var _ = Describe("Resource ApplicationSpace Import by Name", func() {
	const resourceName = "indykite_application_space.development"
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
			`resource "indykite_application_space" "development" {
				customer_id = "` + customerID + `"
				name = "acme"
				display_name = "ACME"
				description = "Just some AppSpace description"
				region = "europe-west1"
				ikg_size = "4GB"
				deletion_protection = false
			}`

		createTime := time.Now()
		updateTime := time.Now()

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/projects"):
				resp := indykite.ApplicationSpaceResponse{
					ID:          appSpaceID,
					CustomerID:  customerID,
					Name:        "acme",
					DisplayName: "ACME",
					Description: "Just some AppSpace description",
					Region:      "europe-west1",
					IKGSize:     "4GB",
					IKGStatus:   "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
					CreateTime:  createTime,
					UpdateTime:  updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/projects/"):
				// Support both ID and name?location=customerID formats
				// Check if it's a name-based lookup or ID-based lookup
				pathAfterProjects := strings.TrimPrefix(r.URL.Path, "/configs/v1/projects/")
				isNameLookup := strings.Contains(pathAfterProjects, "acme")
				isIDLookup := strings.Contains(pathAfterProjects, appSpaceID)

				var resp indykite.ApplicationSpaceResponse
				if isNameLookup || isIDLookup {
					resp = indykite.ApplicationSpaceResponse{
						ID:          appSpaceID,
						CustomerID:  customerID,
						Name:        "acme",
						DisplayName: "ACME",
						Description: "Just some AppSpace description",
						Region:      "europe-west1",
						IKGSize:     "4GB",
						IKGStatus:   "APP_SPACE_IKG_STATUS_STATUS_ACTIVE",
						CreateTime:  createTime,
						UpdateTime:  updateTime,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)
				} else {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
				}

			case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, appSpaceID):
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
						testAppSpaceResourceDataExists(resourceName),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: "acme?location=" + customerID,
				},
			},
		})
	})
})
