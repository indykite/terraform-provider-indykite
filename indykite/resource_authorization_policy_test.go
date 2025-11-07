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

var _ = Describe("Resource Authorization Policy config", func() {
	const resourceName = "indykite_authorization_policy.wonka"
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
		createTime := time.Now()
		updateTime := time.Now()

		//nolint:lll // Long JSON policy string
		policyJSON := `{"meta":{"policyVersion":"1.0-indykite"},"subject":{"type":"Person"},"actions":["CAN_DRIVE","CAN_PERFORM_SERVICE"],"resource":{"type":"Car"},"condition":{"cypher":"MATCH (subject:Person)-[:PART_OF]->(:Household)-[:DISPOSES]->(resource:Car)"}}`

		// Track whether update has been called
		updated := false

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/authorization-policies"):
				resp := indykite.AuthorizationPolicyResponse{
					ID:          sampleID,
					Name:        "wonka-authorization-policy-config",
					DisplayName: "Wonka Authorization for chocolate receipts",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Policy:      policyJSON,
					Status:      "active",
					CreateTime:  createTime,
					UpdateTime:  updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
				var resp indykite.AuthorizationPolicyResponse
				if updated {
					// Return updated state
					resp = indykite.AuthorizationPolicyResponse{
						ID:          sampleID,
						Name:        "wonka-authorization-policy-config",
						Description: "Description of the best Authz Policies by Wonka inc.",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Policy:      policyJSON,
						Status:      "active",
						Tags:        []string{"test", "wonka"},
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				} else {
					// Return initial state (same as POST response)
					resp = indykite.AuthorizationPolicyResponse{
						ID:          sampleID,
						Name:        "wonka-authorization-policy-config",
						DisplayName: "Wonka Authorization for chocolate receipts",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Policy:      policyJSON,
						Status:      "active",
						CreateTime:  createTime,
						UpdateTime:  updateTime,
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
				updated = true
				resp := indykite.AuthorizationPolicyResponse{
					ID:          sampleID,
					Name:        "wonka-authorization-policy-config",
					Description: "Description of the best Authz Policies by Wonka inc.",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Policy:      policyJSON,
					Status:      "active",
					Tags:        []string{"test", "wonka"},
					CreateTime:  createTime,
					UpdateTime:  time.Now(),
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

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Error cases should be always first, easier to avoid missing mocks or incomplete plan
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						name = "wonka-authorization-policy-config"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "location" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
					}
					`,
					ExpectError: regexp.MustCompile(`The argument "name" is required, but no definition was found`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-authorization-policy-config"

						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`The argument "status" is required, but no definition was found.`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						customer_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-authorization-policy-config"
						status = "active"

						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "customer_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						app_space_id = "69857924-098e-4f11-8800-62b10bb188ea"
						name = "wonka-authorization-policy-config"
						status = "active"

						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(
						`Can't configure a value for "app_space_id": its value will be decided`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "something-invalid"
						name = "wonka-authorization-policy-config"
						status = "active"

						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`expected to have 'gid:' prefix`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						name = "Invalid Name @#$"
						status = "active"

						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`Value can have lowercase letters, digits, or hyphens.`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						name = "wonka-authorization-policy-config"
						status = "active"
					}
					`,
					ExpectError: regexp.MustCompile(`argument "json" is required`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + customerID + `"
						name = "wonka-authorization-policy-config"
						status = "active"

						json = "not valid json"
					}
					`,
					ExpectError: regexp.MustCompile(`"json" contains an invalid JSON`),
				},
				{
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = ""
						name = "wonka-authorization-policy-config"
						status = "active"
						json = "{}"
					}
					`,
					ExpectError: regexp.MustCompile(`expected to have 'gid:' prefix`),
				},

				// ---- Run mocked tests here ----
				{
					//nolint:lll
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-authorization-policy-config"
						display_name = "Wonka Authorization for chocolate receipts"
						status = "active"

						json = "{\"meta\":{\"policyVersion\":\"1.0-indykite\"},\"subject\":{\"type\":\"Person\"},\"actions\":[\"CAN_DRIVE\",\"CAN_PERFORM_SERVICE\"],\"resource\":{\"type\":\"Car\"},\"condition\":{\"cypher\":\"MATCH (subject:Person)-[:PART_OF]->(:Household)-[:DISPOSES]->(resource:Car)\"}}" //nolint:lll
					}`,

					Check: resource.ComposeTestCheckFunc(testAuthorizationPolicyResourceDataExists(
						resourceName,
						sampleID,
						nil,
					)),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: sampleID,
				},
				{
					//nolint:lll
					Config: `resource "indykite_authorization_policy" "wonka" {
						location = "` + appSpaceID + `"
						name = "wonka-authorization-policy-config"
						description = "Description of the best Authz Policies by Wonka inc."
						status = "active"
						tags = ["test", "wonka"]

						json = "{\"meta\":{\"policyVersion\":\"1.0-indykite\"},\"subject\":{\"type\":\"Person\"},\"actions\":[\"CAN_DRIVE\",\"CAN_PERFORM_SERVICE\"],\"resource\":{\"type\":\"Car\"},\"condition\":{\"cypher\":\"MATCH (subject:Person)-[:PART_OF]->(:Household)-[:DISPOSES]->(resource:Car)\"}}" //nolint:lll
					}
					`,
					Check: resource.ComposeTestCheckFunc(testAuthorizationPolicyResourceDataExists(
						resourceName,
						sampleID,
						Keys{
							"tags.#": Equal("2"),
							"tags.0": Equal("test"),
							"tags.1": Equal("wonka"),
						},
					)),
				},
			},
		})
	})
})

func testAuthorizationPolicyResourceDataExists(
	n string,
	expectedID string,
	extraKeys Keys,
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

			"location":     Not(BeEmpty()),
			"customer_id":  Equal(customerID),
			"app_space_id": Equal(appSpaceID),
			"name":         Not(BeEmpty()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
			"json":         Not(BeEmpty()),
			"status":       Not(BeEmpty()),
		}

		for k, v := range extraKeys {
			keys[k] = v
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
	}
}

var _ = Describe("Resource Authorization Policy Import by Name", func() {
	const resourceName = "indykite_authorization_policy.wonka"
	var (
		mockServer *httptest.Server
		provider   *schema.Provider
		policyID   = "gid:AAAAAdVzaGFyZWRfcG9saWN5X2lk"
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
			`resource "indykite_authorization_policy" "wonka" {
				location = "` + appSpaceID + `"
				name = "wonka-policy"
				display_name = "Wonka Policy"
				description = "Policy for Wonka"
				status = "active"
				json = jsonencode({
					"meta": {"policyVersion": "1.0-indykite"},
					"subject": {"type": "DigitalTwin"},
					"actions": ["ACTION1"],
					"resource": {"type": "Asset"},
					"condition": {"cypher": "MATCH (subject:DigitalTwin)"}
				})
			}`

		createTime := time.Now()
		updateTime := time.Now()

		//nolint:lll // Long JSON policy string
		policyJSON := `{"meta":{"policyVersion":"1.0-indykite"},"subject":{"type":"DigitalTwin"},"actions":["ACTION1"],"resource":{"type":"Asset"},"condition":{"cypher":"MATCH (subject:DigitalTwin)"}}`

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/authorization-policies"):
				resp := indykite.AuthorizationPolicyResponse{
					ID:          policyID,
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Name:        "wonka-policy",
					DisplayName: "Wonka Policy",
					Description: "Policy for Wonka",
					Status:      "ACTIVE",
					CreateTime:  createTime,
					UpdateTime:  updateTime,
					Policy:      policyJSON,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/authorization-policies/"):
				// Support both ID and name?location=projectID formats
				// Check if it's a name-based lookup or ID-based lookup
				pathAfterPolicies := strings.TrimPrefix(r.URL.Path, "/configs/v1/authorization-policies/")
				isNameLookup := strings.Contains(pathAfterPolicies, "wonka-policy")
				isIDLookup := strings.Contains(pathAfterPolicies, policyID)

				var resp indykite.AuthorizationPolicyResponse
				if isNameLookup || isIDLookup {
					resp = indykite.AuthorizationPolicyResponse{
						ID:          policyID,
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Name:        "wonka-policy",
						DisplayName: "Wonka Policy",
						Description: "Policy for Wonka",
						Status:      "ACTIVE",
						CreateTime:  createTime,
						UpdateTime:  updateTime,
						Policy:      policyJSON,
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)
				} else {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
				}

			case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, policyID):
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
						testAuthorizationPolicyResourceDataExists(resourceName, policyID, nil),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: "wonka-policy?location=" + appSpaceID,
				},
			},
		})
	})
})
