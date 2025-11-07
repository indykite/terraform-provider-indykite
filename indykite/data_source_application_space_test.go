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

var _ = Describe("DataSource Application Space", func() {
	const resourceName = "data.indykite_application_space.development"
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

	It("Test load by ID and name", func() {
		createTime := time.Now()
		updateTime := time.Now()
		appSpaceResp := indykite.ApplicationSpaceResponse{
			ID:            appSpaceID,
			CustomerID:    customerID,
			Name:          "acme",
			DisplayName:   "Some Cool Display name",
			Description:   "Just some AppSpace description",
			CreateTime:    createTime,
			UpdateTime:    updateTime,
			Region:        "us-east1",
			IKGSize:       "4GB",
			ReplicaRegion: "us-west1",
			DBConnection: &indykite.DBConnection{
				URL:      "postgresql://localhost:5432/testdb",
				Username: "testuser",
				Password: "",
				Name:     "testdb",
			},
		}

		// Track which test step we're on
		nameFound := false

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/projects") &&
				r.URL.Query().Get("organization_id") == customerID:
				// List application spaces by customer
				w.WriteHeader(http.StatusOK)
				if nameFound {
					// Return the app space for successful name lookup
					_ = json.NewEncoder(w).Encode(indykite.ListApplicationSpacesResponse{
						AppSpaces: []indykite.ApplicationSpaceResponse{appSpaceResp},
					})
				} else {
					// Return empty list (name not found)
					_ = json.NewEncoder(w).Encode(indykite.ListApplicationSpacesResponse{
						AppSpaces: []indykite.ApplicationSpaceResponse{},
					})
				}
			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/projects/"+appSpaceID):
				// Read by ID - this also triggers nameFound for next test
				nameFound = true
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(appSpaceResp)
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
					Config: `data "indykite_application_space" "development" {
						name = "acme"
						app_space_id = "` + appSpaceID + `"
					}`,
					ExpectError: regexp.MustCompile("only one of `app_space_id,name` can be specified"),
				},
				{
					Config: `data "indykite_application_space" "development" {
						display_name = "anything"
					}`,
					ExpectError: regexp.MustCompile("one of `app_space_id,name` must be specified"),
				},
				{
					Config: `data "indykite_application_space" "development" {
						name = "anything"
					}`,
					ExpectError: regexp.MustCompile("\"name\": all of `customer_id,name` must be specified"),
				},
				{
					Config: `data "indykite_application_space" "development" {
						customer_id = "` + customerID + `"
						name = "acme"
					}`,
					ExpectError: regexp.MustCompile("application space with name 'acme' not found"),
				},

				// Success test cases
				{
					Config: `data "indykite_application_space" "development" {
						app_space_id = "` + appSpaceID + `"
					}`,
					Check: resource.ComposeTestCheckFunc(
						testAppSpaceDataExists(resourceName, &appSpaceResp, appSpaceID)),
				},
				{
					Config: `data "indykite_application_space" "development" {
						customer_id = "` + customerID + `"
						name = "acme"
					}`,
					Check: resource.ComposeTestCheckFunc(testAppSpaceDataExists(resourceName, &appSpaceResp, "")),
				},
			},
		})
	})

	It("Test list by multple names", func() {
		createTime := time.Now()
		updateTime := time.Now()
		appSpaceResp := indykite.ApplicationSpaceResponse{
			ID:            appSpaceID,
			CustomerID:    customerID,
			Name:          "acme",
			DisplayName:   "Some Cool Display name",
			Description:   "Just some AppSpace description",
			CreateTime:    createTime,
			UpdateTime:    updateTime,
			Region:        "us-east1",
			IKGSize:       "4GB",
			ReplicaRegion: "us-west1",
		}
		appSpaceResp2 := indykite.ApplicationSpaceResponse{
			ID:            sampleID,
			CustomerID:    customerID,
			Name:          "wonka",
			CreateTime:    createTime,
			UpdateTime:    updateTime,
			Region:        "europe-west1",
			IKGSize:       "8GB",
			ReplicaRegion: "europe-west2",
		}

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/projects") {
				// List application spaces - return both wrapped in ListApplicationSpacesResponse
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(indykite.ListApplicationSpacesResponse{
					AppSpaces: []indykite.ApplicationSpaceResponse{appSpaceResp, appSpaceResp2},
				})
			} else {
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
					Config: `data "indykite_application_spaces" "development" {
						filter = "acme"
						customer_id = "` + customerID + `"
					}`,
					ExpectError: regexp.MustCompile(
						`Inappropriate value for attribute "filter": list of string required`),
				},
				{
					Config: `data "indykite_application_spaces" "development" {
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile(`The argument "customer_id" is required`),
				},
				{
					Config: `data "indykite_application_spaces" "development" {
						filter = []
						customer_id = "` + customerID + `"
					}`,
					ExpectError: regexp.MustCompile("Attribute filter requires 1 item minimum, but config has only 0"),
				},
				{
					Config: `data "indykite_application_spaces" "development" {
						customer_id = "` + customerID + `"
						filter = [123]
					}`,
					ExpectError: regexp.MustCompile("Value can have lowercase letters, digits, or hyphens."),
				},
				{
					Config: `data "indykite_application_spaces" "development" {
						customer_id = "abc"
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile("expected to have 'gid:' prefix"),
				},
				{
					Config: `data "indykite_application_spaces" "development" {
						customer_id = "` + customerID + `"
						filter = ["acme"]
						app_spaces = []
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "app_spaces":`),
				},

				// Success test cases
				{
					Check: resource.ComposeTestCheckFunc(testAppSpaceListDataExists(
						"data.indykite_application_spaces.development",
						appSpaceResp,
						appSpaceResp2)),
					Config: `data "indykite_application_spaces" "development" {
						customer_id = "` + customerID + `"
						filter = ["acme", "some-another-name", "wonka"]
					}`,
				},
			},
		})
	})
})

func testAppSpaceDataExists(
	n string, data *indykite.ApplicationSpaceResponse, appSpaceID string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID != data.ID {
			return errors.New("ID does not match")
		}

		keys := Keys{
			"id": Equal(data.ID),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"customer_id":    Equal(data.CustomerID),
			"name":           Equal(data.Name),
			"display_name":   Equal(data.DisplayName),
			"description":    Equal(data.Description),
			"create_time":    Not(BeEmpty()),
			"update_time":    Not(BeEmpty()),
			"region":         Equal(data.Region),
			"ikg_size":       Equal(data.IKGSize),
			"replica_region": Equal(data.ReplicaRegion),
		}

		// Add db_connection checks based on whether it exists in the response
		if data.DBConnection != nil {
			keys["db_connection.#"] = Equal("1")
			keys["db_connection.0.%"] = Equal("4")
			keys["db_connection.0.url"] = Equal(data.DBConnection.URL)
			keys["db_connection.0.username"] = Equal(data.DBConnection.Username)
			keys["db_connection.0.password"] = Equal(data.DBConnection.Password)
			keys["db_connection.0.name"] = Equal(data.DBConnection.Name)
		} else {
			keys["db_connection.#"] = Equal("0")
		}
		if appSpaceID != "" {
			keys["app_space_id"] = Equal(data.ID)
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}

func testAppSpaceListDataExists(n string, data ...indykite.ApplicationSpaceResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		expectedID := "gid:AAAAAWluZHlraURlgAAAAAAAAA8/appSpaces/acme,some-another-name,wonka"
		if rs.Primary.ID != expectedID {
			return fmt.Errorf("expected ID to be '%s' got '%s'", expectedID, rs.Primary.ID)
		}

		keys := Keys{
			"id": Equal(expectedID),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"customer_id": Equal(customerID),

			"app_spaces.#": Equal(strconv.Itoa(len(data))), // This is Terraform helper
			"filter.#":     Equal("3"),
			"filter.0":     Equal("acme"),
			"filter.1":     Equal("some-another-name"),
			"filter.2":     Equal("wonka"),
		}

		for i := range data {
			d := &data[i]
			k := "app_spaces." + strconv.Itoa(i) + "."
			keys[k+"%"] = Not(BeEmpty()) // This is Terraform helper

			keys[k+"id"] = Equal(d.ID)
			keys[k+"customer_id"] = Equal(d.CustomerID)
			keys[k+"name"] = Equal(d.Name)
			keys[k+"display_name"] = Equal(d.DisplayName)
			keys[k+"description"] = Equal(d.Description)
			keys[k+"region"] = Equal(d.Region)
			keys[k+"ikg_size"] = Equal(d.IKGSize)
			keys[k+"replica_region"] = Equal(d.ReplicaRegion)
			// Note: db_connection is intentionally omitted from list view for security
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}
