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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v4/jwk"

	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// newECJWK generates a fresh P-256 private key and returns it together with its JWK JSON encoding.
// When kid is non-empty it is set as the "kid" member of the JWK.
func newECJWK(kid string) (*ecdsa.PrivateKey, json.RawMessage) { //nolint:gocritic // key and its JWK encoding
	GinkgoHelper()

	rawKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	Expect(err).To(Succeed())

	key, err := jwk.Import[jwk.Key](rawKey)
	Expect(err).To(Succeed())
	if kid != "" {
		Expect(key.Set(jwk.KeyIDKey, kid)).To(Succeed())
	}

	data, err := json.Marshal(key)
	Expect(err).To(Succeed())

	return rawKey, data
}

var _ = Describe("REST client credentials", func() {
	Describe("ParseJWK", func() {
		It("parses an EC private key JWK with kid", func() {
			rawKey, data := newECJWK("my-kid")

			privateKey, kid, err := indykite.ParsePrivateKeyJWK(data)
			Expect(err).To(Succeed())
			Expect(kid).To(Equal("my-kid"))
			Expect(privateKey).NotTo(BeNil())
			Expect(privateKey.Equal(rawKey)).To(BeTrue())
		})

		It("returns empty kid when JWK has no kid", func() {
			_, data := newECJWK("")

			privateKey, kid, err := indykite.ParsePrivateKeyJWK(data)
			Expect(err).To(Succeed())
			Expect(kid).To(BeEmpty())
			Expect(privateKey).NotTo(BeNil())
		})

		It("fails on invalid JWK JSON", func() {
			_, _, err := indykite.ParsePrivateKeyJWK(json.RawMessage(`{"kty":`))
			Expect(err).To(MatchError(ContainSubstring("failed to parse JWK")))
		})

		It("fails when key is not an ECDSA private key", func() {
			// Symmetric key exports as []byte, not *ecdsa.PrivateKey.
			symmetric, err := jwk.Import[jwk.Key]([]byte("0123456789abcdef0123456789abcdef"))
			Expect(err).To(Succeed())
			data, err := json.Marshal(symmetric)
			Expect(err).To(Succeed())

			_, _, err = indykite.ParsePrivateKeyJWK(data)
			Expect(err).To(MatchError(ContainSubstring("key is not an ECDSA private key, got []uint8")))
		})

		It("fails when JWK holds only the EC public key", func() {
			rawKey, _ := newECJWK("")
			public, err := jwk.Import[jwk.Key](&rawKey.PublicKey)
			Expect(err).To(Succeed())
			data, err := json.Marshal(public)
			Expect(err).To(Succeed())

			_, _, err = indykite.ParsePrivateKeyJWK(data)
			Expect(err).To(MatchError(ContainSubstring("key is not an ECDSA private key, got *ecdsa.PublicKey")))
		})
	})

	Describe("GenerateJWT", func() {
		It("signs an ES256 token with expected claims and kid header", func() {
			rawKey, _ := newECJWK("")

			tokenString, err := indykite.SignAuthJWT(rawKey, "kid-123", "subject-abc")
			Expect(err).To(Succeed())

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (any, error) {
				return &rawKey.PublicKey, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}))
			Expect(err).To(Succeed())
			Expect(token.Valid).To(BeTrue())
			Expect(token.Header).To(HaveKeyWithValue("kid", "kid-123"))
			Expect(claims).To(HaveKeyWithValue("iss", "subject-abc"))
			Expect(claims).To(HaveKeyWithValue("sub", "subject-abc"))
			Expect(claims).To(HaveKey("iat"))
			Expect(claims).To(HaveKey("exp"))
			Expect(claims).To(HaveKey("jti"))
		})
	})

	Describe("ParseCredentials", func() {
		It("uses token and baseUrl directly when present", func() {
			token, baseURL, err := indykite.ParseCredentials(
				`{"baseUrl": "https://example.indykite.com/", "token": "static-token"}`)
			Expect(err).To(Succeed())
			Expect(token).To(Equal("static-token"))
			Expect(baseURL).To(Equal("https://example.indykite.com/configs/v1"))
		})

		It("keeps baseUrl that already ends with /configs/v1", func() {
			_, baseURL, err := indykite.ParseCredentials(
				`{"baseUrl": "https://example.indykite.com/configs/v1", "token": "t"}`)
			Expect(err).To(Succeed())
			Expect(baseURL).To(Equal("https://example.indykite.com/configs/v1"))
		})

		It("derives US base URL from endpoint", func() {
			_, baseURL, err := indykite.ParseCredentials(
				`{"endpoint": "us.api.indykite.com:443", "token": "t"}`)
			Expect(err).To(Succeed())
			Expect(baseURL).To(Equal("https://us.api.indykite.com/configs/v1"))
		})

		It("derives EU base URL from non-US endpoint", func() {
			_, baseURL, err := indykite.ParseCredentials(
				`{"endpoint": "api.indykite.com:443", "token": "t"}`)
			Expect(err).To(Succeed())
			Expect(baseURL).To(Equal("https://eu.api.indykite.com/configs/v1"))
		})

		It("defaults to EU base URL when neither baseUrl nor endpoint is set", func() {
			_, baseURL, err := indykite.ParseCredentials(`{"token": "t"}`)
			Expect(err).To(Succeed())
			Expect(baseURL).To(Equal("https://eu.api.indykite.com/configs/v1"))
		})

		It("generates a JWT from privateKeyJWK using serviceAccountId as subject", func() {
			rawKey, data := newECJWK("cred-kid")
			creds := `{"serviceAccountId": "` + serviceAccountID + `", "appSpaceId": "` + appSpaceID +
				`", "privateKeyJWK": ` + string(data) + `}`

			tokenString, baseURL, err := indykite.ParseCredentials(creds)
			Expect(err).To(Succeed())
			Expect(baseURL).To(Equal("https://eu.api.indykite.com/configs/v1"))

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (any, error) {
				return &rawKey.PublicKey, nil
			})
			Expect(err).To(Succeed())
			Expect(token.Valid).To(BeTrue())
			Expect(token.Header).To(HaveKeyWithValue("kid", "cred-kid"))
			Expect(claims).To(HaveKeyWithValue("sub", serviceAccountID))
		})

		It("falls back to appSpaceId as subject when serviceAccountId is missing", func() {
			rawKey, data := newECJWK("")
			creds := `{"appSpaceId": "` + appSpaceID + `", "privateKeyJWK": ` + string(data) + `}`

			tokenString, _, err := indykite.ParseCredentials(creds)
			Expect(err).To(Succeed())

			claims := jwt.MapClaims{}
			_, err = jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (any, error) {
				return &rawKey.PublicKey, nil
			})
			Expect(err).To(Succeed())
			Expect(claims).To(HaveKeyWithValue("sub", appSpaceID))
		})

		It("fails on invalid credentials JSON", func() {
			_, _, err := indykite.ParseCredentials(`not-json`)
			Expect(err).To(MatchError(ContainSubstring("failed to parse credentials JSON")))
		})

		It("fails when privateKeyJWK is invalid", func() {
			_, _, err := indykite.ParseCredentials(`{"appSpaceId": "x", "privateKeyJWK": {"kty": "EC"}}`)
			Expect(err).To(MatchError(ContainSubstring("failed to parse private key JWK")))
		})
	})
})
