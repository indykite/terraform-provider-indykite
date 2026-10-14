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
	"maps"
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

		mockServer = newServiceAccountCredentialAPI()

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

	It("Test expire_time omitted and service_account_config kept after create", func() {
		tfConfig := `resource "indykite_service_account_credential" "development" {
			service_account_id = "` + serviceAccountID + `"
			display_name = "Service Account Credential"
		}`

		mockServer = newServiceAccountCredentialAPI()

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
					// Create; the implicit plan after apply fails if the API expiry forces replacement.
					Config: tfConfig,
					Check: resource.ComposeTestCheckFunc(
						testServiceAccountCredResourceDataExists(resourceName, serviceAccountCredID),
						resource.TestCheckResourceAttr(resourceName, "expire_time", farFutureExpireTime),
						resource.TestCheckResourceAttrWith(resourceName, "service_account_config",
							func(value string) error {
								var config map[string]any
								if err := json.Unmarshal([]byte(value), &config); err != nil {
									return fmt.Errorf("service_account_config is not a JSON object: %w", err)
								}
								return convertOmegaMatcherToError(And(
									HaveKeyWithValue("serviceAccountId", serviceAccountID),
									HaveKeyWithValue("privateKeyJWK", HaveKeyWithValue("kty", "EC")),
								), config)
							}),
					),
				},
				{
					// Refresh must keep the configuration the read response does not return.
					Config:   tfConfig,
					PlanOnly: true,
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: serviceAccountCredID,
					ImportStateCheck: func(states []*terraform.InstanceState) error {
						if len(states) != 1 {
							return fmt.Errorf("expected 1 imported state, got %d", len(states))
						}
						// The configuration is only returned on create, so an import cannot recover it.
						return convertOmegaMatcherToError(And(
							HaveKeyWithValue("expire_time", farFutureExpireTime),
							Not(HaveKeyWithValue("service_account_config", Not(BeEmpty()))),
						), states[0].Attributes)
					},
				},
			},
		})
	})

	DescribeTable("service_account_config conversion",
		func(raw string, expected string, expectErr bool) {
			value, err := indykite.ConvertServiceAccountConfig(json.RawMessage(raw))
			if expectErr {
				Expect(err).To(HaveOccurred())
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(value).To(Equal(expected))
		},
		Entry("object is stored as compact JSON", `{ "serviceAccountId": "gid:x", "endpoint": "e" }`,
			`{"serviceAccountId":"gid:x","endpoint":"e"}`, false),
		Entry("string is unquoted", `"{\"serviceAccountId\":\"gid:x\"}"`, `{"serviceAccountId":"gid:x"}`, false),
		Entry("missing value", ``, ``, false),
		Entry("null value", `null`, ``, false),
		Entry("invalid JSON", `{"serviceAccountId":`, ``, true),
	)
})

// farFutureExpireTime is the expire_time the API sets on a credential created without one.
const farFutureExpireTime = "9999-12-31T23:59:59Z"

// newServiceAccountCredentialAPI mimics the Config API service account credential endpoints:
// service_account_config is a JSON object returned only by create, the read response omits it,
// and a credential created without expire_time gets farFutureExpireTime.
func newServiceAccountCredentialAPI() *httptest.Server {
	var credential map[string]any
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/service-account-credentials"):
			var req map[string]any
			Expect(json.NewDecoder(r.Body).Decode(&req)).To(Succeed())
			expireTime, _ := req["expire_time"].(string)
			if expireTime == "" {
				expireTime = farFutureExpireTime
			}
			credential = map[string]any{
				"id":                 serviceAccountCredID,
				"service_account_id": req["service_account_id"],
				"kid":                "EfUEiFnOzA5PCp8SSksp7iXv7cHRehCsIGo6NAQ9H7w",
				"display_name":       req["display_name"],
				"create_time":        "2026-10-14T06:49:01Z",
				"created_by":         serviceAccountID,
				"organization_id":    organizationID,
				"expire_time":        expireTime,
			}
			created := maps.Clone(credential)
			delete(created, "organization_id")
			created["service_account_config"] = map[string]any{
				"serviceAccountId":      req["service_account_id"],
				"endpoint":              "api.example.com",
				"privateKeyPKCS8":       "pkcs8-placeholder",
				"privateKeyPKCS8Base64": "pkcs8-base64-placeholder",
				"privateKeyJWK":         map[string]any{"kty": "EC", "use": "sig", "kid": "..."},
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, serviceAccountCredID) && credential != nil:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(credential)

		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

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
