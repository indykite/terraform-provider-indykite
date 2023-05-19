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
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	configm "github.com/indykite/indykite-sdk-go/test/config/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/indykite/terraform-provider-indykite/indykite"
	"github.com/indykite/terraform-provider-indykite/indykite/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Resource ApplicationAgentCredential", func() {
	const resourceName = "indykite_application_agent_credential.development"
	var (
		mockCtrl                *gomock.Controller
		mockConfigClient        *configm.MockConfigManagementAPIClient
		indykiteProviderFactory func() (*schema.Provider, error)
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(TerraformGomockT(GinkgoT()))
		mockConfigClient = configm.NewMockConfigManagementAPIClient(mockCtrl)

		indykiteProviderFactory = func() (*schema.Provider, error) {
			p := indykite.Provider()
			cfgFunc := p.ConfigureContextFunc
			p.ConfigureContextFunc =
				func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
					client, _ := config.NewTestClient(ctx, mockConfigClient)
					ctx = context.WithValue(ctx, indykite.ClientContext, client)
					return cfgFunc(ctx, data)
				}
			return p, nil
		}
	})

	It("Test all CRUD", func() {
		// Terraform created config must be in sync with returned data in expectedApp and expectedUpdatedApp
		// otherwise "After applying this test step, the plan was not empty" error is thrown
		tfConfigDef :=
			`resource "indykite_application_agent_credential" "development" {
				app_agent_id = "` + appAgentID + `"
				display_name = "%s"
				%s
			}`
		appAgentJWKCredResp := &configpb.ApplicationAgentCredential{
			CustomerId:         customerID,
			AppSpaceId:         appSpaceID,
			ApplicationId:      applicationID,
			ApplicationAgentId: appAgentID,
			Id:                 appAgentCredID,
			Kid:                "EfUEiFnOzA5PCp8SSksp7iXv7cHRehCsIGo6NAQ9H7w",
			CreateTime:         timestamppb.Now(),
		}
		// just a placeholder
		appAgentConfig :=
			`{"defaultTenantId": "%s", "appAgentId": "%s", "privateKeyJWK": {"kty":"EC", "use":"sig", "kid":"..."}}`

		appAgentPEMCredResp := &configpb.ApplicationAgentCredential{
			CustomerId:         customerID,
			AppSpaceId:         appSpaceID,
			ApplicationId:      applicationID,
			ApplicationAgentId: appAgentID,
			Id:                 "kMsC87ROQ8mlK-Q6PSoTuw",
			DisplayName:        "OPA credentials",
			Kid:                "BgQgo-U3kF7kf2dXLKFPNcl3haR8k1VD2nTTvp0GBhI",
			CreateTime:         timestamppb.Now(),
		}

		// Create
		mockConfigClient.EXPECT().
			RegisterApplicationAgentCredential(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"ApplicationAgentId": Equal(appAgentID),
				"DisplayName":        BeEmpty(),
				"DefaultTenantId":    Equal(tenantID),
				"ExpireTime":         Not(BeNil()),
				"PublicKey": PointTo(MatchFields(IgnoreExtras, Fields{
					"Jwk": HaveLen(226),
				})),
			})))).
			Return(&configpb.RegisterApplicationAgentCredentialResponse{
				Id:                 appAgentJWKCredResp.Id,
				ApplicationAgentId: appAgentID,
				Kid:                appAgentJWKCredResp.Kid,
				AgentConfig:        []byte(fmt.Sprintf(appAgentConfig, tenantID, appAgentID)),
			}, nil)

		mockConfigClient.EXPECT().
			RegisterApplicationAgentCredential(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"ApplicationAgentId": Equal(appAgentID),
				"DisplayName":        Equal(appAgentPEMCredResp.DisplayName),
				"DefaultTenantId":    BeEmpty(),
				"ExpireTime":         BeNil(),
				"PublicKey": PointTo(MatchFields(IgnoreExtras, Fields{
					"Pem": HaveLen(271),
				})),
			})))).
			Return(&configpb.RegisterApplicationAgentCredentialResponse{
				Id:                 appAgentPEMCredResp.Id,
				ApplicationAgentId: appAgentID,
				Kid:                appAgentPEMCredResp.Kid,
				AgentConfig:        []byte(fmt.Sprintf(appAgentConfig, "", appAgentID)),
			}, nil)

		// Read in given order
		gomock.InOrder(
			mockConfigClient.EXPECT().
				ReadApplicationAgentCredential(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(appAgentJWKCredResp.Id),
				})))).
				Times(6).
				Return(&configpb.ReadApplicationAgentCredentialResponse{
					ApplicationAgentCredential: appAgentJWKCredResp,
				}, nil),

			mockConfigClient.EXPECT().
				ReadApplicationAgentCredential(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
					"Id": Equal(appAgentPEMCredResp.Id),
				})))).
				Times(2).
				Return(&configpb.ReadApplicationAgentCredentialResponse{
					ApplicationAgentCredential: appAgentPEMCredResp,
				}, nil),
		)

		// Delete
		mockConfigClient.EXPECT().
			DeleteApplicationAgentCredential(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(appAgentJWKCredResp.Id),
			})))).
			Return(&configpb.DeleteApplicationAgentCredentialResponse{}, nil)
		mockConfigClient.EXPECT().
			DeleteApplicationAgentCredential(gomock.Any(), test.WrapMatcher(PointTo(MatchFields(IgnoreExtras, Fields{
				"Id": Equal(appAgentPEMCredResp.Id),
			})))).
			Return(&configpb.DeleteApplicationAgentCredentialResponse{}, nil)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"indykite": indykiteProviderFactory,
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
				{
					Config:      fmt.Sprintf(tfConfigDef, "", `expire_time = "abc"`),
					ExpectError: regexp.MustCompile(`expected "expire_time" to be a valid RFC3339 date`),
				},
				{
					// Checking Create and Read (appAgentJWKCredResp)
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
						default_tenant_id = "`+tenantID+`"
						`),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentCredResourceDataExists(
							resourceName,
							appAgentJWKCredResp,
							Keys{
								"expire_time":       Not(BeEmpty()), // Is string, do not compare with BeTemporarily
								"public_key_jwk":    ContainSubstring("xuyd5-9bT0L09mi810mycfREAxBG3KnpctlGQCYtCdM"),
								"default_tenant_id": Equal(tenantID),
								"agent_config":      MatchJSON(fmt.Sprintf(appAgentConfig, tenantID, appAgentID)),
							},
						),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("not found: %s", resourceName)
							}
							v, has := rs.Primary.Attributes["agent_config"]
							if !has || !strings.Contains(v, tenantID) {
								return fmt.Errorf("failed to find default_tenant id in agent config: %s", v)
							}
							return nil
						},
					),
				},
				{
					// Performs in-place update without calling BE
					// Only change here is default_tenant_id
					// However, tests always do double-check so READ is executed twice here too
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
						default_tenant_id = "`+sampleID+`"
						`),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentCredResourceDataExists(
							resourceName,
							appAgentJWKCredResp,
							Keys{
								"expire_time":       Not(BeEmpty()), // Is string, do not compare with BeTemporarily
								"public_key_jwk":    ContainSubstring("xuyd5-9bT0L09mi810mycfREAxBG3KnpctlGQCYtCdM"),
								"default_tenant_id": Equal(sampleID),
								"agent_config":      MatchJSON(fmt.Sprintf(appAgentConfig, sampleID, appAgentID)),
							},
						),
						func(s *terraform.State) error {
							rs, ok := s.RootModule().Resources[resourceName]
							if !ok {
								return fmt.Errorf("not found: %s", resourceName)
							}
							v, has := rs.Primary.Attributes["agent_config"]
							if !has || strings.Contains(v, tenantID) || !strings.Contains(v, sampleID) {
								return fmt.Errorf(
									"agent config contains old/does not contain new default_tenant_id: %s", v)
							}
							return nil
						},
					),
				},
				{
					// Performs 1 read (appAgentJWKCredResp)
					ResourceName:  resourceName,
					ImportState:   true,
					ImportStateId: appAgentJWKCredResp.Id,
				},
				{
					// Checking Create and Read (appAgentPEMCredResp)
					Config: fmt.Sprintf(tfConfigDef, appAgentPEMCredResp.DisplayName, `public_key_pem = <<-EOT
						-----BEGIN PUBLIC KEY-----
						MIGeMA0GCSqGSIb3DQEBAQUAA4GMADCBiAKBgHRMVhhoOrM0ldxMoaXQ6d9z9aBw
						+BnjNPxKKMeyRYNHZW18CK2Av28AXla0sXca8N30lHcaCV0/DfZ+Kg4UC8aNSDlH
						hEhSGYucKHN+kdf56qmA+odF87gvunkwzJuZddBYAKv9pevZBIn/e3TG8xIfI0S7
						j8ZGOIOYXO64OPXFAgMBAAE=
						-----END PUBLIC KEY-----
						EOT`),
					Check: resource.ComposeTestCheckFunc(
						testAppAgentCredResourceDataExists(
							resourceName,
							appAgentPEMCredResp,
							Keys{
								"public_key_pem": ContainSubstring("-----BEGIN PUBLIC KEY-----"),
								"agent_config":   MatchJSON(fmt.Sprintf(appAgentConfig, "", appAgentID)),
							},
						),
					),
				},
			},
		})
	})
})

func testAppAgentCredResourceDataExists(
	n string,
	data *configpb.ApplicationAgentCredential,
	extraKeys Keys,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID != data.Id {
			return errors.New("ID does not match")
		}

		keys := Keys{
			"id": Equal(data.Id),
			"%":  Not(BeEmpty()), // This is Terraform helper

			"customer_id":    Equal(data.CustomerId),
			"app_space_id":   Equal(data.AppSpaceId),
			"application_id": Equal(data.ApplicationId),
			"app_agent_id":   Equal(data.ApplicationAgentId),
			"display_name":   Equal(data.DisplayName),
			"kid":            Equal(data.Kid),
			"create_time":    Not(BeEmpty()),

			// "expire_time":       Not(BeEmpty()),
			// "default_tenant_id": Not(BeEmpty()),
			// "agent_config":   MatchJSON(agentConfig),
			// "public_key_jwk": ContainSubstring(`"P-256"`), // It is only defined by user, not from response
		}

		for k, v := range extraKeys {
			keys[k] = v
		}

		return convertOmegaMatcherToError(MatchAllKeys(keys), rs.Primary.Attributes)
	}
}
