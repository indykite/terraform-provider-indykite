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
	"net/http"
	"net/http/httptest"
	"regexp"
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

var _ = Describe("Data Source customer", func() {
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

	It("Test Read Customer", func() {
		createTime := time.Now()
		updateTime := time.Now()
		wonka := indykite.CustomerResponse{
			ID:          customerID,
			Name:        "wonka",
			DisplayName: "wonka",
			Description: "Just some description",
			CreateTime:  createTime,
			UpdateTime:  updateTime,
		}

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/configs/v1/organizations/current":
				// Get current organization - always return wonka
				// The only endpoint available for customers is /organizations/current
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(wonka)
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
				// Error cases must be always first
				{
					Config: `data "indykite_customer" "error" {
						name = "wonka-"
					}`,
					ExpectError: regexp.MustCompile("Value can have lowercase letters, digits, or hyphens"),
				},
				{
					Config: `data "indykite_customer" "error" {
						customer_id = "gid:not-valid-base64@#$"
					}`,
					ExpectError: regexp.MustCompile("valid Raw URL Base64 string with 'gid:' prefix"),
				},
				{
					Config: `data "indykite_customer" "wonka" {name = "acme"}`,
					ExpectError: regexp.MustCompile(
						`customer with name 'acme' not found \(current organization is 'wonka'\)`),
				},

				// Success cases
				// No parameters - should fetch current organization
				{
					Check: resource.ComposeTestCheckFunc(testDataSourceWonkaCustomer(&wonka, customerID)),
					Config: `data "indykite_customer" "wonka" {
					}`,
				},
				// With customer_id - should validate it matches
				{
					Check: resource.ComposeTestCheckFunc(testDataSourceWonkaCustomer(&wonka, customerID)),
					Config: `data "indykite_customer" "wonka" {
						customer_id = "` + customerID + `"
					}`,
				},
				// With name - should validate it matches
				{
					Check: resource.ComposeTestCheckFunc(testDataSourceWonkaCustomer(&wonka, customerID)),
					Config: `data "indykite_customer" "wonka" {
						name = "wonka"
					}`,
				},
			},
		})
	})
})

func testDataSourceWonkaCustomer(data *indykite.CustomerResponse, customerID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["data.indykite_customer.wonka"]
		if !ok {
			return errors.New("not found: `indykite_customer.wonka`")
		}

		if rs.Primary.ID != data.ID {
			return errors.New("ID does not match")
		}

		keys := Keys{
			"id": Equal(data.ID),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"name":         Equal(data.Name),
			"display_name": Equal(data.DisplayName),
			"description":  Equal(data.Description),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
		}
		if customerID != "" {
			keys["customer_id"] = Equal(customerID)
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}
