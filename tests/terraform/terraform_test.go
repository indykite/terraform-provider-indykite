// Copyright (c) 2023 IndyKite
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

//go:build integration

package terraform_test

import (
	"context"
	"encoding/json"
	"os"

	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	client   *indykite.RestClient
	myResult = make(map[string]string)
)

type Data struct {
	Resources []Resource `json:"resources"`
}

type Resource struct {
	Mode      string     `json:"mode"`
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Instances []Instance `json:"instances"`
}

type Instance struct {
	Attributes Attribute `json:"attributes"`
}

type Attribute struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var _ = Describe("Terraform", func() {
	BeforeEach(func() {
		var err error
		client, err = indykite.NewRestClient(context.Background())
		Expect(err).To(Succeed())

		jsonData, err := os.ReadFile("../provider/terraform.tfstate")
		Expect(err).To(Succeed())
		Expect(jsonData).NotTo(BeNil())

		data := Data{}
		err = json.Unmarshal(jsonData, &data)
		Expect(err).To(Succeed())

		for _, getid := range data.Resources {
			myResult[getid.Name] = getid.Instances[0].Attributes.ID
		}
		Expect(myResult).NotTo(BeNil())
	})

	It("ReadAppSpace", func() {
		var resp indykite.ApplicationSpaceResponse
		err := client.Get(context.Background(), "/projects/"+myResult["appspace"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["appspace"]))
	})

	It("ReadApplication", func() {
		var resp indykite.ApplicationResponse
		err := client.Get(context.Background(), "/applications/"+myResult["application"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["application"]))
	})

	It("ReadAgent", func() {
		var resp indykite.ApplicationAgentResponse
		err := client.Get(context.Background(), "/application-agents/"+myResult["agent"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["agent"]))
	})

	It("ReadCredential", func() {
		var resp indykite.ApplicationAgentCredentialResponse
		err := client.Get(context.Background(), "/application-agent-credentials/"+myResult["with_public"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["with_public"]))
	})

	It("ReadPolicy", func() {
		var resp indykite.AuthorizationPolicyResponse
		err := client.Get(context.Background(), "/authorization-policies/"+myResult["policy_drive_car"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["policy_drive_car"]))
	})

	It("ReadAgentWithAuditPermissions", func() {
		var resp indykite.ApplicationAgentResponse
		err := client.Get(context.Background(), "/application-agents/"+myResult["agent_audit"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["agent_audit"]))
		Expect(resp.APIPermissions).To(ConsistOf("Audit", "ReadAuthZConfigs"))
	})

	// Only one audit signing configuration is allowed per project.
	It("ReadAuditSigning", func() {
		var resp indykite.AuditSigningResponse
		err := client.Get(context.Background(), "/audit-signings/"+myResult["create-audit-signing"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["create-audit-signing"]))
		Expect(resp.AppSpaceID).To(Equal(myResult["appspace"]))
		Expect(resp.Provider).To(Equal("CUSTOMER_AWS_KMS"))
		Expect(resp.KeyResource).To(HaveValue(
			Equal("arn:aws:kms:eu-west-1:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab")))
		Expect(resp.Kid).To(HaveValue(Equal("automation-terraform-aws")))
		// auth_params is write-only: the keys come back, the values are always masked.
		Expect(resp.AuthParams).To(Equal(map[string]string{"role_arn": "", "external_id": ""}))
	})

	// Configurations created with project_id instead of the deprecated location.
	It("ReadPolicyWithProjectID", func() {
		var resp indykite.AuthorizationPolicyResponse
		err := client.Get(context.Background(), "/authorization-policies/"+myResult["policy_project_id"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["policy_project_id"]))
		Expect(resp.AppSpaceID).To(Equal(myResult["appspace"]))
	})

	It("ReadExternalDataResolverWithProjectID", func() {
		var resp indykite.ExternalDataResolverResponse
		err := client.Get(context.Background(), "/external-data-resolvers/"+myResult["resolver_project_id"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["resolver_project_id"]))
		Expect(resp.AppSpaceID).To(Equal(myResult["appspace"]))
	})

	It("ReadKnowledgeQueryWithProjectID", func() {
		var resp indykite.KnowledgeQueryResponse
		err := client.Get(context.Background(), "/knowledge-queries/"+myResult["query_project_id"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["query_project_id"]))
		Expect(resp.AppSpaceID).To(Equal(myResult["appspace"]))
	})

	It("ReadTrustScoreProfileWithProjectID", func() {
		var resp indykite.TrustScoreProfileResponse
		err := client.Get(context.Background(), "/trust-score-profiles/"+myResult["score_project_id"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["score_project_id"]))
		Expect(resp.AppSpaceID).To(Equal(os.Getenv("TF_VAR_LOCATION_ID")))
	})

	It("ReadEntityMatchingPipelineWithProjectID", func() {
		var resp indykite.EntityMatchingPipelineResponse
		err := client.Get(context.Background(), "/entity-matching-pipelines/"+myResult["pipeline_project_id"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["pipeline_project_id"]))
		Expect(resp.AppSpaceID).To(Equal(os.Getenv("TF_VAR_LOCATION_ID")))
	})

	It("ReadTokenIntrospectWithProjectID", func() {
		var resp indykite.TokenIntrospectResponse
		err := client.Get(context.Background(), "/token-introspects/"+myResult["token_introspect_project_id"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["token_introspect_project_id"]))
		Expect(resp.AppSpaceID).To(Equal(myResult["appspace"]))
	})

	// The appspace is created without ikg_schema_version, so the API applies the default v1.
	It("ReadAppSpaceIKGSchemaVersion", func() {
		var resp indykite.ApplicationSpaceResponse
		err := client.Get(context.Background(), "/projects/"+myResult["appspace"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["appspace"]))
		Expect(resp.IKGSchemaVersion).To(Equal("v1"))
	})

	// Token introspect validation modes other than jwt_matcher with online_validation.
	It("ReadTokenIntrospectWithOAuth2Validation", func() {
		var resp indykite.TokenIntrospectResponse
		err := client.Get(context.Background(), "/token-introspects/"+myResult["token_introspect_oauth2"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["token_introspect_oauth2"]))
		Expect(resp.JWT).To(BeNil())
		Expect(resp.Opaque).To(HaveValue(Equal(indykite.TokenIntrospectOpaque{Hint: "automation-terraform-oauth2"})))
		Expect(resp.Offline).To(BeNil())
		Expect(resp.Online).To(BeNil())
		Expect(resp.OnlineOAuth2).NotTo(BeNil())
		Expect(resp.OnlineOAuth2.IntrospectionEndpoint).To(
			Equal("https://auth-oauth2.example.com/oauth2/introspect"))
		Expect(resp.OnlineOAuth2.CacheTTL).To(Equal(300))
		Expect(resp.OnlineOAuth2.SecretBasic).NotTo(BeNil())
		Expect(resp.OnlineOAuth2.SecretBasic.ClientID).To(Equal("automation-terraform-client"))
		// client_secret is write-only: the API must never return it.
		Expect(resp.OnlineOAuth2.SecretBasic.ClientSecret).To(BeEmpty())
	})

	It("ReadTokenIntrospectWithOfflineJWKs", func() {
		var resp indykite.TokenIntrospectResponse
		err := client.Get(context.Background(), "/token-introspects/"+myResult["token_introspect_offline_jwks"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["token_introspect_offline_jwks"]))
		Expect(resp.JWT).To(HaveValue(Equal(indykite.TokenIntrospectJWT{
			Issuer:   "https://auth-offline.example.com",
			Audience: "automation-terraform-offline-jwks",
		})))
		Expect(resp.Offline).NotTo(BeNil())
		Expect(resp.Offline.PublicJWKs).To(HaveLen(1))
		// The modulus is the RSA example key of RFC 7515, split to keep the lines short.
		modulus := "sXchB8A4CkaR3DjDdFUXAYVHx_dWx6gT6tEFc6aOaZ5CPAsqJ8lmwczE6xT74j_xoz1PZtKh6q6tRxX4Mk-GB4KFu0YQH1" +
			"-WN6uEhEcE0fMtsymrZ1OlN3GhxUAt93Q6zpg3ZD9P4ZtXWr7g6OpkqKAv12v2YKvHtJmEznV0OZPZQj5POj1U0EJp0B98VpTmRtT1K2" +
			"fKjddAlDq3t35u5xPdkL9l9yLeaMLqGw3tNhxG8amj_Mlq3zEy_QwOmR6OKO5mIF0kCshRsfKj8cMyOrUdxzqaZGuv9KYkgA3ulFjH" +
			"bInCQwFZ8Z5h7bnQZEkSz94Cz3C4X8hGgSyOShLMuWw"
		Expect(resp.Offline.PublicJWKs[0]).To(MatchJSON(`{
			"kid": "automation-terraform-key-1",
			"use": "sig",
			"alg": "RS256",
			"kty": "RSA",
			"n": "` + modulus + `",
			"e": "AQAB"
		}`))
		Expect(resp.SubClaim).To(HaveValue(Equal(indykite.TokenIntrospectClaim{Selector: "user_id"})))
	})

	It("ReadTokenIntrospectWithRFC8414Discovery", func() {
		var resp indykite.TokenIntrospectResponse
		err := client.Get(context.Background(), "/token-introspects/"+myResult["token_introspect_rfc8414"], &resp)
		Expect(err).To(Succeed())
		Expect(resp.ID).To(Equal(myResult["token_introspect_rfc8414"]))
		Expect(resp.Offline).NotTo(BeNil())
		Expect(resp.Offline.PublicJWKs).To(BeEmpty())
		Expect(resp.Offline.Discovery).To(Equal("rfc8414"))
	})
})
