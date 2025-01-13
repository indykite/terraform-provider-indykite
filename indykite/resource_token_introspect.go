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
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	"google.golang.org/protobuf/types/known/durationpb"
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
	readContext := configReadContextFunc(resourceTokenIntrospectFlatten)

	matcherOneOf := []string{tokenIntrospectJWTKey, tokenIntrospectOpaqueKey}
	validationOneOf := []string{tokenIntrospectOfflineKey, tokenIntrospectOnlineKey}

	return &schema.Resource{
		Description: `Token introspect configuration adds support for 3rd party tokens to identify the user within IndyKite APIs.`,

		CreateContext: configCreateContextFunc(resourceTokenIntrospectBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceTokenIntrospectBuild, readContext),
		DeleteContext: configDeleteContextFunc(),
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

func resourceTokenIntrospectFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) diag.Diagnostics {
	var d diag.Diagnostics
	tiCfg := resp.GetConfigNode().GetTokenIntrospectConfig()
	setData(&d, data, tokenIntrospectPerformUpsertKey, tiCfg.GetPerformUpsert())
	setData(&d, data, tokenIntrospectIKGNodeTypeKey, tiCfg.GetIkgNodeType())

	claimsMapping := make(map[string]any, len(tiCfg.GetClaimsMapping()))
	for key, claim := range tiCfg.GetClaimsMapping() {
		claimsMapping[key] = claim.GetSelector()
	}
	setData(&d, data, tokenIntrospectClaimsMappingKey, claimsMapping)
	setData(&d, data, tokenIntrospectSubClaimKey, tiCfg.GetSubClaim().GetSelector())

	switch matcher := tiCfg.GetTokenMatcher().(type) {
	case *configpb.TokenIntrospectConfig_Jwt:
		setData(&d, data, tokenIntrospectJWTKey, []map[string]any{{
			tokenIntrospectIssuerKey:   matcher.Jwt.GetIssuer(),
			tokenIntrospectAudienceKey: matcher.Jwt.GetAudience(),
		}})
	case *configpb.TokenIntrospectConfig_Opaque_:
		setData(&d, data, tokenIntrospectOpaqueKey, []map[string]any{{
			tokenIntrospectHintKey: matcher.Opaque.GetHint(),
		}})
	default:
		return append(d, buildPluginError(fmt.Sprintf("unsupported Token Matcher: %T", matcher)))
	}

	switch tv := tiCfg.GetValidation().(type) {
	case *configpb.TokenIntrospectConfig_Offline_:
		jwks := make([]string, len(tv.Offline.GetPublicJwks()))
		for i, jwk := range tv.Offline.GetPublicJwks() {
			jwks[i] = string(jwk)
		}
		setData(&d, data, tokenIntrospectOfflineKey, []map[string]any{{
			tokenIntrospectPublicJWKsKey: jwks,
		}})
	case *configpb.TokenIntrospectConfig_Online_:
		setData(&d, data, tokenIntrospectOnlineKey, []map[string]any{{
			tokenIntrospectUserInfoEPKey: tv.Online.GetUserinfoEndpoint(),
			tokenIntrospectCacheTTLKey:   int(tv.Online.GetCacheTtl().GetSeconds()),
		}})
	default:
		return append(d, buildPluginError(fmt.Sprintf("unsupported Validation: %T", tv)))
	}

	return d
}

func resourceTokenIntrospectBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	claimsMapping := data.Get(tokenIntrospectClaimsMappingKey).(map[string]any)
	cfg := &configpb.TokenIntrospectConfig{
		ClaimsMapping: make(map[string]*configpb.TokenIntrospectConfig_Claim, len(claimsMapping)),
		IkgNodeType:   data.Get(tokenIntrospectIKGNodeTypeKey).(string),
		PerformUpsert: data.Get(tokenIntrospectPerformUpsertKey).(bool),
	}
	if val := data.Get(tokenIntrospectSubClaimKey).(string); val != "" {
		cfg.SubClaim = &configpb.TokenIntrospectConfig_Claim{Selector: val}
	}
	for key, selector := range claimsMapping {
		cfg.ClaimsMapping[key] = &configpb.TokenIntrospectConfig_Claim{
			Selector: selector.(string),
		}
	}

	if val, ok := data.GetOk(tokenIntrospectJWTKey); ok {
		mapVal := val.([]any)[0].(map[string]any)
		cfg.TokenMatcher = &configpb.TokenIntrospectConfig_Jwt{
			Jwt: &configpb.TokenIntrospectConfig_JWT{
				Issuer:   mapVal[tokenIntrospectIssuerKey].(string),
				Audience: mapVal[tokenIntrospectAudienceKey].(string),
			},
		}
	}
	if val, ok := data.GetOk(tokenIntrospectOpaqueKey); ok {
		mapVal := val.([]any)[0].(map[string]any)
		cfg.TokenMatcher = &configpb.TokenIntrospectConfig_Opaque_{
			Opaque: &configpb.TokenIntrospectConfig_Opaque{
				Hint: mapVal[tokenIntrospectHintKey].(string),
			},
		}
	}

	if val, ok := data.GetOk(tokenIntrospectOfflineKey); ok {
		mapVal := val.([]any)[0].(map[string]any)
		cfg.Validation = &configpb.TokenIntrospectConfig_Offline_{
			Offline: &configpb.TokenIntrospectConfig_Offline{
				PublicJwks: rawArrayToTypedArray[[]byte](mapVal[tokenIntrospectPublicJWKsKey]),
			},
		}
	}
	if val, ok := data.GetOk(tokenIntrospectOnlineKey); ok {
		mapVal := val.([]any)[0].(map[string]any)
		var cacheTTL *durationpb.Duration
		if val := mapVal[tokenIntrospectCacheTTLKey].(int); val > 0 {
			cacheTTL = durationpb.New(time.Duration(val) * time.Second)
		}
		cfg.Validation = &configpb.TokenIntrospectConfig_Online_{
			Online: &configpb.TokenIntrospectConfig_Online{
				UserinfoEndpoint: mapVal[tokenIntrospectUserInfoEPKey].(string),
				CacheTtl:         cacheTTL,
			},
		}
	}

	builder.WithTokenIntrospectConfig(cfg)
}
