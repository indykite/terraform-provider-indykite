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

import "time"

// SetCredCreateWaits overrides the application agent credential create initial
// wait and retry backoff bounds for tests and returns a function that restores
// the original values. This keeps tests that exercise the propagation retry fast.
func SetCredCreateWaits(initial, waitMin, waitMax time.Duration) func() {
	origInitial, origMin, origMax := credCreateInitialWait, credCreateRetryWaitMin, credCreateRetryWaitMax
	credCreateInitialWait, credCreateRetryWaitMin, credCreateRetryWaitMax = initial, waitMin, waitMax
	return func() {
		credCreateInitialWait, credCreateRetryWaitMin, credCreateRetryWaitMax = origInitial, origMin, origMax
	}
}

// CredCreateMaxRetries exposes the credential create retry bound to tests.
func CredCreateMaxRetries() int {
	return credCreateMaxRetries
}
