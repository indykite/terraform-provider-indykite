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

package indykite

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// RestClient wraps HTTP client for IndyKite Config REST API.
type RestClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewRestClient creates a new REST client for IndyKite Config API.
func NewRestClient(_ context.Context) (*RestClient, error) {
	// Get service account credentials from environment
	credentials := os.Getenv("INDYKITE_SERVICE_ACCOUNT_CREDENTIALS")
	credsFile := os.Getenv("INDYKITE_SERVICE_ACCOUNT_CREDENTIALS_FILE")

	if credentials == "" && credsFile != "" {
		// #nosec G304 -- credsFile is from environment variable, intentional file read
		data, err := os.ReadFile(credsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read credentials file: %w", err)
		}
		credentials = string(data)
	}

	if credentials == "" {
		return nil, errors.New(
			"INDYKITE_SERVICE_ACCOUNT_CREDENTIALS or INDYKITE_SERVICE_ACCOUNT_CREDENTIALS_FILE must be set")
	}

	// Parse credentials to get token and endpoint
	token, baseURL, err := parseCredentials(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &RestClient{
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		baseURL: baseURL,
		token:   token,
	}, nil
}

// parseCredentials extracts token and base URL from service account credentials.
// Returns (token, baseURL, error).
func parseCredentials(credentials string) (string, string, error) { //nolint:revive,gocritic // different concepts
	var creds struct {
		AppSpaceID       string          `json:"appSpaceId"`
		ServiceAccountID string          `json:"serviceAccountId"`
		Endpoint         string          `json:"endpoint"`
		BaseURL          string          `json:"baseUrl"`
		Token            string          `json:"token"`
		PrivateKey       json.RawMessage `json:"privateKeyJWK"`
	}

	if unmarshalErr := json.Unmarshal([]byte(credentials), &creds); unmarshalErr != nil {
		return "", "", fmt.Errorf("failed to parse credentials JSON: %w", unmarshalErr)
	}

	// Determine base URL - use baseUrl from credentials if available, otherwise derive from endpoint
	var baseURL string
	switch {
	case creds.BaseURL != "":
		baseURL = creds.BaseURL
		// Ensure baseURL ends with /configs/v1 if it doesn't already
		if !strings.HasSuffix(baseURL, "/configs/v1") {
			baseURL = strings.TrimSuffix(baseURL, "/") + "/configs/v1"
		}
	case creds.Endpoint != "":
		// Fallback: derive from endpoint for backward compatibility
		if strings.Contains(creds.Endpoint, "us.api.indykite.com") {
			baseURL = "https://us.api.indykite.com/configs/v1"
		} else {
			baseURL = "https://eu.api.indykite.com/configs/v1"
		}
	default:
		// Default to EU if neither baseUrl nor endpoint is provided
		baseURL = "https://eu.api.indykite.com/configs/v1"
	}

	// If token is provided in credentials, use it directly
	if creds.Token != "" {
		return creds.Token, baseURL, nil
	}

	// Otherwise, generate JWT token from private key (backward compatibility)
	privateKey, kid, err := parseJWK(creds.PrivateKey)
	if err != nil {
		return "", baseURL, fmt.Errorf("failed to parse private key JWK: %w", err)
	}

	// Determine the subject for JWT - use serviceAccountId if available, otherwise appSpaceId
	subject := creds.ServiceAccountID
	if subject == "" {
		subject = creds.AppSpaceID
	}

	// Generate JWT token
	token, err := generateJWT(privateKey, kid, subject)
	if err != nil {
		return "", baseURL, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	return token, baseURL, nil
}

// parseJWK parses a JWK (JSON Web Key) using lestrrat-go/jwx library and returns an ECDSA private key and kid.
func parseJWK(jwkData json.RawMessage) (*ecdsa.PrivateKey, string, error) {
	// Parse JWK using jwx library
	key, err := jwk.ParseKey(jwkData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse JWK: %w", err)
	}

	// Get the key ID
	kid := key.KeyID()

	// Convert to raw key (crypto.PrivateKey interface)
	var rawKey any
	if err := key.Raw(&rawKey); err != nil {
		return nil, "", fmt.Errorf("failed to get raw key: %w", err)
	}

	// Assert that it's an ECDSA private key
	privateKey, ok := rawKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, "", fmt.Errorf("key is not an ECDSA private key, got %T", rawKey)
	}

	return privateKey, kid, nil
}

// generateJWT creates a signed JWT token for authentication.
// Claims match the SDK implementation (no audience claim).
func generateJWT(privateKey *ecdsa.PrivateKey, kid, subject string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": subject,
		"sub": subject,
		"iat": now.Unix(),
		"exp": now.Add(1 * time.Hour).Unix(),
		"jti": strconv.FormatInt(now.UnixNano(), 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = kid

	// Sign the token
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}

// Do executes an HTTP request.
func (c *RestClient) Do(ctx context.Context, method, path string, body, response any) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	// Always close the body when we're done
	defer resp.Body.Close() //nolint:errcheck // deferred Body.Close() error is acceptable

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return resp, &RestError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	// Parse response if needed
	if response != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return resp, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return resp, nil
}

// Get executes a GET request.
func (c *RestClient) Get(ctx context.Context, path string, response any) error {
	_, err := c.Do(ctx, http.MethodGet, path, nil, response) //nolint:bodyclose // body is closed in Do()
	return err
}

// Post executes a POST request.
func (c *RestClient) Post(ctx context.Context, path string, body, response any) error {
	_, err := c.Do(ctx, http.MethodPost, path, body, response) //nolint:bodyclose // body is closed in Do()
	return err
}

// Put executes a PUT request.
func (c *RestClient) Put(ctx context.Context, path string, body, response any) error {
	_, err := c.Do(ctx, http.MethodPut, path, body, response) //nolint:bodyclose // body is closed in Do()
	return err
}

// Delete executes a DELETE request.
func (c *RestClient) Delete(ctx context.Context, path string) error {
	_, err := c.Do(ctx, http.MethodDelete, path, nil, nil) //nolint:bodyclose // body is closed in Do()
	return err
}

// RestError represents an error from the REST API.
type RestError struct {
	Message    string
	StatusCode int
}

func (e *RestError) Error() string {
	return "HTTP " + strconv.Itoa(e.StatusCode) + ": " + e.Message
}

// IsNotFoundError checks if the error is a 404 Not Found error.
func IsNotFoundError(err error) bool {
	var restErr *RestError
	if errors.As(err, &restErr) {
		return restErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsServiceError checks if the error is a service error (5xx).
func IsServiceError(err error) bool {
	var restErr *RestError
	if errors.As(err, &restErr) {
		return restErr.StatusCode >= 500
	}
	return false
}

// NewTestRestClient creates a REST client for testing with a custom base URL and HTTP client.
// This function is intended for use in tests only.
func NewTestRestClient(baseURL string, httpClient *http.Client) *RestClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 2 * time.Minute,
		}
	}
	return &RestClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		token:      "test-token", // For testing, we use a dummy token
	}
}
