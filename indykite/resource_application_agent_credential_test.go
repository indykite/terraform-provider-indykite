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

var _ = Describe("Resource ApplicationAgentCredential", func() {
	const resourceName = "indykite_application_agent_credential.development"
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
			`resource "indykite_application_agent_credential" "development" {
				app_agent_id = "` + appAgentID + `"
				display_name = "%s"
				%s
			}`

		createTime := time.Now()
		currentCredType := "jwk" // Track which credential type is being tested

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/application-agent-credentials"):
				var resp indykite.ApplicationAgentCredentialResponse
				if currentCredType == "jwk" {
					resp = indykite.ApplicationAgentCredentialResponse{
						ID:                 appAgentCredID,
						CustomerID:         customerID,
						AppSpaceID:         appSpaceID,
						ApplicationID:      applicationID,
						ApplicationAgentID: appAgentID,
						Kid:                "EfUEiFnOzA5PCp8SSksp7iXv7cHRehCsIGo6NAQ9H7w",
						CreateTime:         createTime,
						CreateBy:           "creator-id",
						AgentConfig:        `{"appAgentId":"` + appAgentID + `","endpoint":"https://example.com"}`,
					}
				} else {
					resp = indykite.ApplicationAgentCredentialResponse{
						ID:                 appAgentCredID2,
						CustomerID:         customerID,
						AppSpaceID:         appSpaceID,
						ApplicationID:      applicationID,
						ApplicationAgentID: appAgentID,
						DisplayName:        "OPA credentials",
						Kid:                "BgQgo-U3kF7kf2dXLKFPNcl3haR8k1VD2nTTvp0GBhI",
						CreateTime:         createTime,
						CreateBy:           "creator-id",
						AgentConfig:        `{"appAgentId":"` + appAgentID + `","endpoint":"https://example.com"}`,
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet:
				var resp indykite.ApplicationAgentCredentialResponse
				if strings.Contains(r.URL.Path, appAgentCredID) {
					resp = indykite.ApplicationAgentCredentialResponse{
						ID:                 appAgentCredID,
						CustomerID:         customerID,
						AppSpaceID:         appSpaceID,
						ApplicationID:      applicationID,
						ApplicationAgentID: appAgentID,
						Kid:                "EfUEiFnOzA5PCp8SSksp7iXv7cHRehCsIGo6NAQ9H7w",
						CreateTime:         createTime,
						CreateBy:           "creator-id",
						AgentConfig:        `{"appAgentId":"` + appAgentID + `","endpoint":"https://example.com"}`,
					}
				} else {
					resp = indykite.ApplicationAgentCredentialResponse{
						ID:                 appAgentCredID2,
						CustomerID:         customerID,
						AppSpaceID:         appSpaceID,
						ApplicationID:      applicationID,
						ApplicationAgentID: appAgentID,
						DisplayName:        "OPA credentials",
						Kid:                "BgQgo-U3kF7kf2dXLKFPNcl3haR8k1VD2nTTvp0GBhI",
						CreateTime:         createTime,
						CreateBy:           "creator-id",
						AgentConfig:        `{"appAgentId":"` + appAgentID + `","endpoint":"https://example.com"}`,
					}
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
					Config:      fmt.Sprintf(tfConfigDef, "", `app_space_id = "`+appSpaceID+`"`),
					ExpectError: regexp.MustCompile("Value for unconfigurable attribute"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, "", `application_id = "`+applicationID+`"`),
					ExpectError: regexp.MustCompile("Value for unconfigurable attribute"),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "", `public_key_pem = "PEM_Placeholder"
					public_key_jwk = "{}"`),
					ExpectError: regexp.MustCompile(`"public_key_pem": conflicts with public_key_jwk`),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, "", `public_key_jwk = "abc"`),
					ExpectError: regexp.MustCompile(`"public_key_jwk" contains an invalid JSON`),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, "", `public_key_jwk = "{}"`),
					ExpectError: regexp.MustCompile(`length of public_key_jwk to be in the range \(96 - 8192\)`),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, "", `public_key_pem = "abc"`),
					ExpectError: regexp.MustCompile(`length of public_key_pem to be in the range \(256 - 8192\)`),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, "", `public_key_pem = "`+strings.Repeat("a", 300)+`"`),
					ExpectError: regexp.MustCompile(
						`key must starts with '-----BEGIN PUBLIC KEY-----' and ends with '-----END PUBLIC KEY-----'`),
				},

				// Test JWK credential
				{
					Config: fmt.Sprintf(tfConfigDef, "", `public_key_jwk = <<-EOT
						{
							"kty": "EC",
							"use": "sig",
							"crv": "P-256",
							"kid": "xuyd5-9bT0L09mi810mycfREAxBG3KnpctlGQCYtCdM",
							"x": "o7LnIMhCPXFV91sE5EKQh8QZ9U6csUqgSENaKt3T0I4",
							"y": "o1Wwws1ZeoSwh_yN8_jeFOWHwK2n_6ow15SxIHyAnpE",
							"alg": "ES256"
						}
						EOT
						expire_time = "`+time.Now().Add(time.Hour).UTC().Format(time.RFC3339)+`"
						`),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentCredResourceDataExists(resourceName, appAgentCredID, "jwk"),
					),
				},
				{
					// In-place update (same config, tests double-check)
					Config: fmt.Sprintf(tfConfigDef, "", `public_key_jwk = <<-EOT
						{
							"kty": "EC",
							"use": "sig",
							"crv": "P-256",
							"kid": "xuyd5-9bT0L09mi810mycfREAxBG3KnpctlGQCYtCdM",
							"x": "o7LnIMhCPXFV91sE5EKQh8QZ9U6csUqgSENaKt3T0I4",
							"y": "o1Wwws1ZeoSwh_yN8_jeFOWHwK2n_6ow15SxIHyAnpE",
							"alg": "ES256"
						}
						EOT
						expire_time = "`+time.Now().Add(time.Hour).UTC().Format(time.RFC3339)+`"
						`),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentCredResourceDataExists(resourceName, appAgentCredID, "jwk"),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: appAgentCredID,
				},
			},
		})

		// Switch to PEM credentials test
		currentCredType = "pem"

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(tfConfigDef, "OPA credentials", `public_key_pem = <<-EOT
						-----BEGIN PUBLIC KEY-----
						MIGeMA0GCSqGSIb3DQEBAQUAA4GMADCBiAKBgHRMVhhoOrM0ldxMoaXQ6d9z9aBw
						+BnjNPxKKMeyRYNHZW18CK2Av28AXla0sXca8N30lHcaCV0/DfZ+Kg4UC8aNSDlH
						hEhSGYucKHN+kdf56qmA+odF87gvunkwzJuZddBYAKv9pevZBIn/e3TG8xIfI0S7
						j8ZGOIOYXO64OPXFAgMBAAE=
						-----END PUBLIC KEY-----
						EOT`),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentCredResourceDataExists(resourceName, appAgentCredID2, "pem"),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: appAgentCredID2,
				},
			},
		})
	})
})

func testAppAgentCredResourceDataExists(
	n string,
	expectedID string,
	credType string,
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
			"id":             Equal(expectedID),
			"%":              Not(BeEmpty()),
			"customer_id":    Equal(customerID),
			"app_space_id":   Equal(appSpaceID),
			"application_id": Equal(applicationID),
			"app_agent_id":   Equal(appAgentID),
			"kid":            Not(BeEmpty()),
			"create_time":    Not(BeEmpty()),
			"agent_config":   ContainSubstring(appAgentID),
		}

		switch credType {
		case "jwk":
			keys["public_key_jwk"] = ContainSubstring("xuyd5-9bT0L09mi810mycfREAxBG3KnpctlGQCYtCdM")
			keys["expire_time"] = Not(BeEmpty())
		case "pem":
			keys["public_key_pem"] = ContainSubstring("BEGIN PUBLIC KEY")
			keys["display_name"] = Equal("OPA credentials")
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), rs.Primary.Attributes)
	}
}
