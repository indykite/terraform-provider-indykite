// Copyright (c) 2025 IndyKite
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
	"time"
)

// Common structures

// BaseResponse contains common fields in all responses.
type BaseResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Description string    `json:"description,omitempty"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	Etag        string    `json:"etag,omitempty"`
}

// Application structures

// CreateApplicationRequest represents the request to create an application.
type CreateApplicationRequest struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ApplicationResponse represents an application resource.
type ApplicationResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Description string    `json:"description,omitempty"`
	CustomerID  string    `json:"organization_id"`
	AppSpaceID  string    `json:"project_id"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	Etag        string    `json:"etag,omitempty"`
}

// UpdateApplicationRequest represents the request to update an application.
type UpdateApplicationRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ListApplicationsResponse represents the response from listing applications.
type ListApplicationsResponse struct {
	Applications []ApplicationResponse `json:"applications"`
}

// Application Space structures

// DBConnection represents database connection information.
type DBConnection struct {
	URL      string `json:"url,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Name     string `json:"name,omitempty"`
}

// CreateApplicationSpaceRequest represents the request to create an application space.
type CreateApplicationSpaceRequest struct {
	DBConnection  *DBConnection `json:"db_connection,omitempty"`
	CustomerID    string        `json:"organization_id"`
	Name          string        `json:"name"`
	DisplayName   string        `json:"display_name,omitempty"`
	Description   string        `json:"description,omitempty"`
	Region        string        `json:"region,omitempty"`
	IKGSize       string        `json:"ikg_size,omitempty"`
	ReplicaRegion string        `json:"replica_region,omitempty"`
}

// ApplicationSpaceResponse represents an application space resource.
type ApplicationSpaceResponse struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	DisplayName   string        `json:"display_name,omitempty"`
	Description   string        `json:"description,omitempty"`
	CustomerID    string        `json:"organization_id"`
	Region        string        `json:"region,omitempty"`
	IKGSize       string        `json:"ikg_size,omitempty"`
	ReplicaRegion string        `json:"replica_region,omitempty"`
	IKGStatus     string        `json:"ikg_status,omitempty"`
	DBConnection  *DBConnection `json:"db_connection,omitempty"`
	CreateTime    time.Time     `json:"create_time"`
	UpdateTime    time.Time     `json:"update_time"`
	Etag          string        `json:"etag,omitempty"`
}

// UpdateApplicationSpaceRequest represents the request to update an application space.
type UpdateApplicationSpaceRequest struct {
	DisplayName  *string       `json:"display_name,omitempty"`
	Description  *string       `json:"description,omitempty"`
	DBConnection *DBConnection `json:"db_connection,omitempty"`
}

// ListApplicationSpacesResponse represents the response from listing application spaces.
type ListApplicationSpacesResponse struct {
	AppSpaces []ApplicationSpaceResponse `json:"appSpaces"`
}

// Application Agent structures

// CreateApplicationAgentRequest represents the request to create an application agent.
type CreateApplicationAgentRequest struct {
	ApplicationID  string   `json:"application_id"`
	Name           string   `json:"name"`
	DisplayName    string   `json:"display_name,omitempty"`
	Description    string   `json:"description,omitempty"`
	APIPermissions []string `json:"api_permissions,omitempty"`
}

// ApplicationAgentResponse represents an application agent resource.
type ApplicationAgentResponse struct {
	CreateTime     time.Time `json:"create_time"`
	UpdateTime     time.Time `json:"update_time"`
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	DisplayName    string    `json:"display_name,omitempty"`
	Description    string    `json:"description,omitempty"`
	CustomerID     string    `json:"organization_id"`
	AppSpaceID     string    `json:"project_id"`
	ApplicationID  string    `json:"application_id"`
	Etag           string    `json:"etag,omitempty"`
	APIPermissions []string  `json:"api_permissions,omitempty"`
}

// UpdateApplicationAgentRequest represents the request to update an application agent.
type UpdateApplicationAgentRequest struct {
	DisplayName    *string  `json:"display_name,omitempty"`
	Description    *string  `json:"description,omitempty"`
	APIPermissions []string `json:"api_permissions,omitempty"`
}

// ListApplicationAgentsResponse represents the response from listing application agents.
type ListApplicationAgentsResponse struct {
	Agents []ApplicationAgentResponse `json:"agents"`
}

// Application Agent Credential structures

// CreateApplicationAgentCredentialRequest represents the request to create an application agent credential.
type CreateApplicationAgentCredentialRequest struct {
	ApplicationAgentID string `json:"application_agent_id"`
	DisplayName        string `json:"display_name,omitempty"`
	ExpireTime         string `json:"expire_time,omitempty"`
	PublicKeyPEM       string `json:"public_key_pem,omitempty"`
	PublicKeyJWK       string `json:"public_key_jwk,omitempty"`
	DefaultTenantID    string `json:"default_tenant_id,omitempty"` // For backward compatibility with SDK
}

// ApplicationAgentCredentialResponse represents an application agent credential resource.
type ApplicationAgentCredentialResponse struct {
	ID                 string    `json:"id"`
	Kid                string    `json:"kid"`
	DisplayName        string    `json:"display_name,omitempty"`
	CustomerID         string    `json:"customer_id"`
	AppSpaceID         string    `json:"app_space_id"`
	ApplicationID      string    `json:"application_id"`
	ApplicationAgentID string    `json:"application_agent_id"`
	CreateTime         time.Time `json:"create_time"`
	ExpireTime         time.Time `json:"expire_time,omitempty"`
	AgentConfig        string    `json:"agent_config,omitempty"`
	DefaultTenantID    string    `json:"default_tenant_id,omitempty"`
	CreateBy           string    `json:"create_by,omitempty"` // For backward compatibility with SDK
}

// Authorization Policy structures

// CreateAuthorizationPolicyRequest represents the request to create an authorization policy.
type CreateAuthorizationPolicyRequest struct {
	ProjectID   string   `json:"project_id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Policy      string   `json:"policy"`
	Status      string   `json:"status"`
	Tags        []string `json:"tags,omitempty"`
}

// AuthorizationPolicyResponse represents an authorization policy resource.
type AuthorizationPolicyResponse struct {
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Description string    `json:"description,omitempty"`
	CustomerID  string    `json:"organization_id"`
	AppSpaceID  string    `json:"app_space_id,omitempty"`
	Policy      string    `json:"policy"`
	Status      string    `json:"status"`
	Etag        string    `json:"etag,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}

// UpdateAuthorizationPolicyRequest represents the request to update an authorization policy.
type UpdateAuthorizationPolicyRequest struct {
	DisplayName *string  `json:"display_name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Policy      *string  `json:"policy,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ListAuthorizationPoliciesResponse represents the response from listing authorization policies.
type ListAuthorizationPoliciesResponse struct {
	Policies []AuthorizationPolicyResponse `json:"policies"`
}

// Token Introspect structures

// TokenIntrospectJWT represents JWT token matcher configuration.
type TokenIntrospectJWT struct {
	Issuer   string `json:"issuer"`
	Audience string `json:"audience"`
}

// TokenIntrospectOpaque represents opaque token matcher configuration.
type TokenIntrospectOpaque struct {
	Hint string `json:"hint"`
}

// TokenIntrospectOffline represents offline validation configuration.
type TokenIntrospectOffline struct {
	PublicJWKs []string `json:"public_jwks,omitempty"`
}

// TokenIntrospectOnline represents online validation configuration.
type TokenIntrospectOnline struct {
	UserinfoEndpoint string `json:"userinfo_endpoint,omitempty"`
	CacheTTL         int    `json:"cache_ttl,omitempty"` // In seconds
}

// TokenIntrospectClaim represents a claim mapping configuration.
type TokenIntrospectClaim struct {
	Selector string `json:"selector"`
}

// CreateTokenIntrospectRequest represents the request to create a token introspect configuration.
type CreateTokenIntrospectRequest struct {
	ProjectID     string                           `json:"project_id"`
	Name          string                           `json:"name"`
	DisplayName   string                           `json:"display_name,omitempty"`
	Description   string                           `json:"description,omitempty"`
	JWT           *TokenIntrospectJWT              `json:"jwt,omitempty"`
	Opaque        *TokenIntrospectOpaque           `json:"opaque,omitempty"`
	Offline       *TokenIntrospectOffline          `json:"offline,omitempty"`
	Online        *TokenIntrospectOnline           `json:"online,omitempty"`
	ClaimsMapping map[string]*TokenIntrospectClaim `json:"claims_mapping,omitempty"`
	SubClaim      *TokenIntrospectClaim            `json:"sub_claim,omitempty"`
	IKGNodeType   string                           `json:"ikg_node_type"`
	PerformUpsert bool                             `json:"perform_upsert"`
}

// TokenIntrospectResponse represents a token introspect configuration resource.
type TokenIntrospectResponse struct {
	ClaimsMapping map[string]*TokenIntrospectClaim `json:"claims_mapping,omitempty"`
	JWT           *TokenIntrospectJWT              `json:"jwt,omitempty"`
	Opaque        *TokenIntrospectOpaque           `json:"opaque,omitempty"`
	Offline       *TokenIntrospectOffline          `json:"offline,omitempty"`
	Online        *TokenIntrospectOnline           `json:"online,omitempty"`
	SubClaim      *TokenIntrospectClaim            `json:"sub_claim,omitempty"`
	CreateTime    time.Time                        `json:"create_time"`
	UpdateTime    time.Time                        `json:"update_time"`
	CreatedBy     string                           `json:"created_by"`
	UpdatedBy     string                           `json:"update_by"`
	ID            string                           `json:"id"`
	Name          string                           `json:"name"`
	DisplayName   string                           `json:"display_name,omitempty"`
	Description   string                           `json:"description,omitempty"`
	CustomerID    string                           `json:"organization_id"`
	AppSpaceID    string                           `json:"project_id,omitempty"`
	IKGNodeType   string                           `json:"ikg_node_type"`
	Etag          string                           `json:"etag,omitempty"`
	PerformUpsert bool                             `json:"perform_upsert"`
}

// UpdateTokenIntrospectRequest represents the request to update a token introspect configuration.
type UpdateTokenIntrospectRequest struct {
	DisplayName   *string                          `json:"display_name,omitempty"`
	Description   *string                          `json:"description,omitempty"`
	JWT           *TokenIntrospectJWT              `json:"jwt,omitempty"`
	Opaque        *TokenIntrospectOpaque           `json:"opaque,omitempty"`
	Offline       *TokenIntrospectOffline          `json:"offline,omitempty"`
	Online        *TokenIntrospectOnline           `json:"online,omitempty"`
	ClaimsMapping map[string]*TokenIntrospectClaim `json:"claims_mapping,omitempty"`
	SubClaim      *TokenIntrospectClaim            `json:"sub_claim,omitempty"`
	IKGNodeType   *string                          `json:"ikg_node_type,omitempty"`
	PerformUpsert *bool                            `json:"perform_upsert,omitempty"`
}

// Ingest Pipeline structures

// CreateIngestPipelineRequest represents the request to create an ingest pipeline.
type CreateIngestPipelineRequest struct {
	ProjectID     string   `json:"project_id"`
	Name          string   `json:"name"`
	DisplayName   string   `json:"display_name,omitempty"`
	Description   string   `json:"description,omitempty"`
	AppAgentToken string   `json:"app_agent_token"`
	Sources       []string `json:"sources"`
}

// IngestPipelineResponse represents an ingest pipeline resource.
type IngestPipelineResponse struct {
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Description string    `json:"description,omitempty"`
	CustomerID  string    `json:"organization_id"`
	AppSpaceID  string    `json:"project_id,omitempty"`
	Etag        string    `json:"etag,omitempty"`
	Sources     []string  `json:"sources"`
}

// UpdateIngestPipelineRequest represents the request to update an ingest pipeline.
type UpdateIngestPipelineRequest struct {
	DisplayName   *string  `json:"display_name,omitempty"`
	Description   *string  `json:"description,omitempty"`
	AppAgentToken *string  `json:"app_agent_token,omitempty"`
	Sources       []string `json:"sources,omitempty"`
}

// External Data Resolver structures

// ExternalDataResolverHeader represents a header in external data resolver configuration.
type ExternalDataResolverHeader struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// CreateExternalDataResolverRequest represents the request to create an external data resolver.
type CreateExternalDataResolverRequest struct {
	ProjectID        string         `json:"project_id"`
	Name             string         `json:"name"`
	DisplayName      string         `json:"display_name,omitempty"`
	Description      string         `json:"description,omitempty"`
	URL              string         `json:"url"`
	Method           string         `json:"method"`
	Headers          map[string]any `json:"headers,omitempty"`
	RequestType      string         `json:"request_content_type"`
	RequestPayload   string         `json:"request_payload,omitempty"`
	ResponseType     string         `json:"response_content_type"`
	ResponseSelector string         `json:"response_selector"`
}

// ExternalDataResolverResponse represents an external data resolver resource.
type ExternalDataResolverResponse struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	DisplayName      string         `json:"display_name,omitempty"`
	Description      string         `json:"description,omitempty"`
	CustomerID       string         `json:"organization_id"`
	AppSpaceID       string         `json:"project_id,omitempty"`
	URL              string         `json:"url"`
	Method           string         `json:"method"`
	Headers          map[string]any `json:"headers,omitempty"`
	RequestType      string         `json:"request_content_type"`
	RequestPayload   string         `json:"request_payload,omitempty"`
	ResponseType     string         `json:"response_content_type"`
	ResponseSelector string         `json:"response_selector"`
	CreateTime       time.Time      `json:"create_time"`
	UpdateTime       time.Time      `json:"update_time"`
	Etag             string         `json:"etag,omitempty"`
}

// UpdateExternalDataResolverRequest represents the request to update an external data resolver.
type UpdateExternalDataResolverRequest struct {
	DisplayName      *string        `json:"display_name,omitempty"`
	Description      *string        `json:"description,omitempty"`
	URL              *string        `json:"url,omitempty"`
	Method           *string        `json:"method,omitempty"`
	Headers          map[string]any `json:"headers,omitempty"`
	RequestType      *string        `json:"request_content_type,omitempty"`
	RequestPayload   *string        `json:"request_payload,omitempty"`
	ResponseType     *string        `json:"response_content_type,omitempty"`
	ResponseSelector *string        `json:"response_selector,omitempty"`
}

// Entity Matching Pipeline structures

// EntityMatchingNodeFilter represents node filter configuration.
type EntityMatchingNodeFilter struct {
	SourceNodeTypes []string `json:"source_node_types"`
	TargetNodeTypes []string `json:"target_node_types"`
}

// CreateEntityMatchingPipelineRequest represents the request to create an entity matching pipeline.
type CreateEntityMatchingPipelineRequest struct {
	NodeFilter            *EntityMatchingNodeFilter `json:"node_filter"`
	ProjectID             string                    `json:"project_id"`
	Name                  string                    `json:"name"`
	DisplayName           string                    `json:"display_name,omitempty"`
	Description           string                    `json:"description,omitempty"`
	RerunInterval         string                    `json:"rerun_interval,omitempty"`
	SimilarityScoreCutoff float32                   `json:"similarity_score_cutoff"`
}

// EntityMatchingPipelineResponse represents an entity matching pipeline resource.
type EntityMatchingPipelineResponse struct {
	CreateTime            time.Time                 `json:"create_time"`
	UpdateTime            time.Time                 `json:"update_time"`
	NodeFilter            *EntityMatchingNodeFilter `json:"node_filter"`
	ID                    string                    `json:"id"`
	Name                  string                    `json:"name"`
	DisplayName           string                    `json:"display_name,omitempty"`
	Description           string                    `json:"description,omitempty"`
	CustomerID            string                    `json:"organization_id"`
	AppSpaceID            string                    `json:"project_id,omitempty"`
	RerunInterval         string                    `json:"rerun_interval,omitempty"`
	Etag                  string                    `json:"etag,omitempty"`
	SimilarityScoreCutoff float32                   `json:"similarity_score_cutoff,omitempty"`
}

// UpdateEntityMatchingPipelineRequest represents the request to update an entity matching pipeline.
type UpdateEntityMatchingPipelineRequest struct {
	DisplayName           *string  `json:"display_name,omitempty"`
	Description           *string  `json:"description,omitempty"`
	SimilarityScoreCutoff *float32 `json:"similarity_score_cutoff,omitempty"`
	RerunInterval         *string  `json:"rerun_interval,omitempty"`
}

// Knowledge Query structures

// CreateKnowledgeQueryRequest represents the request to create a knowledge query.
type CreateKnowledgeQueryRequest struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
	Query       string `json:"query"`
	Status      string `json:"status"`
	PolicyID    string `json:"policy_id"`
}

// KnowledgeQueryResponse represents a knowledge query resource.
type KnowledgeQueryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Description string    `json:"description,omitempty"`
	CustomerID  string    `json:"organization_id"`
	AppSpaceID  string    `json:"project_id,omitempty"`
	Query       string    `json:"query"`
	Status      string    `json:"status"`
	PolicyID    string    `json:"policy_id"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	Etag        string    `json:"etag,omitempty"`
}

// UpdateKnowledgeQueryRequest represents the request to update a knowledge query.
type UpdateKnowledgeQueryRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
	Query       *string `json:"query,omitempty"`
	Status      *string `json:"status,omitempty"`
	PolicyID    *string `json:"policy_id,omitempty"`
}

// Trust Score Profile structures

// TrustScoreDimension represents a trust score dimension configuration.
type TrustScoreDimension struct {
	Name   string  `json:"name"`
	Weight float32 `json:"weight"`
}

// CreateTrustScoreProfileRequest represents the request to create a trust score profile.
type CreateTrustScoreProfileRequest struct {
	ProjectID          string                 `json:"project_id"`
	Name               string                 `json:"name"`
	DisplayName        string                 `json:"display_name,omitempty"`
	Description        string                 `json:"description,omitempty"`
	NodeClassification string                 `json:"node_classification"`
	Schedule           string                 `json:"schedule"`
	Dimensions         []*TrustScoreDimension `json:"dimensions"`
}

// TrustScoreProfileResponse represents a trust score profile resource.
type TrustScoreProfileResponse struct {
	CreateTime         time.Time              `json:"create_time"`
	UpdateTime         time.Time              `json:"update_time"`
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	DisplayName        string                 `json:"display_name,omitempty"`
	Description        string                 `json:"description,omitempty"`
	CustomerID         string                 `json:"organization_id"`
	AppSpaceID         string                 `json:"project_id,omitempty"`
	NodeClassification string                 `json:"node_classification"`
	Schedule           string                 `json:"schedule"`
	Etag               string                 `json:"etag,omitempty"`
	Dimensions         []*TrustScoreDimension `json:"dimensions"`
}

// UpdateTrustScoreProfileRequest represents the request to update a trust score profile.
type UpdateTrustScoreProfileRequest struct {
	DisplayName *string                `json:"display_name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Schedule    *string                `json:"schedule,omitempty"`
	Dimensions  []*TrustScoreDimension `json:"dimensions,omitempty"`
}

// Config Node structures (generic for multiple resource types)

// CreateConfigNodeRequest represents the request to create a config node.
type CreateConfigNodeRequest struct {
	Config      any    `json:"config"`
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ConfigNodeResponse represents a config node resource.
type ConfigNodeResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Description string    `json:"description,omitempty"`
	CustomerID  string    `json:"organization_id"`
	AppSpaceID  string    `json:"project_id,omitempty"`
	Config      any       `json:"config"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	Etag        string    `json:"etag,omitempty"`
}

// UpdateConfigNodeRequest represents the request to update a config node.
type UpdateConfigNodeRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
	Config      any     `json:"config,omitempty"`
}

// Customer structures

// CustomerResponse represents a customer resource.
type CustomerResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name,omitempty"`
	Description string    `json:"description,omitempty"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
	Etag        string    `json:"etag,omitempty"`
}

// ListCustomersResponse represents the response from listing customers.
type ListCustomersResponse struct {
	Customers []CustomerResponse `json:"organizations"`
}

// Event Sink structures

// CreateEventSinkRequest represents the request to create an event sink.
type CreateEventSinkRequest struct {
	ProjectID   string         `json:"project_id"`
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name,omitempty"`
	Description string         `json:"description,omitempty"`
	Providers   map[string]any `json:"providers"`
	Routes      []any          `json:"routes"`
}

// EventSinkResponse represents an event sink resource.
type EventSinkResponse struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name,omitempty"`
	Description string         `json:"description,omitempty"`
	CustomerID  string         `json:"organization_id"`
	AppSpaceID  string         `json:"project_id,omitempty"`
	Config      map[string]any `json:"config"`
	CreateTime  time.Time      `json:"create_time"`
	UpdateTime  time.Time      `json:"update_time"`
	Etag        string         `json:"etag,omitempty"`
}

// UpdateEventSinkRequest represents the request to update an event sink.
type UpdateEventSinkRequest struct {
	DisplayName *string        `json:"display_name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Config      map[string]any `json:"config,omitempty"`
}

// Service Account structures

// CreateServiceAccountRequest represents the request to create a service account.
type CreateServiceAccountRequest struct {
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	DisplayName    string `json:"display_name,omitempty"`
	Description    string `json:"description,omitempty"`
	Role           string `json:"role"`
}

// ServiceAccountResponse represents a service account resource.
type ServiceAccountResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	DisplayName    string    `json:"display_name,omitempty"`
	Description    string    `json:"description,omitempty"`
	OrganizationID string    `json:"organization_id"`
	Role           string    `json:"role"`
	CreateTime     time.Time `json:"create_time"`
	UpdateTime     time.Time `json:"update_time"`
	Etag           string    `json:"etag,omitempty"`
}

// UpdateServiceAccountRequest represents the request to update a service account.
type UpdateServiceAccountRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Service Account Credential structures

// CreateServiceAccountCredentialRequest represents the request to create a service account credential.
type CreateServiceAccountCredentialRequest struct {
	ServiceAccountID string `json:"service_account_id"`
	DisplayName      string `json:"display_name,omitempty"`
	ExpireTime       string `json:"expire_time,omitempty"`
}

// ServiceAccountCredentialResponse represents a service account credential resource.
type ServiceAccountCredentialResponse struct {
	ID                   string    `json:"id"`
	DisplayName          string    `json:"display_name,omitempty"`
	Kid                  string    `json:"kid"`
	ServiceAccountID     string    `json:"service_account_id"`
	OrganizationID       string    `json:"organization_id"`
	ServiceAccountConfig string    `json:"service_account_config,omitempty"`
	CreateTime           time.Time `json:"create_time"`
	ExpireTime           string    `json:"expire_time,omitempty"`
}
