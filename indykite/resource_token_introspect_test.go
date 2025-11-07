// Copyright (c) 2024 IndyKite
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

var _ = Describe("Resource TokenIntrospect", func() {
	const (
		resourceName = "indykite_token_introspect.development"
	)
	var (
		mockServer      *httptest.Server
		provider        *schema.Provider
		currentResponse string
	)

	BeforeEach(func() {
		provider = indykite.Provider()
		currentResponse = "initial"
	})

	AfterEach(func() {
		if mockServer != nil {
			mockServer.Close()
		}
	})

	It("Test CRUD of Token Introspect configuration", func() {
		tfConfigDef :=
			`resource "indykite_token_introspect" "development" {
				location = "%s"
				name = "%s"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()

		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/token-introspects"):
				resp := indykite.TokenIntrospectResponse{
					ID:          sampleID,
					Name:        "my-first-token-introspect",
					DisplayName: "Display name of Token Introspect configuration",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					JWT: &indykite.TokenIntrospectJWT{
						Issuer:   "https://example.com",
						Audience: "audience-id",
					},
					Online: &indykite.TokenIntrospectOnline{
						CacheTTL: 600,
					},
					ClaimsMapping: map[string]*indykite.TokenIntrospectClaim{
						"email": {Selector: "mail"},
						"name":  {Selector: "full_name"},
					},
					IKGNodeType:   "MyUser",
					PerformUpsert: true,
					CreateTime:    createTime,
					UpdateTime:    updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, sampleID):
				var resp indykite.TokenIntrospectResponse

				switch currentResponse {
				case "initial", "after_create":
					resp = indykite.TokenIntrospectResponse{
						ID:          sampleID,
						Name:        "my-first-token-introspect",
						DisplayName: "Display name of Token Introspect configuration",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						JWT: &indykite.TokenIntrospectJWT{
							Issuer:   "https://example.com",
							Audience: "audience-id",
						},
						Online: &indykite.TokenIntrospectOnline{
							CacheTTL: 600,
						},
						ClaimsMapping: map[string]*indykite.TokenIntrospectClaim{
							"email": {Selector: "mail"},
							"name":  {Selector: "full_name"},
						},
						IKGNodeType:   "MyUser",
						PerformUpsert: true,
						CreateTime:    createTime,
						UpdateTime:    updateTime,
					}
				case "after_update1":
					resp = indykite.TokenIntrospectResponse{
						ID:          sampleID,
						Name:        "my-first-token-introspect",
						Description: "token introspect description",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Opaque: &indykite.TokenIntrospectOpaque{
							Hint: "my.domain.com",
						},
						Online: &indykite.TokenIntrospectOnline{
							UserinfoEndpoint: "https://data.example.com/userinfo",
						},
						IKGNodeType: "MyUser",
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				case "after_update2":
					resp = indykite.TokenIntrospectResponse{
						ID:          sampleID,
						Name:        "my-first-token-introspect",
						Description: "token introspect description",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						JWT: &indykite.TokenIntrospectJWT{
							Issuer:   "https://example.com",
							Audience: "audience-id",
						},
						Offline: &indykite.TokenIntrospectOffline{
							PublicJWKs: []string{
								`{"kid":"abc","use":"sig","alg":"RS256",` +
									`"n":"--nothing-real-just-random-xyqwerasf--","kty":"RSA"}`,
								`{"kid":"jkl","use":"sig","alg":"RS256",` +
									`"n":"--nothing-real-just-random-435asdf43--","kty":"RSA"}`,
							},
						},
						IKGNodeType: "MyUser",
						SubClaim:    &indykite.TokenIntrospectClaim{Selector: "custom_sub"},
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodPut && strings.Contains(r.URL.Path, sampleID):
				// Determine which update we're handling
				var reqBody map[string]any
				_ = json.NewDecoder(r.Body).Decode(&reqBody)

				if reqBody["opaque"] != nil {
					currentResponse = "after_update1"
				} else if reqBody["offline"] != nil {
					currentResponse = "after_update2"
				}

				var resp indykite.TokenIntrospectResponse
				if currentResponse == "after_update1" {
					resp = indykite.TokenIntrospectResponse{
						ID:          sampleID,
						Name:        "my-first-token-introspect",
						Description: "token introspect description",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						Opaque: &indykite.TokenIntrospectOpaque{
							Hint: "my.domain.com",
						},
						Online: &indykite.TokenIntrospectOnline{
							UserinfoEndpoint: "https://data.example.com/userinfo",
						},
						IKGNodeType: "MyUser",
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
					}
				} else {
					resp = indykite.TokenIntrospectResponse{
						ID:          sampleID,
						Name:        "my-first-token-introspect",
						Description: "token introspect description",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						JWT: &indykite.TokenIntrospectJWT{
							Issuer:   "https://example.com",
							Audience: "audience-id",
						},
						Offline: &indykite.TokenIntrospectOffline{
							PublicJWKs: []string{
								`{"kid":"abc","use":"sig","alg":"RS256",` +
									`"n":"--nothing-real-just-random-xyqwerasf--","kty":"RSA"}`,
								`{"kid":"jkl","use":"sig","alg":"RS256",` +
									`"n":"--nothing-real-just-random-435asdf43--","kty":"RSA"}`,
							},
						},
						IKGNodeType: "MyUser",
						SubClaim:    &indykite.TokenIntrospectClaim{Selector: "custom_sub"},
						CreateTime:  createTime,
						UpdateTime:  time.Now(),
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

		testResourceDataExists := func(
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
				attrs := rs.Primary.Attributes

				keys := Keys{
					"id": Equal(expectedID),
					"%":  Not(BeEmpty()),

					"location":     Equal(appSpaceID),
					"customer_id":  Equal(customerID),
					"app_space_id": Equal(appSpaceID),
					"name":         Not(BeEmpty()),
					"create_time":  Not(BeEmpty()),
					"update_time":  Not(BeEmpty()),

					"ikg_node_type": Not(BeEmpty()),
				}

				// Check for offline/online validation
				if attrs["offline_validation.#"] != "" && attrs["offline_validation.#"] != "0" {
					keys["offline_validation.#"] = Equal("1")
					keys["online_validation.#"] = Equal("0")
				} else if attrs["online_validation.#"] != "" && attrs["online_validation.#"] != "0" {
					keys["offline_validation.#"] = Equal("0")
					keys["online_validation.#"] = Equal("1")
				}

				// Check for JWT/Opaque matcher
				if attrs["jwt_matcher.#"] != "" && attrs["jwt_matcher.#"] != "0" {
					keys["jwt_matcher.#"] = Equal("1")
					keys["opaque_matcher.#"] = Equal("0")
				} else if attrs["opaque_matcher.#"] != "" && attrs["opaque_matcher.#"] != "0" {
					keys["opaque_matcher.#"] = Equal("1")
					keys["jwt_matcher.#"] = Equal("0")
				}

				// Check perform_upsert if present
				if attrs["perform_upsert"] != "" {
					keys["perform_upsert"] = Not(BeEmpty())
				}

				return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), attrs)
			}
		}

		validSettings := `
		opaque_matcher {
			hint = "my.domain.com"
		}
		online_validation {
			user_info_endpoint = "https://data.example.com/userinfo"
		}
		ikg_node_type = "MyUser"
		`

		resource.Test(GinkgoT(), resource.TestCase{
			Providers: map[string]*schema.Provider{
				"indykite": provider,
			},
			Steps: []resource.TestStep{
				// Errors case must be always first
				{
					Config:      fmt.Sprintf(tfConfigDef, "ccc", "name", validSettings),
					ExpectError: regexp.MustCompile("Invalid ID value"),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name", `ikg_node_type = "MyUser"`),
					ExpectError: regexp.MustCompile(
						"one of `offline_validation,online_validation` must be(\\s+)specified"),
				},
				{
					Config:      fmt.Sprintf(tfConfigDef, appSpaceID, "name", `ikg_node_type = "MyUser"`),
					ExpectError: regexp.MustCompile("one of `jwt_matcher,opaque_matcher` must be specified"),
				},
				{
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "name",
						`opaque_matcher {
							hint = "my.domain.com"
						}
						online_validation {}
						ikg_node_type = "MyUser"`),
					ExpectError: regexp.MustCompile(
						"`online_validation.0.user_info_endpoint,opaque_matcher` must be specified"),
				},

				{
					// Checking Create and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-token-introspect",
						`display_name = "Display name of Token Introspect configuration"
						jwt_matcher {
							issuer = "https://example.com"
							audience = "audience-id"
						}
						online_validation {
							cache_ttl = 600
						}
						claims_mapping = {
							"email" = "mail",
							"name" = "full_name"
						}
						ikg_node_type = "MyUser"
						perform_upsert = true
						`,
					),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(resourceName, sampleID),
					),
				},
				{
					// Performs 1 read
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: sampleID,
				},
				{
					// Checking Update and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-token-introspect",
						`description = "token introspect description"
						opaque_matcher {
							hint = "my.domain.com"
						}
						online_validation {
							user_info_endpoint = "https://data.example.com/userinfo"
						}
						ikg_node_type = "MyUser"
					`,
					),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(resourceName, sampleID),
					),
				},
				{
					// Checking Update and Read
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "my-first-token-introspect",
						`description = "token introspect description"
						jwt_matcher {
							issuer = "https://example.com"
							audience = "audience-id"
						}
						offline_validation {
							public_jwks = [
								jsonencode({
									"kid": "abc",
									"use": "sig",
									"alg": "RS256",
									"n": "--nothing-real-just-random-xyqwerasf--",
									"kty": "RSA"
								}),
								jsonencode({
									"kid": "jkl",
									"use": "sig",
									"alg": "RS256",
									"n": "--nothing-real-just-random-435asdf43--",
									"kty": "RSA"
								})
							]
						}
						ikg_node_type = "MyUser"
						sub_claim = "custom_sub"
					`,
					),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(resourceName, sampleID),
					),
				},
			},
		})
	})

	It("Test import by name with location", func() {
		tfConfigDef := `resource "indykite_token_introspect" "development" {
				location = "%s"
				name = "%s"
				%s
			}`

		createTime := time.Now()
		updateTime := time.Now()

		//nolint:lll // Long JWK string
		testJWK := `{"kid":"abc123abc123abc123abc123","use":"sig","alg":"RS256","kty":"RSA","n":"sXchB8A4CkaR3DjDdFUXAYVHx_dWx6gT6tEFc6aOaZ5CPAsqJ8lmwczE6xT74j_xoz1PZtKh6q6tRxX4Mk-GB4KFu0YQH1-WN6uEhEcE0fMtsymrZ1OlN3GhxUAt93Q6zpg3ZD9P4ZtXWr7g6OpkqKAv12v2YKvHtJmEznV0OZPZQj5POj1U0EJp0B98VpTmRtT1K2fKjddAlDq3t35u5xPdkL9l9yLeaMLqGw3tNhxG8amj_Mlq3zEy_QwOmR6OKO5mIF0kCshRsfKj8cMyOrUdxzqaZGuv9KYkgA3ulFjHbInCQwFZ8Z5h7bnQZEkSz94Cz3C4X8hGgSyOShLMuWw","e":"AQAB"}`

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/token-introspects"):
				resp := indykite.TokenIntrospectResponse{
					ID:          sampleID,
					Name:        "wonka-introspect",
					DisplayName: "Wonka Introspect",
					CustomerID:  customerID,
					AppSpaceID:  appSpaceID,
					JWT: &indykite.TokenIntrospectJWT{
						Issuer:   "https://example.com",
						Audience: "audience-id",
					},
					Offline: &indykite.TokenIntrospectOffline{
						PublicJWKs: []string{testJWK},
					},
					IKGNodeType: "Person",
					CreateTime:  createTime,
					UpdateTime:  updateTime,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)

			case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/token-introspects/"):
				// Support both ID and name?location=appSpaceID formats
				pathAfterIntrospects := strings.TrimPrefix(r.URL.Path, "/configs/v1/token-introspects/")
				isNameLookup := strings.Contains(pathAfterIntrospects, "wonka-introspect")
				isIDLookup := strings.Contains(pathAfterIntrospects, sampleID)

				if isNameLookup || isIDLookup {
					resp := indykite.TokenIntrospectResponse{
						ID:          sampleID,
						Name:        "wonka-introspect",
						DisplayName: "Wonka Introspect",
						CustomerID:  customerID,
						AppSpaceID:  appSpaceID,
						JWT: &indykite.TokenIntrospectJWT{
							Issuer:   "https://example.com",
							Audience: "audience-id",
						},
						Offline: &indykite.TokenIntrospectOffline{
							PublicJWKs: []string{testJWK},
						},
						IKGNodeType: "Person",
						CreateTime:  createTime,
						UpdateTime:  updateTime,
					}
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
		defer mockServer.Close()

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
					//nolint:lll // Long JWK modulus value in Terraform config
					Config: fmt.Sprintf(tfConfigDef, appSpaceID, "wonka-introspect",
						`display_name = "Wonka Introspect"
							jwt_matcher {
								issuer = "https://example.com"
								audience = "audience-id"
							}
							offline_validation {
								public_jwks = [
									jsonencode({
									  "kid": "abc123abc123abc123abc123",
									  "use": "sig",
									  "alg": "RS256",
									  "kty": "RSA",
									  "n": join("", [
										"sXchB8A4CkaR3DjDdFUXAYVHx_dWx6gT6tEFc6aOaZ5CPAsqJ8lmwczE6xT74j_xoz1PZtKh6q6tRxX4Mk-GB4KFu0YQH1-WN6uEhEcE0fMtsymrZ1OlN3GhxUAt93Q6zpg3ZD9P4ZtXWr7g6OpkqKAv12v2YKvHtJmEznV0",
										"OZPZQj5POj1U0EJp0B98VpTmRtT1K2fKjddAlDq3t35u5xPdkL9l9yLeaMLqGw3tNhxG8amj_Mlq3zEy_QwOmR6OKO5mIF0kCshRsfKj8cMyOrUdxzqaZGuv9KYkgA3ulFjHbInCQwFZ8Z5h7bnQZEkSz94Cz3C4X8hGgSyOShLMuWw"
									  ]),
									  "e": "AQAB"
									})
								]
							}
							ikg_node_type = "Person"
							`,
					),
					Check: resource.ComposeTestCheckFunc(
						testResourceDataExists(resourceName, sampleID),
					),
				},
				{
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: "wonka-introspect?location=" + appSpaceID,
				},
			},
		})
	})
})

func testResourceDataExists(n, expectedID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID != expectedID {
			return errors.New("ID does not match")
		}

		keys := Keys{
			"id": Equal(expectedID),
			"%":  Not(BeEmpty()),

			"customer_id":  Equal(customerID),
			"app_space_id": Equal(appSpaceID),
			"name":         Not(BeEmpty()),
			"create_time":  Not(BeEmpty()),
			"update_time":  Not(BeEmpty()),
		}

		return convertOmegaMatcherToError(MatchKeys(IgnoreExtras, keys), rs.Primary.Attributes)
	}
}
