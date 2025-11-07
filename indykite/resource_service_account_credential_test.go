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

var _ = Describe("Resource ServiceAccountCredential", func() {
	const resourceName = "indykite_service_account_credential.development"
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
		tfConfigDef :=
			`resource "indykite_service_account_credential" "development" {
				service_account_id = "` + serviceAccountID + `"
				display_name = "%s"
				%s
			}`

		createTime := time.Now()
		serviceAccountConfig := `{"serviceAccountId": "` + serviceAccountID +
			`", "privateKeyJWK": {"kty":"EC", "use":"sig", "kid":"..."}}`

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/service-account-credentials"):
				resp := indykite.ServiceAccountCredentialResponse{
					ID:                   serviceAccountCredID,
					OrganizationID:       organizationID,
					ServiceAccountID:     serviceAccountID,
					Kid:                  "EfUEiFnOzA5PCp8SSksp7iXv7cHRehCsIGo6NAQ9H7w",
					ServiceAccountConfig: serviceAccountConfig,
					CreateTime:           createTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, serviceAccountCredID):
				resp := indykite.ServiceAccountCredentialResponse{
					ID:                   serviceAccountCredID,
					OrganizationID:       organizationID,
					ServiceAccountID:     serviceAccountID,
					DisplayName:          "Service Account Credential",
					Kid:                  "EfUEiFnOzA5PCp8SSksp7iXv7cHRehCsIGo6NAQ9H7w",
					ServiceAccountConfig: serviceAccountConfig,
					CreateTime:           createTime,
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
				// Errors case must be always first
				{
					Config:      fmt.Sprintf(tfConfigDef, "", `customer_id = "`+customerID+`"`),
					ExpectError: regexp.MustCompile("Value for unconfigurable attribute"),
				},
				{
					Config: fmt.Sprintf(
						tfConfigDef,
						"Service Account Credential",
						`expire_time = "`+time.Now().Add(time.Hour).UTC().Format(time.RFC3339)+`"`,
					),
					Check: resource.ComposeTestCheckFunc(
						testServiceAccountCredResourceDataExists(resourceName, serviceAccountCredID),
					),
				},
				{
					// In-place update (same config, tests double-check)
					Config: fmt.Sprintf(
						tfConfigDef,
						"Service Account Credential",
						`expire_time = "`+time.Now().Add(time.Hour).UTC().Format(time.RFC3339)+`"`,
					),
					Check: resource.ComposeTestCheckFunc(
						testServiceAccountCredResourceDataExists(resourceName, serviceAccountCredID),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: serviceAccountCredID,
				},
			},
		})
	})
})

func testServiceAccountCredResourceDataExists(
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

		keys := Keys{
			"id":                     Equal(expectedID),
			"%":                      Not(BeEmpty()),
			"customer_id":            Equal(customerID),
			"service_account_id":     Equal(serviceAccountID),
			"kid":                    Not(BeEmpty()),
			"create_time":            Not(BeEmpty()),
			"service_account_config": ContainSubstring(serviceAccountID),
			"display_name":           Equal("Service Account Credential"),
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), rs.Primary.Attributes)
	}
}
