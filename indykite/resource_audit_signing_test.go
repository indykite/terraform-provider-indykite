// Copyright (c) 2026 IndyKite
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
	"sync"
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

const (
	auditSigningResourceName       = "indykite_audit_signing.development"
	auditSigningAuthParamsMaxPairs = 32
)

var _ = Describe("Resource AuditSigning", func() {
	const (
		resourceName = auditSigningResourceName
		gcpKeyRes    = "projects/p/locations/eu/keyRings/r/cryptoKeys/k/cryptoKeyVersions/1"
	)
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

	// maskAuthParams mirrors the API: keys are echoed back, values are blanked.
	maskAuthParams := func(in map[string]string) map[string]string {
		if in == nil {
			return nil
		}
		out := make(map[string]string, len(in))
		for k := range in {
			out[k] = ""
		}
		return out
	}

	It("Test CRUD of Audit Signing configuration", func() {
		tfConfigDef :=
			`resource "indykite_audit_signing" "development" {
				location = "%s"
				name = "%s"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()

		// The mock keeps the last written config so reads reflect create/update payloads,
		// with auth_params masked the way the real API does.
		var (
			mu     sync.Mutex
			stored indykite.AuditSigningResponse
		)

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/audit-signings"):
				var req indykite.CreateAuditSigningRequest
				_ = json.NewDecoder(r.Body).Decode(&req)

				if req.Name == "" || req.ProjectID != appSpaceID || req.Provider == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				for _, v := range req.AuthParams {
					if v == "" {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
				}

				stored = indykite.AuditSigningResponse{
					ID:          sampleID,
					Name:        req.Name,
					DisplayName: req.DisplayName,
					Description: req.Description,
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					Provider:    req.Provider,
					KeyResource: new(req.KeyResource),
					Kid:         new(req.Kid),
					AuthParams:  maskAuthParams(req.AuthParams),
					CreateTime:  createTime,
					UpdateTime:  updateTime,
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]string{"id": sampleID})

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(stored)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
				var req indykite.UpdateAuditSigningRequest
				_ = json.NewDecoder(r.Body).Decode(&req)
				if req.Provider == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				// The real API never returns secret material, so a masked value being sent
				// back on update would be a provider bug.
				for _, v := range req.AuthParams {
					if v == "" {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
				}
				if req.DisplayName != nil {
					stored.DisplayName = *req.DisplayName
				}
				if req.Description != nil {
					stored.Description = *req.Description
				}
				stored.Provider = req.Provider
				stored.KeyResource = new(req.KeyResource)
				stored.Kid = new(req.Kid)
				stored.AuthParams = maskAuthParams(req.AuthParams)
				stored.UpdateTime = time.Now()
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{"id": sampleID})

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

		platformManaged := `
		key_provider = "PLATFORM_MANAGED"
		`

		customerGCP := `
		key_provider = "CUSTOMER_GCP_KMS"
		key_resource = "` + gcpKeyRes + `"
		kid          = "audit-key-1"
		auth_params = {
			credentials_json = "{\"type\":\"service_account\"}"
		}
		`

		customerGCPRotated := `
		key_provider = "CUSTOMER_GCP_KMS"
		key_resource = "` + gcpKeyRes + `"
		kid          = "audit-key-2"
		auth_params = {
			credentials_json = "{\"type\":\"service_account\",\"v\":2}"
			project          = "p"
		}
		`

		var tooManyPairs strings.Builder
		for i := range auditSigningAuthParamsMaxPairs + 1 {
			fmt.Fprintf(&tooManyPairs, "p%d = \"v\"\n", i)
		}
		tooManyAuthParams := "auth_params = {\n" + tooManyPairs.String() + "}"

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				// Errors case must always come first
				{
					Config:      fmt.Sprintf(tfConfigDef, "ccc", "name", platformManaged),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", `key_provider = "CUSTOMER_VAULT"`),
					ExpectError: regexp.MustCompile(`expected key_provider to be one of`),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name",
						`auth_params = { credentials_json = "" }`),
					ExpectError: regexp.MustCompile(
						`expected length of auth_params\.credentials_json to be in the range`),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name",
						`auth_params = { "`+strings.Repeat("k", 257)+`" = "v" }`),
					ExpectError: regexp.MustCompile(`expected length of auth_params key "k+" to be in the range`),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", tooManyAuthParams),
					ExpectError: regexp.MustCompile(`expected at most 32 auth_params entries, got 33`),
				},
				{
					// Customer managed providers must name the key; checked at plan time.
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name",
						`key_provider = "CUSTOMER_GCP_KMS"
						kid          = "audit-key-1"`),
					ExpectError: regexp.MustCompile(`"key_resource" is required when key_provider is CUSTOMER_GCP_KMS`),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name",
						`key_provider = "CUSTOMER_AZURE_KEY_VAULT"
						key_resource = "https://vault.example.net/keys/audit/1"`),
					ExpectError: regexp.MustCompile(`"kid" is required when key_provider is CUSTOMER_AZURE_KEY_VAULT`),
				},

				{
					// Create and Read with default provider
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-audit-signing",
						`display_name = "Display name of Audit Signing"`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAuditSigningResourceDataExists("PLATFORM_MANAGED", "", "", nil),
					),
				},
				{
					// Import by ID
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateId:     sampleID,
					ImportStateVerify: true,
				},
				{
					// Update to a customer managed key and Read; the mock masks auth_params
					// so the state must retain the configured secret value.
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-audit-signing",
						`description = "audit signing description"
						`+customerGCP+``,
					),
					Check: resource.ComposeTestCheckFunc(
						testAuditSigningResourceDataExists("CUSTOMER_GCP_KMS",
							gcpKeyRes, "audit-key-1",
							map[string]string{"credentials_json": `{"type":"service_account"}`}),
					),
				},
				{
					// Plan must be empty: masked values from the API do not cause drift.
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-audit-signing",
						`description = "audit signing description"
						`+customerGCP+``,
					),
					PlanOnly: true,
				},
				{
					// Rotate the key and change the auth_params set
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-audit-signing",
						`description = "audit signing description"
						`+customerGCPRotated+``,
					),
					Check: resource.ComposeTestCheckFunc(
						testAuditSigningResourceDataExists("CUSTOMER_GCP_KMS",
							gcpKeyRes, "audit-key-2",
							map[string]string{
								"credentials_json": `{"type":"service_account","v":2}`,
								"project":          "p",
							}),
					),
				},
			},
		})
	})

	It("Test import by name with location", func() {
		tfConfigDef := `resource "indykite_audit_signing" "development" {
				location = "%s"
				name = "%s"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()
		keyResource := "arn:aws:kms:eu-west-1:123456789012:key/abc"
		kid := "aws-audit"

		var (
			mu          sync.Mutex
			description string
		)
		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			resp := indykite.AuditSigningResponse{
				ID:          sampleID,
				Name:        "wonka-audit",
				DisplayName: "Wonka Audit",
				Description: description,
				CustomerID:  customerID,
				AppSpaceID:  appSpaceID,
				Provider:    "CUSTOMER_AWS_KMS",
				KeyResource: &keyResource,
				Kid:         &kid,
				AuthParams:  map[string]string{"access_key_id": "", "secret_access_key": ""},
				CreateTime:  createTime,
				UpdateTime:  updateTime,
			}
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/audit-signings"):
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]string{"id": sampleID})

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
				// The backend replaces the whole config on update, so the provider must always
				// send the real auth_params, never the masked values it read back from the API.
				var req indykite.UpdateAuditSigningRequest
				_ = json.NewDecoder(r.Body).Decode(&req)
				if len(req.AuthParams) != 2 {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "auth_params must be sent in full"})
					return
				}
				for k, v := range req.AuthParams {
					if v == "" {
						w.WriteHeader(http.StatusBadRequest)
						_ = json.NewEncoder(w).Encode(map[string]string{"message": "masked value sent for " + k})
						return
					}
				}
				if req.Description != nil {
					description = *req.Description
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{"id": sampleID})

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/audit-signings/"):
				pathAfter := strings.TrimPrefix(r.URL.Path, "/configs/v1/audit-signings/")
				isNameLookup := strings.Contains(pathAfter, "wonka-audit") &&
					r.URL.Query().Get("project_id") == appSpaceID
				isIDLookup := strings.Contains(pathAfter, sampleID)

				if isNameLookup || isIDLookup {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(resp)
				} else {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
				}

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

		importedConfig := fmt.Sprintf(tfConfigDef, appSpaceID, "wonka-audit",
			`display_name = "Wonka Audit"
			key_provider = "CUSTOMER_AWS_KMS"
			key_resource = "`+keyResource+`"
			kid          = "`+kid+`"
			auth_params = {
				access_key_id     = "example-access-key-id"
				secret_access_key = "example-secret-access-key"
			}
			`,
		)
		expectedAuthParams := map[string]string{
			"access_key_id":     "example-access-key-id",
			"secret_access_key": "example-secret-access-key",
		}

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": func() (*schema.Provider, error) { return provider, nil },
			},
			Steps: []resource.TestStep{
				{
					// Import by name within the location, and keep the imported state so the
					// following steps start from it, the way a real import does.
					Config:             importedConfig,
					ResourceName:       resourceName,
					ImportState:        true,
					ImportStateId:      "wonka-audit?location=" + appSpaceID,
					ImportStatePersist: true,
					// Imported state can only know the auth_params keys, values come back masked.
					ImportStateCheck: func(states []*terraform.InstanceState) error {
						if len(states) != 1 {
							return fmt.Errorf("expected 1 state, got %d", len(states))
						}
						attrs := states[0].Attributes
						return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, Keys{
							"id":                            Equal(sampleID),
							"name":                          Equal("wonka-audit"),
							"key_provider":                  Equal("CUSTOMER_AWS_KMS"),
							"key_resource":                  Equal(keyResource),
							"kid":                           Equal(kid),
							"auth_params.%":                 Equal("2"),
							"auth_params.access_key_id":     Equal(""),
							"auth_params.secret_access_key": Equal(""),
						}), attrs)
					},
				},
				{
					// First apply after import: the plan replaces the masked values with the
					// configured secrets (the mock rejects masked values), and state keeps them.
					Config: importedConfig,
					Check: resource.ComposeTestCheckFunc(
						testAuditSigningResourceDataExists("CUSTOMER_AWS_KMS", keyResource, kid, expectedAuthParams),
					),
				},
				{
					// An unrelated change must still carry the full auth_params, because the
					// backend replaces the whole config on update.
					Config: strings.Replace(importedConfig, `display_name = "Wonka Audit"`,
						`display_name = "Wonka Audit"
						description  = "added after import"`, 1),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "description", "added after import"),
						testAuditSigningResourceDataExists("CUSTOMER_AWS_KMS", keyResource, kid, expectedAuthParams),
					),
				},
			},
		})
	})
})

func testAuditSigningResourceDataExists(
	expectedProvider, expectedKeyResource, expectedKid string,
	expectedAuthParams map[string]string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[auditSigningResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", auditSigningResourceName)
		}

		if rs.Primary.ID != sampleID {
			return errors.New("ID does not match")
		}

		keys := Keys{
			"id": Equal(sampleID),
			"%":  Not(BeEmpty()),

			"customer_id":  Equal(customerID),
			"app_space_id": Equal(appSpaceID),
			"location":     Equal(appSpaceID),
			"name":         Not(BeEmpty()),
			"key_provider": Equal(expectedProvider),
			"key_resource": Equal(expectedKeyResource),
			"kid":          Equal(expectedKid),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
		}
		// An unset map is absent from the flatmap state entirely, no "%" count is written.
		if len(expectedAuthParams) > 0 {
			keys["auth_params.%"] = Equal(strconv.Itoa(len(expectedAuthParams)))
		}
		for k, v := range expectedAuthParams {
			keys["auth_params."+k] = Equal(v)
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), rs.Primary.Attributes)
	}
}
