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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// authUnavailableBody is the documented 503 body returned when the API cannot verify the token.
const authUnavailableBody = `{"message":"Unable to verify the credential token, retry the request"}`

type scriptedResponse struct {
	body   string
	status int
}

// scriptedServer answers each request with the next scripted response, repeating the last one, and
// records the body of every request it receives.
type scriptedServer struct {
	*httptest.Server

	responses []scriptedResponse
	bodies    []string
	mu        sync.Mutex
}

func newScriptedServer(responses ...scriptedResponse) *scriptedServer {
	s := &scriptedServer{responses: responses}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s.mu.Lock()
		resp := s.responses[min(len(s.bodies), len(s.responses)-1)]
		s.bodies = append(s.bodies, string(body))
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.status)
		_, _ = w.Write([]byte(resp.body))
	}))
	return s
}

func (s *scriptedServer) attempts() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.bodies...)
}

func (s *scriptedServer) client() *indykite.RestClient {
	return indykite.NewTestRestClient(s.URL+"/configs/v1", s.Client())
}

var _ = Describe("REST client retries on 503", func() {
	var server *scriptedServer

	BeforeEach(func() {
		DeferCleanup(indykite.SetUnavailableRetryWaits(time.Millisecond, time.Millisecond))
	})

	AfterEach(func() {
		server.Close()
	})

	unavailable := scriptedResponse{status: http.StatusServiceUnavailable, body: `{"message":"busy"}`}
	authUnavailable := scriptedResponse{status: http.StatusServiceUnavailable, body: authUnavailableBody}
	ok := scriptedResponse{status: http.StatusOK, body: `{"id":"gid:x"}`}

	DescribeTable("idempotent methods retry any 503",
		func(call func(*indykite.RestClient) error) {
			server = newScriptedServer(unavailable, ok)
			Expect(call(server.client())).To(Succeed())
			Expect(server.attempts()).To(HaveLen(2))
		},
		Entry("GET", func(c *indykite.RestClient) error {
			return c.Get(context.Background(), "/projects/gid:x", &map[string]any{})
		}),
		Entry("PUT", func(c *indykite.RestClient) error {
			return c.Put(context.Background(), "/projects/gid:x", map[string]any{"description": "d"}, nil)
		}),
		Entry("DELETE", func(c *indykite.RestClient) error {
			return c.Delete(context.Background(), "/projects/gid:x")
		}),
	)

	It("returns the 503 after exhausting the retries", func() {
		server = newScriptedServer(unavailable)
		err := server.client().Get(context.Background(), "/projects/gid:x", &map[string]any{})
		Expect(indykite.IsServiceError(err)).To(BeTrue(), "error: %v", err)
		Expect(server.attempts()).To(HaveLen(indykite.UnavailableRetryMax() + 1))
	})

	It("does not retry other server errors", func() {
		server = newScriptedServer(scriptedResponse{status: http.StatusInternalServerError, body: `{}`}, ok)
		err := server.client().Get(context.Background(), "/projects/gid:x", &map[string]any{})
		Expect(indykite.IsServiceError(err)).To(BeTrue(), "error: %v", err)
		Expect(server.attempts()).To(HaveLen(1))
	})

	It("retries a POST on the credential verification 503 and resends the same body", func() {
		created := scriptedResponse{status: http.StatusCreated, body: `{"id":"gid:x"}`}
		server = newScriptedServer(authUnavailable, created)
		var resp map[string]any
		Expect(server.client().Post(context.Background(), "/projects", map[string]any{"name": "n"}, &resp)).
			To(Succeed())
		Expect(resp).To(HaveKeyWithValue("id", "gid:x"))
		Expect(server.attempts()).To(Equal([]string{`{"name":"n"}`, `{"name":"n"}`}))
	})

	It("does not retry a POST on any other 503", func() {
		server = newScriptedServer(unavailable, ok)
		err := server.client().Post(context.Background(), "/projects", map[string]any{"name": "n"}, nil)
		Expect(indykite.IsServiceError(err)).To(BeTrue(), "error: %v", err)
		Expect(err).To(MatchError(ContainSubstring("busy")))
		Expect(server.attempts()).To(HaveLen(1))
	})

	It("does not retry a PATCH on any other 503", func() {
		server = newScriptedServer(unavailable, ok)
		_, err := server.client().Do(context.Background(), http.MethodPatch,
			"/projects/gid:x", map[string]any{"description": "d"}, nil)
		Expect(indykite.IsServiceError(err)).To(BeTrue(), "error: %v", err)
		Expect(server.attempts()).To(HaveLen(1))
	})

	It("retries a PATCH on the credential verification 503", func() {
		server = newScriptedServer(authUnavailable, ok)
		_, err := server.client().Do(context.Background(), http.MethodPatch,
			"/projects/gid:x", map[string]any{"description": "d"}, nil)
		Expect(err).To(Succeed())
		Expect(server.attempts()).To(HaveLen(2))
	})

	DescribeTable("PostWithRetryOnNotFound",
		func(first scriptedResponse, expectedAttempts int, expectSuccess bool) {
			server = newScriptedServer(first, scriptedResponse{status: http.StatusCreated, body: `{"id":"gid:x"}`})
			err := server.client().PostWithRetryOnNotFound(context.Background(), "/application-agent-credentials",
				map[string]any{"application_agent_id": "gid:a"}, nil, 3, time.Millisecond, time.Millisecond)
			if expectSuccess {
				Expect(err).To(Succeed())
			} else {
				Expect(indykite.IsServiceError(err)).To(BeTrue(), "error: %v", err)
			}
			Expect(server.attempts()).To(HaveLen(expectedAttempts))
		},
		Entry("retries a 404", scriptedResponse{status: http.StatusNotFound, body: `{}`}, 2, true),
		Entry("retries the credential verification 503", authUnavailable, 2, true),
		Entry("does not retry any other 503", unavailable, 1, false),
	)

	It("caps a long Retry-After at the maximum wait", func() {
		var attempts int
		var mu sync.Mutex
		retryAfterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			mu.Lock()
			attempts++
			first := attempts == 1
			mu.Unlock()
			if first {
				w.Header().Set("Retry-After", "300")
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		DeferCleanup(retryAfterServer.Close)
		server = newScriptedServer(ok) // closed by AfterEach

		client := indykite.NewTestRestClient(retryAfterServer.URL+"/configs/v1", retryAfterServer.Client())
		start := time.Now()
		Expect(client.Get(context.Background(), "/projects/gid:x", &map[string]any{})).To(Succeed())
		Expect(time.Since(start)).To(BeNumerically("<", 5*time.Second))
	})

	It("stops retrying when the context is done", func() {
		DeferCleanup(indykite.SetUnavailableRetryWaits(time.Second, time.Second))
		server = newScriptedServer(unavailable)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := server.client().Get(ctx, "/projects/gid:x", &map[string]any{})
		Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue(), "error: %v", err)
		Expect(time.Since(start)).To(BeNumerically("<", time.Second))
		Expect(server.attempts()).To(HaveLen(1))
	})
})
