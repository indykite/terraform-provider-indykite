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
	"errors"
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
	tokenIntrospectDiscoveryKey  = "discovery"
	tokenIntrospectOnlineKey     = "online_validation"
	tokenIntrospectUserInfoEPKey = "user_info_endpoint"
	tokenIntrospectCacheTTLKey   = "cache_ttl"

	tokenIntrospectOnlineOAuth2Key    = "online_oauth2_validation"
	tokenIntrospectIntrospectionEPKey = "introspection_endpoint"
	tokenIntrospectSecretBasicKey     = "secret_basic"
	tokenIntrospectClientIDKey        = "client_id"
	tokenIntrospectClientSecretKey    = "client_secret"

	tokenIntrospectDiscoveryOIDC    = "oidc"
	tokenIntrospectDiscoveryRFC8414 = "rfc8414"

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
	validationOneOf := []string{tokenIntrospectOfflineKey, tokenIntrospectOnlineKey, tokenIntrospectOnlineOAuth2Key}

	return &schema.Resource{
		Description: `Token introspect configuration adds support for 3rd party tokens to identify the user within IndyKite APIs.
		Token introspect enables the IndyKite platform to identify end users by third party tokens,
		validate these tokens, and use their content in the IndyKite platform.
		To verify these tokens, you need to create a configuration that describes how to do the token introspection.
		`,
		CustomizeDiff: validateTokenIntrospectOpaqueValidation,
		CreateContext: resTokenIntrospectCreate,
		ReadContext:   resTokenIntrospectRead,
		UpdateContext: resTokenIntrospectUpdate,
		DeleteContext: resTokenIntrospectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: projectStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:   locationSchema(),
			projectIDKey:  projectIDSchema(),
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
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validateCharLength(1, 150),
						Description:  "Audience is used to exact match based on `aud` claim in JWT, up to 150 characters.",
					},
				}},
			}, tokenIntrospectJWTKey, matcherOneOf),
			tokenIntrospectOpaqueKey: setExactlyOneOf(&schema.Schema{
				Type: schema.TypeList,
				Description: "Specify opaque token matcher. A project can have several opaque matchers, " +
					"each with a different hint.",
				MaxItems: 1,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					tokenIntrospectHintKey: {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "To differentiate between multiple opaque tokens configurations, hint must be provided. Hint is case sensitive plain text, that is expected to be provided in token introspect request, if there are multiple opaque tokens configurations.",
						ValidateFunc: validation.StringLenBetween(1, 50),
					},
				}},
			}, tokenIntrospectOpaqueKey, matcherOneOf),

			tokenIntrospectOfflineKey: setExactlyOneOf(&schema.Schema{
				Type:        schema.TypeList,
				MaxItems:    1,
				Description: "Offline validation works only with JWT and checks token locally.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					tokenIntrospectPublicJWKsKey: {
						Type:        schema.TypeList,
						Description: "Public JWKs to validate signature of JWT. If there are no public keys specified, they will be fetched and cached from the jwks_uri of the metadata document selected by `discovery`.",
						Optional:    true,
						MinItems:    0,
						MaxItems:    10,
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							ValidateFunc:     validation.StringLenBetween(96, 8192),
							DiffSuppressFunc: structure.SuppressJsonDiff,
						},
					},
					tokenIntrospectDiscoveryKey: {
						Type: schema.TypeString,
						Description: `Metadata document used to find the issuer's jwks_uri when no public_jwks are specified.
    "oidc" (default) reads https://jwt-issuer.tld/.well-known/openid-configuration.
    "rfc8414" reads the RFC 8414 document https://jwt-issuer.tld/.well-known/oauth-authorization-server,
    with the well-known suffix inserted between the host and the path of the issuer.
    Not used when public_jwks are specified.`,
						Optional: true,
						ValidateFunc: validation.StringInSlice(
							[]string{tokenIntrospectDiscoveryOIDC, tokenIntrospectDiscoveryRFC8414}, false),
						DiffSuppressFunc: suppressTokenIntrospectDiscoveryDiff,
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
			tokenIntrospectOnlineOAuth2Key: setExactlyOneOf(&schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Description: "OAuth 2.0 online validation works with both JWT and Opaque tokens. " +
					"It will call an RFC 7662 token introspection endpoint to validate token and fetch token claims. " +
					"The token is accepted only when the endpoint answers with `\"active\": true`.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					tokenIntrospectIntrospectionEPKey: {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "URI of the RFC 7662 token introspection endpoint, must be https. There is no discovery for it.",
						ValidateFunc: validation.IsURLWithHTTPS,
					},
					tokenIntrospectSecretBasicKey: {
						Type:     schema.TypeList,
						Required: true,
						MinItems: 1,
						MaxItems: 1,
						Description: "HTTP Basic authentication against the introspection endpoint " +
							"(`client_secret_basic`).",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							tokenIntrospectClientIDKey: {
								Type:         schema.TypeString,
								Required:     true,
								Description:  "Client ID used for HTTP Basic authentication, up to 256 characters.",
								ValidateFunc: validation.StringLenBetween(1, 256),
							},
							tokenIntrospectClientSecretKey: {
								Type:      schema.TypeString,
								Required:  true,
								Sensitive: true,
								Description: "Client secret used for HTTP Basic authentication, up to 1024 characters. " +
									"It is write-only: the API never returns it, " +
									"so the value is kept from the Terraform state.",
								ValidateFunc: validation.StringLenBetween(1, 1024),
							},
						}},
					},
					tokenIntrospectCacheTTLKey: {
						Type: schema.TypeInt,
						Description: `Cache TTL of token validity can be used to minimize calls to introspection endpoint.
    The final cache TTL will be set to lower limit of this value and exp claim of JWT token.
    If not set, token will not be cached and call to introspection endpoint will be made on every request.

    However, token validity will be checked first if possible (JWT tokens).
    If token is expired, introspection endpoint will not be called, nor cache checked.`,
						ValidateFunc: validation.IntBetween(0, 3600),
						Optional:     true,
					},
				}},
			}, tokenIntrospectOnlineOAuth2Key, validationOneOf),

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
					validateMapValueCharLength(1, 200),
				),
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			tokenIntrospectSubClaimKey: {
				Type:         schema.TypeString,
				Description:  `Sub claim is used to match DigitalTwin with external_id. If not specified, standard 'sub' claim will be used. Either 'sub' or specified claim will then also be mapped to 'external_id' claim.`,
				Optional:     true,
				ValidateFunc: validateCharLength(1, 200),
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
	req.ProjectID = projectIDFromData(data)
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
	setProjectIDData(&d, data, resp.AppSpaceID, resp.CustomerID)

	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, createdByKey, resp.CreatedBy)
	setData(&d, data, updatedByKey, resp.UpdatedBy)

	// Set token matcher (JWT or Opaque), clearing the one the API did not return
	var jwtMatcher, opaqueMatcher []map[string]any
	if resp.JWT != nil {
		jwtMatcher = []map[string]any{{
			tokenIntrospectIssuerKey:   resp.JWT.Issuer,
			tokenIntrospectAudienceKey: resp.JWT.Audience,
		}}
	} else if resp.Opaque != nil {
		opaqueMatcher = []map[string]any{{
			tokenIntrospectHintKey: resp.Opaque.Hint,
		}}
	}
	setData(&d, data, tokenIntrospectJWTKey, jwtMatcher)
	setData(&d, data, tokenIntrospectOpaqueKey, opaqueMatcher)

	// Set validation (Offline, Online or OnlineOAuth2), clearing the ones the API did not return
	var offline, online, onlineOAuth2 []map[string]any
	switch {
	case resp.Offline != nil:
		offline = []map[string]any{{
			tokenIntrospectPublicJWKsKey: resp.Offline.PublicJWKs,
			tokenIntrospectDiscoveryKey:  resp.Offline.Discovery,
		}}
	case resp.Online != nil:
		online = []map[string]any{{
			tokenIntrospectUserInfoEPKey: resp.Online.UserinfoEndpoint,
			tokenIntrospectCacheTTLKey:   resp.Online.CacheTTL,
		}}
	case resp.OnlineOAuth2 != nil:
		onlineOAuth2 = flattenTokenIntrospectOnlineOAuth2(data, resp.OnlineOAuth2)
	}
	setData(&d, data, tokenIntrospectOfflineKey, offline)
	setData(&d, data, tokenIntrospectOnlineKey, online)
	setData(&d, data, tokenIntrospectOnlineOAuth2Key, onlineOAuth2)

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

	if hasNameChanged(&d, data) {
		return d
	}

	req := UpdateTokenIntrospectRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	req.IKGNodeType = data.Get(tokenIntrospectIKGNodeTypeKey).(string)
	req.PerformUpsert = data.Get(tokenIntrospectPerformUpsertKey).(bool)

	// The backend validates the full object on update (PUT), so always send the
	// matcher, validation and claim fields - not only when they changed.
	tokenReq := buildTokenIntrospectRequest(data)
	req.JWT = tokenReq.JWT
	req.Opaque = tokenReq.Opaque
	req.Offline = tokenReq.Offline
	req.Online = tokenReq.Online
	req.OnlineOAuth2 = tokenReq.OnlineOAuth2
	req.ClaimsMapping = tokenReq.ClaimsMapping
	req.SubClaim = tokenReq.SubClaim

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
			discovery, _ := mapVal[tokenIntrospectDiscoveryKey].(string)
			req.Offline = &TokenIntrospectOffline{
				PublicJWKs: publicJWKs,
				Discovery:  discovery,
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
	if val, ok := data.GetOk(tokenIntrospectOnlineOAuth2Key); ok {
		if listVal := val.([]any); len(listVal) > 0 && listVal[0] != nil {
			mapVal := listVal[0].(map[string]any)
			req.OnlineOAuth2 = &TokenIntrospectOnlineOAuth2{
				IntrospectionEndpoint: mapVal[tokenIntrospectIntrospectionEPKey].(string),
				CacheTTL:              mapVal[tokenIntrospectCacheTTLKey].(int),
			}
			secretList, _ := mapVal[tokenIntrospectSecretBasicKey].([]any)
			if len(secretList) > 0 && secretList[0] != nil {
				secretVal := secretList[0].(map[string]any)
				req.OnlineOAuth2.SecretBasic = &TokenIntrospectSecretBasic{
					ClientID:     secretVal[tokenIntrospectClientIDKey].(string),
					ClientSecret: secretVal[tokenIntrospectClientSecretKey].(string),
				}
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

// flattenTokenIntrospectOnlineOAuth2 converts the API OAuth 2.0 validation into a Terraform block.
// The API never returns client_secret, so it is preserved from the current state.
func flattenTokenIntrospectOnlineOAuth2(
	data *schema.ResourceData,
	oauth2 *TokenIntrospectOnlineOAuth2,
) []map[string]any {
	block := map[string]any{
		tokenIntrospectIntrospectionEPKey: oauth2.IntrospectionEndpoint,
		tokenIntrospectCacheTTLKey:        oauth2.CacheTTL,
	}
	if oauth2.SecretBasic != nil {
		secretPath := tokenIntrospectOnlineOAuth2Key + ".0." + tokenIntrospectSecretBasicKey + ".0."
		clientSecret, _ := data.Get(secretPath + tokenIntrospectClientSecretKey).(string)
		block[tokenIntrospectSecretBasicKey] = []map[string]any{{
			tokenIntrospectClientIDKey:     oauth2.SecretBasic.ClientID,
			tokenIntrospectClientSecretKey: clientSecret,
		}}
	}
	return []map[string]any{block}
}

// suppressTokenIntrospectDiscoveryDiff hides the diffs caused by API defaults of discovery:
// an unset value is read back as "oidc", and with pinned public_jwks it is read back empty.
func suppressTokenIntrospectDiscoveryDiff(_, oldValue, newValue string, data *schema.ResourceData) bool {
	if newValue == "" && oldValue == tokenIntrospectDiscoveryOIDC {
		return true
	}
	publicJWKs, _ := data.Get(tokenIntrospectOfflineKey + ".0." + tokenIntrospectPublicJWKsKey).([]any)
	return oldValue == "" && len(publicJWKs) > 0
}

// validateTokenIntrospectOpaqueValidation enforces that an opaque_matcher is paired with an online
// validation: offline_validation works only with JWT, and online_validation needs user_info_endpoint
// as there is no issuer to discover it from.
func validateTokenIntrospectOpaqueValidation(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if opaque, _ := d.Get(tokenIntrospectOpaqueKey).([]any); len(opaque) == 0 {
		return nil
	}
	if offline, _ := d.Get(tokenIntrospectOfflineKey).([]any); len(offline) > 0 {
		return errors.New("`offline_validation` can be used only with `jwt_matcher`")
	}
	userInfoEPPath := tokenIntrospectOnlineKey + ".0." + tokenIntrospectUserInfoEPKey
	if online, _ := d.Get(tokenIntrospectOnlineKey).([]any); len(online) > 0 && d.NewValueKnown(userInfoEPPath) {
		if v, _ := d.Get(userInfoEPPath).(string); v == "" {
			return errors.New("`online_validation.0.user_info_endpoint` is required with `opaque_matcher`")
		}
	}
	return nil
}
