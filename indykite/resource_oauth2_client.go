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
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

const (
	providerTypeKey          = "provider_type"
	clientIDKey              = "client_id"
	clientSecretKey          = "client_secret"
	redirectURIKey           = "redirect_uri"
	defaultScopesKey         = "default_scopes"
	allowedScopesKey         = "allowed_scopes"
	allowSignupKey           = "allow_signup"
	issuerKey                = "issuer"
	authorizationEndpointKey = "authorization_endpoint"
	tokenEndpointKey         = "token_endpoint"
	discoveryURLKey          = "discovery_url"
	userinfoEndpointKey      = "userinfo_endpoint"
	jwksURIKey               = "jwks_uri"
	imageURLKey              = "image_url"
	tenantKey                = "tenant"
	hostedDomainKey          = "hosted_domain"
	authStyleKey             = "auth_style"
	privateKeyPemKey         = "private_key_pem"
	privateKeyIDKey          = "private_key_id"
	teamIDKey                = "team_id"
)

var (
	oauth2AppProviderTypes = map[string]configpb.ProviderType{
		"apple.com":             configpb.ProviderType_PROVIDER_TYPE_APPLE_COM,
		"amazon.com":            configpb.ProviderType_PROVIDER_TYPE_AMAZON_COM,
		"amazoncognito.com":     configpb.ProviderType_PROVIDER_TYPE_AMAZONCOGNITO_COM,
		"bitbucket":             configpb.ProviderType_PROVIDER_TYPE_BITBUCKET,
		"cern.ch":               configpb.ProviderType_PROVIDER_TYPE_CERN_CH,
		"facebook.com":          configpb.ProviderType_PROVIDER_TYPE_FACEBOOK_COM,
		"fitbit.com":            configpb.ProviderType_PROVIDER_TYPE_FITBIT_COM,
		"foursquare.com":        configpb.ProviderType_PROVIDER_TYPE_FOURSQUARE_COM,
		"github.com":            configpb.ProviderType_PROVIDER_TYPE_GITHUB_COM,
		"gitlab.com":            configpb.ProviderType_PROVIDER_TYPE_GITLAB_COM,
		"google.com":            configpb.ProviderType_PROVIDER_TYPE_GOOGLE_COM,
		"heroku.com":            configpb.ProviderType_PROVIDER_TYPE_HEROKU_COM,
		"hipchat.com":           configpb.ProviderType_PROVIDER_TYPE_HIPCHAT_COM,
		"instagram.com":         configpb.ProviderType_PROVIDER_TYPE_INSTAGRAM_COM,
		"kakao.com":             configpb.ProviderType_PROVIDER_TYPE_KAKAO_COM,
		"linkedin.com":          configpb.ProviderType_PROVIDER_TYPE_LINKEDIN_COM,
		"mailchimp.com":         configpb.ProviderType_PROVIDER_TYPE_MAILCHIMP_COM,
		"mail.ru":               configpb.ProviderType_PROVIDER_TYPE_MAIL_RU,
		"mediamath.com":         configpb.ProviderType_PROVIDER_TYPE_MEDIAMATH_COM,
		"sandbox.mediamath.com": configpb.ProviderType_PROVIDER_TYPE_SANDBOX_MEDIAMATH_COM,
		"live.com":              configpb.ProviderType_PROVIDER_TYPE_LIVE_COM,
		"microsoft.com":         configpb.ProviderType_PROVIDER_TYPE_MICROSOFT_COM,
		"health.nokia.com":      configpb.ProviderType_PROVIDER_TYPE_HEALTH_NOKIA_COM,
		"odnoklassniki.ru":      configpb.ProviderType_PROVIDER_TYPE_ODNOKLASSNIKI_RU,
		"paypal.com":            configpb.ProviderType_PROVIDER_TYPE_PAYPAL_COM,
		"sandbox.paypal.com":    configpb.ProviderType_PROVIDER_TYPE_SANDBOX_PAYPAL_COM,
		"slack.com":             configpb.ProviderType_PROVIDER_TYPE_SLACK_COM,
		"spotify.com":           configpb.ProviderType_PROVIDER_TYPE_SPOTIFY_COM,
		"stackoverflow.com":     configpb.ProviderType_PROVIDER_TYPE_STACKOVERFLOW_COM,
		"twitch.tv":             configpb.ProviderType_PROVIDER_TYPE_TWITCH_TV,
		"uber.com":              configpb.ProviderType_PROVIDER_TYPE_UBER_COM,
		"vk.com":                configpb.ProviderType_PROVIDER_TYPE_VK_COM,
		"yahoo.com":             configpb.ProviderType_PROVIDER_TYPE_YAHOO_COM,
		"yandex.com":            configpb.ProviderType_PROVIDER_TYPE_YANDEX_COM,
		"authenteq.com":         configpb.ProviderType_PROVIDER_TYPE_AUTHENTEQ_COM,
		"indykite.id":           configpb.ProviderType_PROVIDER_TYPE_INDYKITE_ID,
		"indykite.me":           configpb.ProviderType_PROVIDER_TYPE_INDYKITE_ME,
		"bankid.no":             configpb.ProviderType_PROVIDER_TYPE_BANKID_NO,
		"bankid.com":            configpb.ProviderType_PROVIDER_TYPE_BANKID_COM,
		"custom":                configpb.ProviderType_PROVIDER_TYPE_CUSTOM,
		"vipps.no":              configpb.ProviderType_PROVIDER_TYPE_VIPPS_NO,
	}

	oauth2AppAuthStyles = map[string]configpb.AuthStyle{
		"auto_detect": configpb.AuthStyle_AUTH_STYLE_AUTO_DETECT,
		"in_params":   configpb.AuthStyle_AUTH_STYLE_IN_PARAMS,
		"in_header":   configpb.AuthStyle_AUTH_STYLE_IN_HEADER,
	}
)

func resourceOAuth2Client() *schema.Resource {
	readContext := configReadContextFunc(resourceOAuth2ClientFlatten)

	return &schema.Resource{
		CreateContext: configCreateContextFunc(resourceOAuth2ClientBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceOAuth2ClientBuild, readContext),
		DeleteContext: configDeleteContextFunc(),
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:    locationSchema(),
			customerIDKey:  setComputed(customerIDSchema()),
			appSpaceIDKey:  setComputed(appSpaceIDSchema()),
			tenantIDKey:    setComputed(tenantIDSchema()),
			nameKey:        nameSchema(),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),

			providerTypeKey:          oauth2AppProviderTypeSchema(),
			clientIDKey:              oauth2AppClientIDSchema(),
			clientSecretKey:          oauth2AppClientSecretSchema(),
			redirectURIKey:           oauth2AppRedirectURISchema(),
			defaultScopesKey:         oauth2AppDefaultScopesSchema(),
			allowedScopesKey:         oauth2AppAllowedScopesSchema(),
			allowSignupKey:           oauth2AppAllowSignupSchema(),
			issuerKey:                oauth2IssuerSchema(),
			authorizationEndpointKey: oauth2AuthorizationEndpointSchema(),
			tokenEndpointKey:         oauth2TokenEndpointSchema(),
			discoveryURLKey:          oauth2DiscoveryURLSchema(),
			userinfoEndpointKey:      oauth2UserinfoEndpointSchema(),
			jwksURIKey:               oauth2JWKsURISchema(),
			imageURLKey:              oauth2ImageURLSchema(),
			tenantKey:                oauth2TenantSchema(),
			hostedDomainKey:          oauth2HostedDomainSchema(),
			authStyleKey:             oauth2AuthStyleSchema(),
			privateKeyPemKey:         oauth2AppPrivateKeyPemSchema(),
			privateKeyIDKey:          oauth2AppPrivateKeyIDSchema(),
			teamIDKey:                oauth2AppTeamIDSchema(),
		},
	}
}

func resourceOAuth2ClientFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) diag.Diagnostics {
	var d diag.Diagnostics
	clientConf := resp.GetConfigNode().GetOauth2ClientConfig()
	if clientConf == nil {
		return diag.Diagnostics{buildPluginError("config in the response is not valid OAuth2ClientConfig")}
	}

	providerType, exist := ReverseProtoEnumMap(oauth2AppProviderTypes)[clientConf.ProviderType]
	if !exist {
		d = append(d, buildPluginError("BE send unsupported OAuth2 Provider Type: "+clientConf.ProviderType.String()))
	}
	setData(&d, data, providerTypeKey, providerType)

	setData(&d, data, clientIDKey, clientConf.ClientId)
	// Never set ClientSecret as it is never sent back
	setData(&d, data, redirectURIKey, clientConf.RedirectUri)
	setData(&d, data, defaultScopesKey, clientConf.DefaultScopes)
	setData(&d, data, allowedScopesKey, clientConf.AllowedScopes)
	setData(&d, data, allowSignupKey, clientConf.AllowSignup)
	setData(&d, data, issuerKey, clientConf.Issuer)
	setData(&d, data, authorizationEndpointKey, clientConf.AuthorizationEndpoint)
	setData(&d, data, tokenEndpointKey, clientConf.TokenEndpoint)
	setData(&d, data, discoveryURLKey, clientConf.DiscoveryUrl)
	setData(&d, data, userinfoEndpointKey, clientConf.UserinfoEndpoint)
	setData(&d, data, jwksURIKey, clientConf.JwksUri)
	setData(&d, data, imageURLKey, clientConf.ImageUrl)
	setData(&d, data, tenantKey, clientConf.Tenant)
	setData(&d, data, hostedDomainKey, clientConf.HostedDomain)

	authStyle, exist := ReverseProtoEnumMap(oauth2AppAuthStyles)[clientConf.AuthStyle]
	if !exist {
		d = append(d, buildPluginError("BE send unsupported OAuth2 AuthStyle: "+clientConf.AuthStyle.String()))
	}
	setData(&d, data, authStyleKey, authStyle)

	// Don't set PrivateKeyPEM as it is never sent back
	setData(&d, data, privateKeyIDKey, clientConf.PrivateKeyId)
	setData(&d, data, teamIDKey, clientConf.TeamId)

	return d
}

func resourceOAuth2ClientBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	configNode := &configpb.OAuth2ClientConfig{
		ProviderType:          oauth2AppProviderTypes[data.Get(providerTypeKey).(string)],
		ClientId:              data.Get(clientIDKey).(string),
		RedirectUri:           rawArrayToStringArray(data.Get(redirectURIKey)),
		DefaultScopes:         rawArrayToStringArray(data.Get(defaultScopesKey)),
		AllowedScopes:         rawArrayToStringArray(data.Get(allowedScopesKey)),
		AllowSignup:           data.Get(allowSignupKey).(bool),
		Issuer:                data.Get(issuerKey).(string),
		AuthorizationEndpoint: data.Get(authorizationEndpointKey).(string),
		TokenEndpoint:         data.Get(tokenEndpointKey).(string),
		DiscoveryUrl:          data.Get(discoveryURLKey).(string),
		UserinfoEndpoint:      data.Get(userinfoEndpointKey).(string),
		JwksUri:               data.Get(jwksURIKey).(string),
		ImageUrl:              data.Get(imageURLKey).(string),
		Tenant:                data.Get(tenantKey).(string),
		HostedDomain:          data.Get(hostedDomainKey).(string),
		AuthStyle:             oauth2AppAuthStyles[data.Get(authStyleKey).(string)],
		PrivateKeyId:          data.Get(privateKeyIDKey).(string),
		TeamId:                data.Get(teamIDKey).(string),
	}

	if data.HasChange(clientSecretKey) {
		configNode.ClientSecret = data.Get(clientSecretKey).(string)
	}
	if data.HasChange(privateKeyPemKey) {
		trim := strings.TrimSpace(data.Get(privateKeyPemKey).(string))
		configNode.PrivateKeyPem = []byte(trim)
	}

	builder.WithOAuth2ClientConfig(configNode)
}

func oauth2AppProviderTypeSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(getMapStringKeys(oauth2AppProviderTypes), false),
		Description:  `The OAuth2 Application provider type`,
	}
}

func oauth2AppClientIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(8, 512),
		Description:  `The OAuth2 Application Client ID`,
	}
}

func oauth2AppClientSecretSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true, // Some providers might use different secret, like Apple uses PEM
		Sensitive:    true,
		ValidateFunc: validation.StringLenBetween(8, 512),
		Description:  `The OAuth2 Application Client Secret`,
	}
}

func oauth2AppRedirectURISchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		// Add ValidateFunc with unique list of strings when supported
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			Description:  `The OAuth2 Application redirect URI`,
		},
	}
}

func oauth2AppDefaultScopesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		// Add ValidateFunc with unique list of strings when supported
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringLenBetween(1, 1024),
			Description:  `The OAuth2 Application default scopes`,
		},
	}
}

func oauth2AppAllowedScopesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		// Add ValidateFunc with unique list of strings when supported
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringLenBetween(1, 1024),
			Description:  `The OAuth2 Application allowed scopes`,
		},
	}
}

func oauth2AppAllowSignupSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Description: `The OAuth2 Application allow signup - used for Github`,
		Optional:    true,
	}
}

func oauth2IssuerSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  `URL using the https scheme with no query or fragment component that the OP asserts as its Issuer Identifier.`,
		Optional:     true,
		ValidateFunc: validation.IsURLWithHTTPS,
	}
}

func oauth2AuthorizationEndpointSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  `URL of the OP's OAuth 2.0 Authorization Endpoint.`,
		Optional:     true,
		ValidateFunc: validation.IsURLWithHTTPS,
	}
}

func oauth2TokenEndpointSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  `URL of the OP's OAuth 2.0 Token Endpoint.`,
		Optional:     true,
		ValidateFunc: validation.IsURLWithHTTPS,
	}
}

func oauth2DiscoveryURLSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  `The OAuth2 Application Discovery URL`,
		Optional:     true,
		ValidateFunc: validation.IsURLWithHTTPS,
	}
}

func oauth2UserinfoEndpointSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  `URL of the OP's UserInfo Endpoint`,
		Optional:     true,
		ValidateFunc: validation.IsURLWithHTTPS,
	}
}

func oauth2JWKsURISchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  `URL of the OP's JSON Web Key Set [JWK] document.`,
		Optional:     true,
		ValidateFunc: validation.IsURLWithHTTPS,
	}
}

func oauth2ImageURLSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  `The OAuth2 Application URL of image`,
		Optional:     true,
		ValidateFunc: validation.IsURLWithHTTPS,
	}
}

func oauth2TenantSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  `The OAuth2 Application tenant might be required for some providers like Microsoft`,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(2, 254),
	}
}

func oauth2HostedDomainSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  `The OAuth2 Application hosted domain`,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(2, 254),
	}
}

func oauth2AuthStyleSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(getMapStringKeys(oauth2AppAuthStyles), false),
		Description:  `AuthStyle represents how requests for tokens are authenticated to the server.`,
	}
}

func oauth2AppPrivateKeyPemSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  `Required if using Apple as provider. Used to sign JWT token which acts as client_secret.`,
		ValidateFunc: validation.StringMatch(pemRegex, "invalid format of PEM private key"),
	}
}

func oauth2AppPrivateKeyIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  `Required if using Apple as provider. Used to generate JWT token for authorization code.`,
		ValidateFunc: validation.StringLenBetween(2, 254),
	}
}

func oauth2AppTeamIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  `Required if using Apple as provider. Used to generate JWT token for authorization code.`,
		ValidateFunc: validation.StringLenBetween(2, 254),
	}
}
