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
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeConfigAPIActor is reported by fakeConfigAPI as the creator and last updater of the object.
const fakeConfigAPIActor = "gid:AAAAEkF4Ym9vdHN0cmFwAAAA"

// fakeConfigAPI mimics the update behaviour of a Config API configuration object: the
// configuration fields listed in replaced are reset to their zero value when an update omits
// them, display_name and description keep their value when omitted and are cleared by an
// empty string (an empty display_name falls back to the name).
type fakeConfigAPI struct {
	*httptest.Server

	stored    map[string]any
	putBodies []map[string]any
}

// newFakeConfigAPI serves a single configuration object with the given ID under path, e.g.
// "/authorization-policies". replaced maps each configuration field to its zero value.
func newFakeConfigAPI(path, id string, replaced map[string]any) *fakeConfigAPI {
	fake := &fakeConfigAPI{}
	fake.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, path):
			var body map[string]any
			Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
			fake.stored = maps.Clone(body)
			fake.stored["id"] = id
			fake.stored["organization_id"] = customerID
			fake.stored["create_time"] = "2026-10-14T08:00:00Z"
			fake.stored["update_time"] = "2026-10-14T08:00:00Z"
			fake.stored["created_by"] = fakeConfigAPIActor
			fake.stored["updated_by"] = fakeConfigAPIActor
			if v, _ := body["display_name"].(string); v == "" {
				fake.stored["display_name"] = body["name"]
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": id})

		case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, path+"/"+id):
			var body map[string]any
			Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
			fake.putBodies = append(fake.putBodies, body)
			for field, zero := range replaced {
				if v, ok := body[field]; ok && v != nil {
					fake.stored[field] = v
				} else {
					fake.stored[field] = zero
				}
			}
			for _, field := range []string{"display_name", "description"} {
				if v, ok := body[field].(string); ok {
					fake.stored[field] = v
				}
			}
			if v, ok := body["display_name"].(string); ok && v == "" {
				fake.stored["display_name"] = fake.stored["name"]
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": id})

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, id) && fake.stored != nil:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(fake.stored)

		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	return fake
}

// lastPut returns the body of the latest update request, or nil when there was none.
func (f *fakeConfigAPI) lastPut() map[string]any {
	if len(f.putBodies) == 0 {
		return nil
	}
	return f.putBodies[len(f.putBodies)-1]
}
