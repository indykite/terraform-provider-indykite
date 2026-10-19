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

package indykite

import (
	"crypto/ecdsa"
	"encoding/json"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ParsePrivateKeyJWK exposes parseJWK to tests.
func ParsePrivateKeyJWK(jwkData json.RawMessage) (*ecdsa.PrivateKey, string, error) {
	return parseJWK(jwkData)
}

// ParseCredentials exposes parseCredentials to tests.
func ParseCredentials(credentials string) (string, string, error) { //nolint:revive,gocritic // mirrors parseCredentials
	return parseCredentials(credentials)
}

// SignAuthJWT exposes generateJWT to tests.
func SignAuthJWT(privateKey *ecdsa.PrivateKey, kid, subject string) (string, error) {
	return generateJWT(privateKey, kid, subject)
}

// SetCredCreateWaits overrides the application agent credential create initial
// wait and retry backoff bounds for tests and returns a function that restores
// the original values. This keeps tests that exercise the propagation retry fast.
func SetCredCreateWaits(initial, waitMin, waitMax time.Duration) func() {
	origInitial, origMin, origMax := credCreateInitialWait, credCreateRetryMinWait, credCreateRetryMaxWait
	credCreateInitialWait, credCreateRetryMinWait, credCreateRetryMaxWait = initial, waitMin, waitMax
	return func() {
		credCreateInitialWait, credCreateRetryMinWait, credCreateRetryMaxWait = origInitial, origMin, origMax
	}
}

// ReadPathForID exposes buildReadPath to tests for a resource with the given ID.
func ReadPathForID(resourcePath, id string) string {
	data := (&schema.Resource{}).Data(nil)
	data.SetId(id)
	return buildReadPath(resourcePath, data)
}

// CheckAliasMapping exposes validateAliasMapping to tests.
func CheckAliasMapping(value any) ([]string, []error) {
	return validateAliasMapping(value, "alias_mapping")
}

// ConvertServiceAccountConfig exposes serviceAccountConfigString to tests.
func ConvertServiceAccountConfig(raw json.RawMessage) (string, error) {
	return serviceAccountConfigString(raw)
}

// SetUnavailableRetryWaits overrides the 503 retry backoff bounds for tests and returns a function
// that restores the original values.
func SetUnavailableRetryWaits(waitMin, waitMax time.Duration) func() {
	origMin, origMax := unavailableRetryMinWait, unavailableRetryMaxWait
	unavailableRetryMinWait, unavailableRetryMaxWait = waitMin, waitMax
	return func() {
		unavailableRetryMinWait, unavailableRetryMaxWait = origMin, origMax
	}
}

// CheckOptionalTextLength exposes validateOptionalTextLength to tests.
func CheckOptionalTextLength(maxLen int) func(any, string) ([]string, []error) {
	return validateOptionalTextLength(maxLen)
}

// CheckCharLength exposes validateCharLength to tests.
func CheckCharLength(minLen, maxLen int) func(any, string) ([]string, []error) {
	return validateCharLength(minLen, maxLen)
}

// CheckMapValueCharLength exposes validateMapValueCharLength to tests.
func CheckMapValueCharLength(minLen, maxLen int) func(any, cty.Path) diag.Diagnostics {
	return validateMapValueCharLength(minLen, maxLen)
}

// UnavailableRetryMax exposes the 503 retry bound to tests.
func UnavailableRetryMax() int {
	return unavailableRetryMax
}

// CredCreateMaxRetries exposes the credential create retry bound to tests.
func CredCreateMaxRetries() int {
	return credCreateMaxRetries
}

// SetAgentCreateWait overrides the application agent create initial wait for
// tests and returns a function that restores the original value.
func SetAgentCreateWait(d time.Duration) func() {
	orig := agentCreateInitialWait
	agentCreateInitialWait = d
	return func() {
		agentCreateInitialWait = orig
	}
}
