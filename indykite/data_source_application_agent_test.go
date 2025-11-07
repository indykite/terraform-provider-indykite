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

var _ = Describe("DataSource ApplicationAgent", func() {
	const resourceName = "data.indykite_application_agent.development"
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
		appAgentResp := indykite.ApplicationAgentResponse{
			ID:             appAgentID,
			CustomerID:     customerID,
			AppSpaceID:     appSpaceID,
			ApplicationID:  applicationID,
			Name:           "acme",
			DisplayName:    "Some Cool Display name",
			Description:    "ApplicationAgent description",
			APIPermissions: []string{"Authorization", "Capture"},
			CreateTime:     createTime,
			UpdateTime:     updateTime,
		}

		// Track which test step we're on
		nameFound := false

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/application-agents") &&
				r.URL.Query().Get("project_id") == appSpaceID:
				// List application agents by app space
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(indykite.ListApplicationAgentsResponse{
					Agents: []indykite.ApplicationAgentResponse{appAgentResp},
				})
			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/application-agents/acme") &&
				r.URL.Query().Get("location") == appSpaceID:
				// Get application agent by name
				if nameFound {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(appAgentResp)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/application-agents/"+appAgentID):
				// Read by ID - this also triggers nameFound for next test
				nameFound = true
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(appAgentResp)
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
					Config: `data "indykite_application_agent" "development" {
						customer_id = "` + customerID + `"
						app_space_id = "` + appSpaceID + `"
						name = "acme"
						api_permissions = ["Authorization","Capture"]
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "customer_id"`),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						application_id = "` + applicationID + `"
						app_space_id = "` + appSpaceID + `"
						name = "acme"
						api_permissions = ["Authorization","Capture"]
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "application_id"`),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						name = "acme"
						api_permissions = ["Authorization","Capture"]
						app_agent_id = "` + applicationID + `"
					}`,
					ExpectError: regexp.MustCompile("only one of `app_agent_id,name` can be specified"),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						display_name = "anything"
						api_permissions = ["Authorization","Capture"]
					}`,
					ExpectError: regexp.MustCompile("one of `app_agent_id,name` must be specified"),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						name = "anything"
						api_permissions = ["Authorization","Capture"]
					}`,
					ExpectError: regexp.MustCompile("\"name\": all of `app_space_id,\\s*name` must be specified"),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						app_space_id = "` + appSpaceID + `"
						name = "acme"
						api_permissions = ["Authorization","Capture"]
					}`,
					ExpectError: regexp.MustCompile("failed to get application agent by name: HTTP 404"),
				},
				// Success test cases
				{
					Config: `data "indykite_application_agent" "development" {
						app_agent_id = "` + appAgentID + `"
						api_permissions = ["Authorization","Capture"]
					}`,
					Check: resource.ComposeTestCheckFunc(
						testApplicationAgentDataExists(resourceName, &appAgentResp, appAgentID),
					),
				},
				{
					Config: `data "indykite_application_agent" "development" {
						app_space_id = "` + appSpaceID + `"
						name = "acme"
						api_permissions = ["Authorization","Capture"]
					}`,
					Check: resource.ComposeTestCheckFunc(
						testApplicationAgentDataExists(resourceName, &appAgentResp, "")),
				},
			},
		})
	})

	It("Test list by multiple names", func() {
		createTime := time.Now()
		updateTime := time.Now()
		appAgentResp := indykite.ApplicationAgentResponse{
			ID:             appAgentID,
			CustomerID:     customerID,
			AppSpaceID:     appSpaceID,
			ApplicationID:  applicationID,
			Name:           "loompaland",
			DisplayName:    "Some Cool Display name",
			Description:    "Just some ApplicationAgent description",
			CreateTime:     createTime,
			UpdateTime:     updateTime,
			APIPermissions: []string{"Authorization", "Capture"},
		}
		appAgentResp2 := indykite.ApplicationAgentResponse{
			ID:             sampleID,
			CustomerID:     customerID,
			AppSpaceID:     appSpaceID,
			ApplicationID:  applicationID,
			Name:           "wonka-opa-agent",
			CreateTime:     createTime,
			UpdateTime:     updateTime,
			APIPermissions: []string{"Authorization", "Capture"},
		}

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/application-agents") {
				// List application agents - return both wrapped in ListApplicationAgentsResponse
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(indykite.ListApplicationAgentsResponse{
					Agents: []indykite.ApplicationAgentResponse{appAgentResp, appAgentResp2},
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
					Config: `data "indykite_application_agents" "development" {
						filter = "acme"
						app_space_id = "` + appSpaceID + `"
					}`,
					ExpectError: regexp.MustCompile(
						`Inappropriate value for attribute "filter": list of string required`),
				},
				{
					Config: `data "indykite_application_agents" "development" {
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile(`Missing required argument|app_space_id`),
				},
				{
					Config: `data "indykite_application_agents" "development" {
						filter = []
						app_space_id = "` + appSpaceID + `"
					}`,
					ExpectError: regexp.MustCompile("Attribute filter requires 1 item minimum, but config has only 0"),
				},
				{
					Config: `data "indykite_application_agents" "development" {
						app_space_id = "` + appSpaceID + `"
						filter = [123]
					}`,
					ExpectError: regexp.MustCompile("Value can have lowercase letters, digits, or hyphens."),
				},
				{
					Config: `data "indykite_application_agents" "development" {
						app_space_id = "abc"
						filter = ["acme"]
					}`,
					ExpectError: regexp.MustCompile("expected to have 'gid:' prefix"),
				},
				{
					Config: `data "indykite_application_agents" "development" {
						app_space_id = "` + appSpaceID + `"
						filter = ["acme"]
						app_agents = []
					}`,
					ExpectError: regexp.MustCompile(`Can't configure a value for "app_agents":`),
				},
				// Success test cases
				{
					Check: resource.ComposeTestCheckFunc(testApplicationAgentListDataExists(
						"data.indykite_application_agents.development",
						appAgentResp,
						appAgentResp2)),
					Config: `data "indykite_application_agents" "development" {
						app_space_id = "` + appSpaceID + `"
						filter = ["loompaland", "some-another-name", "wonka-opa-agent"]
					}`,
				},
			},
		})
	})
})

func testApplicationAgentDataExists(
	n string,
	data *indykite.ApplicationAgentResponse,
	appAgentID string,
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
			"id":                Equal(data.ID),
			"%":                 Not(BeEmpty()), // This is Terraform helper
			"api_permissions.#": Equal("2"),
			"api_permissions.0": Equal("Authorization"),
			"api_permissions.1": Equal("Capture"),

			"customer_id":    Equal(data.CustomerID),
			"app_space_id":   Equal(data.AppSpaceID),
			"application_id": Equal(data.ApplicationID),
			"name":           Equal(data.Name),
			"display_name":   Equal(data.DisplayName),
			"description":    Equal(data.Description),
			"create_time":    Not(BeEmpty()),
			"update_time":    Not(BeEmpty()),
		}
		if appAgentID != "" {
			keys["app_agent_id"] = Equal(appAgentID)
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}

func testApplicationAgentListDataExists(n string, data ...indykite.ApplicationAgentResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		expectedID := "gid:AAAAAmluZHlraURlgAABDwAAAAA/app_agents/loompaland,some-another-name,wonka-opa-agent"
		if rs.Primary.ID != expectedID {
			return fmt.Errorf("expected ID to be '%s' got '%s'", expectedID, rs.Primary.ID)
		}

		keys := Keys{
			"id": Equal(expectedID),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"app_space_id": Equal(appSpaceID),

			"app_agents.#": Equal(strconv.Itoa(len(data))), // This is Terraform helper
			"filter.#":     Equal("3"),
			"filter.0":     Equal("loompaland"),
			"filter.1":     Equal("some-another-name"),
			"filter.2":     Equal("wonka-opa-agent"),
		}

		for i := range data {
			d := &data[i]
			k := "app_agents." + strconv.Itoa(i) + "."
			keys[k+"%"] = Not(BeEmpty()) // This is Terraform helper
			keys[k+"api_permissions.#"] = Equal("2")
			keys[k+"api_permissions.0"] = Equal("Authorization")
			keys[k+"api_permissions.1"] = Equal("Capture")
			keys[k+"id"] = Equal(d.ID)
			keys[k+"customer_id"] = Equal(d.CustomerID)
			keys[k+"app_space_id"] = Equal(d.AppSpaceID)
			keys[k+"application_id"] = Equal(d.ApplicationID)
			keys[k+"name"] = Equal(d.Name)
			keys[k+"display_name"] = Equal(d.DisplayName)
			keys[k+"description"] = Equal(d.Description)
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}
