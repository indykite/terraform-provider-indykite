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

package indykite

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

//nolint:gosec // there are no secrets
const (
	tokenIntrospectJWTKey      = "jwt_matcher"
	tokenIntrospectIssuerKey   = "issuer"
	tokenIntrospectAudienceKey = "audience"
	tokenIntrospectOpaqueKey   = "opaque_matcher"
	tokenIntrospectHintKey     = "hint"

	tokenIntrospectOfflineKey    = "offline_validation"
	tokenIntrospectPublicJWKsKey = "public_jwks"
	tokenIntrospectOnlineKey     = "online_validation"
	tokenIntrospectUserInfoEPKey = "user_info_endpoint"
	tokenIntrospectCacheTTLKey   = "cache_ttl"

	tokenIntrospectClaimsMappingKey  = "claims_mapping"
	tokenIntrospectSubClaimKey       = "sub_claim"
	tokenIntrospectCMPropertyNameKey = "ikg_name"
	tokenIntrospectCMSelectorKey     = "selector"

	tokenIntrospectIKGNodeTypeKey   = "ikg_node_type"
	tokenIntrospectPerformUpsertKey = "perform_upsert"
)

var (
	tokenIntrospectIkgNodeTypeRegex = regexp.MustCompile(`^([A-Z][a-z]+)+$`)
	tokenIntrospectIkgPropertyRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]+$`)
)

func resourceTokenIntrospect() *schema.Resource {
	matcherOneOf := []string{tokenIntrospectJWTKey, tokenIntrospectOpaqueKey}
	validationOneOf := []string{tokenIntrospectOfflineKey, tokenIntrospectOnlineKey}

	return &schema.Resource{
		Description: `Token introspect configuration adds support for 3rd party tokens to identify the user within IndyKite APIs.
		Token introspect enables the IndyKite platform to identify end users by third party tokens,
		validate these tokens, and use their content in the IndyKite platform.
		To verify these tokens, you need to create a configuration that describes how to do the token introspection.
		`,
		CreateContext: resTokenIntrospectCreate,
		ReadContext:   resTokenIntrospectRead,
		UpdateContext: resTokenIntrospectUpdate,
		DeleteContext: resTokenIntrospectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:   locationSchema(),
			customerIDKey: setComputed(customerIDSchema()),
			appSpaceIDKey: setComputed(appSpaceIDSchema()),

			nameKey:        nameSchema(),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),
			createdByKey:   createdBySchema(),
			updatedByKey:   updatedBySchema(),

			tokenIntrospectJWTKey: setExactlyOneOf(&schema.Schema{
				Type:        schema.TypeList,
				MaxItems:    1,
				Description: "Specifies all attributes required to match a JWT token.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					tokenIntrospectIssuerKey: {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.IsURLWithHTTPS,
						Description:  "Issuer is used to exact match based on `iss` claim in JWT.",
					},
					tokenIntrospectAudienceKey: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Audience is used to exact match based on `aud` claim in JWT.",
					},
				}},
			}, tokenIntrospectJWTKey, matcherOneOf),
			tokenIntrospectOpaqueKey: setExactlyOneOf(&schema.Schema{
				Type:        schema.TypeList,
				Description: "Specify opaque token matcher. Currently we support only 1 opaque matcher per application space.",
				MaxItems:    1,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					tokenIntrospectHintKey: {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "To differentiate between multiple opaque tokens configurations, hint must be provided. Hint is case sensitive plain text, that is expected to be provided in token introspect request, if there are multiple opaque tokens configurations.",
						ValidateFunc: validation.StringLenBetween(1, 50),
					},
				}},
				RequiredWith: []string{tokenIntrospectOnlineKey + ".0." + tokenIntrospectUserInfoEPKey},
			}, tokenIntrospectOpaqueKey, matcherOneOf),

			tokenIntrospectOfflineKey: setExactlyOneOf(&schema.Schema{
				Type:        schema.TypeList,
				MaxItems:    1,
				Description: "Offline validation works only with JWT and checks token locally.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					tokenIntrospectPublicJWKsKey: {
						Type:        schema.TypeList,
						Description: "Public JWKs to validate signature of JWT. If there are no public keys specified, they will be fetched and cached from jwks_uri at https://jwt-issuer.tld/.well-known/openid-configuration",
						Optional:    true,
						MinItems:    0,
						MaxItems:    10,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							ValidateFunc:     validation.StringLenBetween(96, 8192),
							DiffSuppressFunc: structure.SuppressJsonDiff,
						},
					},
				}},
			}, tokenIntrospectOfflineKey, validationOneOf),
			tokenIntrospectOnlineKey: setExactlyOneOf(&schema.Schema{
				Type:        schema.TypeList,
				MaxItems:    1,
				Description: "Online validation works with both JWT and Opaque tokens. It will call userinfo endpoint to validate token and fetch user claims.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					tokenIntrospectUserInfoEPKey: {
						Type: schema.TypeString,
						Description: `URI of userinfo endpoint which will be used to validate access token.
    And also fetch user claims when opaque token is received

    It can remain empty, if JWT token matcher is used.
    Then the URI under "userinfo_endpoint" in .well-known/openid-configuration endpoint is used.`,
						Optional:     true,
						ValidateFunc: validation.IsURLWithHTTPS,
					},
					tokenIntrospectCacheTTLKey: {
						Type: schema.TypeInt,
						Description: `Cache TTL of token validity can be used to minimize calls to userinfo endpoint.
    The final cache TTL will be set to lower limit of this value and exp claim of JWT token.
    If not set, token will not be cached and call to userinfo endpoint will be made on every request.

    However, token validity will be checked first if possible (JWT tokens).
    If token is expired, userinfo endpoint will not be called, nor cache checked.`,
						ValidateFunc: validation.IntBetween(0, 3600),
						Optional:     true,
					},
				}},
			}, tokenIntrospectOnlineKey, validationOneOf),

			tokenIntrospectClaimsMappingKey: {
				Type: schema.TypeMap,
				//nolint:lll
				Description: `ClaimsMapping specifies which claims from the token should be mapped to new names and name of property in IKG.
    Be aware, that this can override any existing claims, which might not be accessible anymore by internal services.
    And with the highest priority, there is mapping of sub claim to 'external_id'. So you shouldn't ever use 'external_id' as a key.

    Key specifies the new name and also the name of the property in IKG.
    Value specifies which claim to map and how.`,
				Optional: true,
				ValidateDiagFunc: validation.AllDiag(
					validation.MapKeyLenBetween(2, 256),
					validation.MapKeyMatch(tokenIntrospectIkgPropertyRegex, "invalid IKG property name"),
					validation.MapValueLenBetween(1, 256),
				),
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			tokenIntrospectSubClaimKey: {
				Type:         schema.TypeString,
				Description:  `Sub claim is used to match DigitalTwin with external_id. If not specified, standard 'sub' claim will be used. Either 'sub' or specified claim will then also be mapped to 'external_id' claim.`,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			tokenIntrospectIKGNodeTypeKey: {
				Type:        schema.TypeString,
				Description: "Node type in IKG to which we will try to match sub claim with DT external_id.",
				Required:    true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 64),
					validation.StringMatch(tokenIntrospectIkgNodeTypeRegex, "must be valid IKG Node Type"),
				),
			},
			tokenIntrospectPerformUpsertKey: {
				Type: schema.TypeBool,
				Description: `Perform Upsert specify, if we should create and/or update DigitalTwin in IKG if it doesn't exist with.
	In future this will perform upsert also on properties that are derived from token.`,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resTokenIntrospectCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	req := buildTokenIntrospectRequest(data)
	req.ProjectID = data.Get(locationKey).(string)
	req.Name = data.Get(nameKey).(string)
	req.DisplayName = stringValue(optionalString(data, displayNameKey))
	req.Description = stringValue(optionalString(data, descriptionKey))

	var resp TokenIntrospectResponse
	err := clientCtx.GetClient().Post(ctx, "/token-introspects", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

	return resTokenIntrospectRead(ctx, data, meta)
}

func resTokenIntrospectRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp TokenIntrospectResponse
	// Support both ID and name?location=parent_id formats
	path := buildReadPath("/token-introspects", data)
	err := clientCtx.GetClient().Get(ctx, path, &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceID)

	// Set location based on which is present
	if resp.AppSpaceID != "" {
		setData(&d, data, locationKey, resp.AppSpaceID)
	} else if resp.CustomerID != "" {
		setData(&d, data, locationKey, resp.CustomerID)
	}

	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, createdByKey, resp.CreatedBy)
	setData(&d, data, updatedByKey, resp.UpdatedBy)

	// Set token matcher (JWT or Opaque)
	if resp.JWT != nil {
		setData(&d, data, tokenIntrospectJWTKey, []map[string]any{{
			tokenIntrospectIssuerKey:   resp.JWT.Issuer,
			tokenIntrospectAudienceKey: resp.JWT.Audience,
		}})
	} else if resp.Opaque != nil {
		setData(&d, data, tokenIntrospectOpaqueKey, []map[string]any{{
			tokenIntrospectHintKey: resp.Opaque.Hint,
		}})
	}

	// Set validation (Offline or Online)
	if resp.Offline != nil {
		setData(&d, data, tokenIntrospectOfflineKey, []map[string]any{{
			tokenIntrospectPublicJWKsKey: resp.Offline.PublicJWKs,
		}})
	} else if resp.Online != nil {
		setData(&d, data, tokenIntrospectOnlineKey, []map[string]any{{
			tokenIntrospectUserInfoEPKey: resp.Online.UserinfoEndpoint,
			tokenIntrospectCacheTTLKey:   resp.Online.CacheTTL,
		}})
	}

	// Set claims mapping
	claimsMapping := make(map[string]any, len(resp.ClaimsMapping))
	for key, claim := range resp.ClaimsMapping {
		if claim != nil {
			claimsMapping[key] = claim.Selector
		}
	}
	setData(&d, data, tokenIntrospectClaimsMappingKey, claimsMapping)

	// Set sub claim
	if resp.SubClaim != nil {
		setData(&d, data, tokenIntrospectSubClaimKey, resp.SubClaim.Selector)
	}

	setData(&d, data, tokenIntrospectIKGNodeTypeKey, resp.IKGNodeType)
	setData(&d, data, tokenIntrospectPerformUpsertKey, resp.PerformUpsert)

	return d
}

func resTokenIntrospectUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := UpdateTokenIntrospectRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	// Add changed fields
	if data.HasChange(tokenIntrospectJWTKey) || data.HasChange(tokenIntrospectOpaqueKey) ||
		data.HasChange(tokenIntrospectOfflineKey) || data.HasChange(tokenIntrospectOnlineKey) ||
		data.HasChange(tokenIntrospectClaimsMappingKey) || data.HasChange(tokenIntrospectSubClaimKey) ||
		data.HasChange(tokenIntrospectIKGNodeTypeKey) || data.HasChange(tokenIntrospectPerformUpsertKey) {
		tokenReq := buildTokenIntrospectRequest(data)
		req.JWT = tokenReq.JWT
		req.Opaque = tokenReq.Opaque
		req.Offline = tokenReq.Offline
		req.Online = tokenReq.Online
		req.ClaimsMapping = tokenReq.ClaimsMapping
		req.SubClaim = tokenReq.SubClaim

		if data.HasChange(tokenIntrospectIKGNodeTypeKey) {
			nodeType := data.Get(tokenIntrospectIKGNodeTypeKey).(string)
			req.IKGNodeType = &nodeType
		}
		if data.HasChange(tokenIntrospectPerformUpsertKey) {
			performUpsert := data.Get(tokenIntrospectPerformUpsertKey).(bool)
			req.PerformUpsert = &performUpsert
		}
	}

	var resp TokenIntrospectResponse
	err := clientCtx.GetClient().Put(ctx, "/token-introspects/"+data.Id(), req, &resp)
	if HasFailed(&d, err) {
		return d
	}

	return resTokenIntrospectRead(ctx, data, meta)
}

func resTokenIntrospectDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	err := clientCtx.GetClient().Delete(ctx, "/token-introspects/"+data.Id())
	HasFailed(&d, err)
	return d
}

// buildTokenIntrospectRequest builds a CreateTokenIntrospectRequest from schema data.
func buildTokenIntrospectRequest(data *schema.ResourceData) CreateTokenIntrospectRequest {
	req := CreateTokenIntrospectRequest{
		IKGNodeType:   data.Get(tokenIntrospectIKGNodeTypeKey).(string),
		PerformUpsert: data.Get(tokenIntrospectPerformUpsertKey).(bool),
	}

	// Set JWT or Opaque matcher
	if val, ok := data.GetOk(tokenIntrospectJWTKey); ok {
		mapVal := val.([]any)[0].(map[string]any)
		req.JWT = &TokenIntrospectJWT{
			Issuer:   mapVal[tokenIntrospectIssuerKey].(string),
			Audience: mapVal[tokenIntrospectAudienceKey].(string),
		}
	}
	if val, ok := data.GetOk(tokenIntrospectOpaqueKey); ok {
		mapVal := val.([]any)[0].(map[string]any)
		req.Opaque = &TokenIntrospectOpaque{
			Hint: mapVal[tokenIntrospectHintKey].(string),
		}
	}

	// Set Offline or Online validation
	if val, ok := data.GetOk(tokenIntrospectOfflineKey); ok {
		listVal := val.([]any)
		if len(listVal) > 0 && listVal[0] != nil {
			mapVal := listVal[0].(map[string]any)
			publicJWKs := rawArrayToTypedArray[string](mapVal[tokenIntrospectPublicJWKsKey])
			req.Offline = &TokenIntrospectOffline{
				PublicJWKs: publicJWKs,
			}
		} else {
			// Empty offline validation
			req.Offline = &TokenIntrospectOffline{}
		}
	}
	if val, ok := data.GetOk(tokenIntrospectOnlineKey); ok {
		if listVal := val.([]any); len(listVal) > 0 && listVal[0] != nil {
			mapVal := listVal[0].(map[string]any)
			req.Online = &TokenIntrospectOnline{
				UserinfoEndpoint: mapVal[tokenIntrospectUserInfoEPKey].(string),
				CacheTTL:         mapVal[tokenIntrospectCacheTTLKey].(int),
			}
		}
	}

	// Set claims mapping
	if claimsMapping := data.Get(tokenIntrospectClaimsMappingKey).(map[string]any); len(claimsMapping) > 0 {
		req.ClaimsMapping = make(map[string]*TokenIntrospectClaim, len(claimsMapping))
		for key, selector := range claimsMapping {
			req.ClaimsMapping[key] = &TokenIntrospectClaim{
				Selector: selector.(string),
			}
		}
	}

	// Set sub claim
	if val := data.Get(tokenIntrospectSubClaimKey).(string); val != "" {
		req.SubClaim = &TokenIntrospectClaim{Selector: val}
	}

	return req
}
