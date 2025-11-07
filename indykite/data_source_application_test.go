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

var _ = Describe("DataSource Application", func() {
	const resourceName = "data.indykite_application.development"
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
		applicationResp := indykite.ApplicationResponse{
			ID:          applicationID,
			CustomerID:  customerID,
			AppSpaceID:  appSpaceID,
			Name:        "acme",
			DisplayName: "Some Cool Display name",
			Description: "Application description",
			CreateTime:  createTime,
			UpdateTime:  updateTime,
		}

		// Track which test step we're on
		nameFound := false

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/applications") &&
				r.URL.Query().Get("project_id") == appSpaceID:
				// List applications by app space
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(indykite.ListApplicationsResponse{
					Applications: []indykite.ApplicationResponse{applicationResp},
				})
			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/applications/acme") &&
				r.URL.Query().Get("location") == appSpaceID:
				// Get application by name
				if nameFound {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(applicationResp)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/applications/"+applicationID):
				// Read by ID - this also triggers nameFound for next test
				nameFound = true
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(applicationResp)
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
					Config: `data "indykite_application" "development" {
						customer_id = "` + customerID + `"
						app_space_id = "` + appSpaceID + `"
						name = "acme"
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "customer_id"`),
				},
				{
					Config: `data "indykite_application" "development" {
						name = "acme"
						application_id = "` + applicationID + `"
					}`,
					ExpectError: regexp.MustCompile("only one of `application_id,name` can be specified"),
				},
				{
					Config: `data "indykite_application" "development" {
						display_name = "anything"
					}`,
					ExpectError: regexp.MustCompile("one of `application_id,name` must be specified"),
				},
				{
					Config: `data "indykite_application" "development" {
						name = "anything"
					}`,
					ExpectError: regexp.MustCompile("\"name\": all of `app_space_id, name` must be specified"),
				},
				{
					Config: `data "indykite_application" "development" {
						app_space_id = "` + appSpaceID + `"
						name = "acme"
					}`,
					ExpectError: regexp.MustCompile("failed to get application by name: HTTP 404"),
				},

				// Success test cases
				{
					Config: `data "indykite_application" "development" {
						application_id = "` + applicationID + `"
					}`,
					Check: resource.ComposeTestCheckFunc(
						testApplicationDataExists(resourceName, &applicationResp, applicationID)),
				},
				{
					Config: `data "indykite_application" "development" {
					app_space_id = "` + appSpaceID + `"
					name = "acme"
				}`,
					Check: resource.ComposeTestCheckFunc(testApplicationDataExists(resourceName, &applicationResp, "")),
				},
			},
		})
	})

	It("Test list by multple names", func() {
		createTime := time.Now()
		updateTime := time.Now()
		applicationResp := indykite.ApplicationResponse{
			ID:          applicationID,
			CustomerID:  customerID,
			AppSpaceID:  appSpaceID,
			Name:        "acme",
			DisplayName: "Some Cool Display name",
			Description: "Just some AppSpace description",
			CreateTime:  createTime,
			UpdateTime:  updateTime,
		}
		applicationResp2 := indykite.ApplicationResponse{
			ID:         sampleID,
			CustomerID: customerID,
			AppSpaceID: appSpaceID,
			Name:       "wonka-bars",
			CreateTime: createTime,
			UpdateTime: updateTime,
		}

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/applications") {
				// List applications - return both wrapped in ListApplicationsResponse
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(indykite.ListApplicationsResponse{
					Applications: []indykite.ApplicationResponse{applicationResp, applicationResp2},
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
					Config: `data "indykite_applications" "development" {
						filter = "acme"
						app_space_id = "` + appSpaceID + `"
					}`,
					ExpectError: regexp.MustCompile(
						`Inappropriate value for attribute "filter": list of string required`),
				},
				{
					Config: `data "indykite_applications" "development" {
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile(`Missing required argument|app_space_id`),
				},
				{
					Config: `data "indykite_applications" "development" {
						filter = []
						app_space_id = "` + appSpaceID + `"
					}`,
					ExpectError: regexp.MustCompile("Attribute filter requires 1 item minimum, but config has only 0"),
				},
				{
					Config: `data "indykite_applications" "development" {
						app_space_id = "` + appSpaceID + `"
						filter = [123]
					}`,
					ExpectError: regexp.MustCompile("Value can have lowercase letters, digits, or hyphens."),
				},
				{
					Config: `data "indykite_applications" "development" {
						app_space_id = "abc"
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile("expected to have 'gid:' prefix"),
				},
				{
					Config: `data "indykite_applications" "development" {
						app_space_id = "` + appSpaceID + `"
						filter = ["acme"]
						applications = []
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "applications":`),
				},

				// Success test cases
				{
					Check: resource.ComposeTestCheckFunc(testApplicationListDataExists(
						"data.indykite_applications.development",
						applicationResp,
						applicationResp2)),
					Config: `data "indykite_applications" "development" {
						app_space_id = "` + appSpaceID + `"
						filter = ["acme", "some-another-name", "wonka-bars"]
					}`,
				},
			},
		})
	})
})

func testApplicationDataExists(
	n string, data *indykite.ApplicationResponse, applicationID string,
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

			"customer_id":  Equal(data.CustomerID),
			"app_space_id": Equal(data.AppSpaceID),
			"name":         Equal(data.Name),
			"display_name": Equal(data.DisplayName),
			"description":  Equal(data.Description),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
		}
		if applicationID != "" {
			keys["application_id"] = Equal(applicationID)
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}

func testApplicationListDataExists(n string, data ...indykite.ApplicationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		expectedID := "gid:AAAAAmluZHlraURlgAABDwAAAAA/apps/acme,some-another-name,wonka-bars"
		if rs.Primary.ID != expectedID {
			return fmt.Errorf("expected ID to be '%s' got '%s'", expectedID, rs.Primary.ID)
		}

		keys := Keys{
			"id": Equal(expectedID),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"app_space_id": Equal(appSpaceID),

			"applications.#": Equal(strconv.Itoa(len(data))), // This is Terraform helper
			"filter.#":       Equal("3"),
			"filter.0":       Equal("acme"),
			"filter.1":       Equal("some-another-name"),
			"filter.2":       Equal("wonka-bars"),
		}

		for i := range data {
			d := &data[i]
			k := "applications." + strconv.Itoa(i) + "."
			keys[k+"%"] = Not(BeEmpty()) // This is Terraform helper

			keys[k+"id"] = Equal(d.ID)
			keys[k+"customer_id"] = Equal(d.CustomerID)
			keys[k+"app_space_id"] = Equal(d.AppSpaceID)
			keys[k+"name"] = Equal(d.Name)
			keys[k+"display_name"] = Equal(d.DisplayName)
			keys[k+"description"] = Equal(d.Description)
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}
