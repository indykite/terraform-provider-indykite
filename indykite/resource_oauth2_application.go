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
	"context"
	"fmt"
	"net/mail"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	config "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
)

const (
	oauth2ApplicationClientIDKey                    = "client_id"
	oauth2ApplicationClientSecretKey                = "client_secret"
	oauth2ApplicationDisplayNameKey                 = "oauth2_application_display_name"
	oauth2ApplicationDescriptionKey                 = "oauth2_application_description"
	oauth2ApplicationRedirectUrisKey                = "redirect_uris"
	oauth2ApplicationOwnerKey                       = "owner"
	oauth2ApplicationPolicyURIKey                   = "policy_uri"
	oauth2ApplicationAllowedCorsOriginsKey          = "allowed_cors_origins"
	oauth2ApplicationTermsOfServiceURIKey           = "terms_of_service_uri"
	oauth2ApplicationClientURIKey                   = "client_uri"
	oauth2ApplicationLogoURIKey                     = "logo_uri"
	oauth2ApplicationUserSupportEmailAddressKey     = "user_support_email_address"
	oauth2ApplicationAdditionalContactsKey          = "additional_contacts"
	oauth2ApplicationSectorIdentifierURIKey         = "sector_identifier_uri"
	oauth2ApplicationAudiencesKey                   = "audiences"
	oauth2ApplicationTokenEndpointAuthMethodKey     = "token_endpoint_auth_method"
	oauth2ApplicationTokenEndpointAuthSigningAlgKey = "token_endpoint_auth_signing_alg"
	oauth2ApplicationUserinfoSignedResponseAlgKey   = "userinfo_signed_response_alg"
)

func resourceOAuth2Application() *schema.Resource {
	return &schema.Resource{
		CreateContext: resOAuth2ApplicationCreateContext,
		ReadContext:   resOAuth2ApplicationReadContext,
		UpdateContext: resOAuth2ApplicationUpdateContext,
		DeleteContext: resOAuth2ApplicationDeleteContext,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			oauth2ProviderIDKey:                             oauth2ProviderIDSchema(),
			appSpaceIDKey:                                   setComputed(appSpaceIDSchema()),
			customerIDKey:                                   setComputed(customerIDSchema()),
			nameKey:                                         nameSchema(),
			displayNameKey:                                  displayNameSchema(),
			descriptionKey:                                  descriptionSchema(),
			createTimeKey:                                   createTimeSchema(),
			updateTimeKey:                                   updateTimeSchema(),
			oauth2ApplicationClientIDKey:                    oauth2ApplicationClientIDSchema(),
			oauth2ApplicationClientSecretKey:                oauth2ApplicationClientSecretSchema(),
			oauth2ApplicationDisplayNameKey:                 oauth2ApplicationDisplayNameSchema(),
			oauth2ApplicationDescriptionKey:                 oauth2ApplicationDescriptionSchema(),
			oauth2ApplicationRedirectUrisKey:                oauth2ApplicationRedirectUrisSchema(),
			oauth2ApplicationOwnerKey:                       oauth2ApplicationOwnerSchema(),
			oauth2ApplicationPolicyURIKey:                   oauth2ApplicationPolicyURISchema(),
			oauth2ApplicationAllowedCorsOriginsKey:          oauth2ApplicationAllowedCorsOriginsSchema(),
			oauth2ApplicationTermsOfServiceURIKey:           oauth2ApplicationTermsOfServiceURISchema(),
			oauth2ApplicationClientURIKey:                   oauth2ApplicationClientURISchema(),
			oauth2ApplicationLogoURIKey:                     oauth2ApplicationLogoURISchema(),
			oauth2ApplicationUserSupportEmailAddressKey:     oauth2ApplicationUserSupportEmailAddressSchema(),
			oauth2ApplicationAdditionalContactsKey:          oauth2ApplicationAdditionalContactsSchema(),
			oauth2ClientSubjectTypeKey:                      oauth2ClientSubjectTypeSchema(),
			oauth2ApplicationSectorIdentifierURIKey:         oauth2ApplicationSectorIdentifierURISchema(),
			oauth2GrantTypeKey:                              oauth2ApplicationGrantTypeSchema(),
			oauth2ResponseTypeKey:                           oauth2ApplicationResponseTypeSchema(),
			oauth2ScopesKey:                                 oauth2ScopesSchema(),
			oauth2ApplicationAudiencesKey:                   oauth2ApplicationAudiencesSchema(),
			oauth2ApplicationTokenEndpointAuthMethodKey:     oauth2ApplicationTokenEndpointAuthMethodSchema(),
			oauth2ApplicationTokenEndpointAuthSigningAlgKey: oauth2ApplicationTokenEndpointAuthSigningAlgSchema(),
			oauth2ApplicationUserinfoSignedResponseAlgKey:   oauth2ApplicationUserinfoSignedResponseAlgSchema(),
			deletionProtectionKey:                           deletionProtectionSchema(),
		},
	}
}

func resOAuth2ApplicationCreateContext(ctx context.Context,
	data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}

	name := data.Get(nameKey).(string)
	request := &config.CreateOAuth2ApplicationRequest{
		Oauth2ProviderId: data.Get(oauth2ProviderIDKey).(string),
		Name:             name,
		DisplayName:      optionalString(data, displayNameKey),
		Description:      optionalString(data, descriptionKey),
		Config: &config.OAuth2ApplicationConfig{
			DisplayName:             data.Get(oauth2ApplicationDisplayNameKey).(string),
			Description:             data.Get(oauth2ApplicationDescriptionKey).(string),
			RedirectUris:            rawArrayToStringArray(data.Get(oauth2ApplicationRedirectUrisKey)),
			Owner:                   data.Get(oauth2ApplicationOwnerKey).(string),
			PolicyUri:               data.Get(oauth2ApplicationPolicyURIKey).(string),
			AllowedCorsOrigins:      rawArrayToStringArray(data.Get(oauth2ApplicationAllowedCorsOriginsKey)),
			TermsOfServiceUri:       data.Get(oauth2ApplicationTermsOfServiceURIKey).(string),
			ClientUri:               data.Get(oauth2ApplicationClientURIKey).(string),
			LogoUri:                 data.Get(oauth2ApplicationLogoURIKey).(string),
			UserSupportEmailAddress: data.Get(oauth2ApplicationUserSupportEmailAddressKey).(string),
			AdditionalContacts:      rawArrayToStringArray(data.Get(oauth2ApplicationAdditionalContactsKey)),
			SubjectType:             OAuth2ClientSubjectTypes[data.Get(oauth2ClientSubjectTypeKey).(string)],
			SectorIdentifierUri:     data.Get(oauth2ApplicationSectorIdentifierURIKey).(string),
			GrantTypes:              rawArrayToGrantTypeArray(&d, data.Get(oauth2GrantTypeKey)),
			ResponseTypes:           rawArrayToResponseTypeArray(&d, data.Get(oauth2ResponseTypeKey)),
			Scopes:                  rawArrayToStringArray(data.Get(oauth2ScopesKey)),
			Audiences:               rawArrayToStringArray(data.Get(oauth2ApplicationAudiencesKey)),
			//nolint:lll
			TokenEndpointAuthMethod:     OAuth2TokenEndpointAuthMethods[data.Get(oauth2ApplicationTokenEndpointAuthMethodKey).(string)],
			TokenEndpointAuthSigningAlg: data.Get(oauth2ApplicationTokenEndpointAuthSigningAlgKey).(string),
			UserinfoSignedResponseAlg:   data.Get(oauth2ApplicationUserinfoSignedResponseAlgKey).(string),
		},
	}

	resp, err := client.getClient().CreateOAuth2Application(ctx, request)
	if hasFailed(&d, err) {
		return d
	}
	data.SetId(resp.Id)
	setData(&d, data, oauth2ApplicationClientSecretKey, resp.ClientSecret)

	if d.HasError() {
		return d
	}

	return resOAuth2ApplicationReadContext(ctx, data, meta)
}

func resOAuth2ApplicationReadContext(ctx context.Context,
	data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	resp, err := client.getClient().ReadOAuth2Application(ctx, &config.ReadOAuth2ApplicationRequest{
		Id: data.Id(),
	})
	if hasFailed(&d, err) {
		return d
	}
	if resp.GetOauth2Application().GetConfig() == nil {
		return diag.Diagnostics{buildPluginError("empty OAuth2Application response")}
	}

	data.SetId(resp.Oauth2Application.Id)
	setData(&d, data, nameKey, resp.Oauth2Application.Name)
	setData(&d, data, displayNameKey, resp.Oauth2Application.DisplayName)
	setData(&d, data, descriptionKey, resp.Oauth2Application.Description)
	setData(&d, data, createTimeKey, resp.Oauth2Application.CreateTime)
	setData(&d, data, updateTimeKey, resp.Oauth2Application.UpdateTime)
	setData(&d, data, customerIDKey, resp.Oauth2Application.CustomerId)
	setData(&d, data, appSpaceIDKey, resp.Oauth2Application.AppSpaceId)
	setData(&d, data, oauth2ProviderIDKey, resp.Oauth2Application.Oauth2ProviderId)
	setData(&d, data, oauth2ApplicationClientIDKey, resp.Oauth2Application.Config.ClientId)
	setData(&d, data, oauth2ApplicationDisplayNameKey, resp.Oauth2Application.Config.DisplayName)
	setData(&d, data, oauth2ApplicationDescriptionKey, resp.Oauth2Application.Config.Description)
	setData(&d, data, oauth2ApplicationRedirectUrisKey, resp.Oauth2Application.Config.RedirectUris)
	setData(&d, data, oauth2ApplicationOwnerKey, resp.Oauth2Application.Config.Owner)
	setData(&d, data, oauth2ApplicationPolicyURIKey, resp.Oauth2Application.Config.PolicyUri)
	setData(&d, data, oauth2ApplicationAllowedCorsOriginsKey, resp.Oauth2Application.Config.AllowedCorsOrigins)
	setData(&d, data, oauth2ApplicationTermsOfServiceURIKey, resp.Oauth2Application.Config.TermsOfServiceUri)
	setData(&d, data, oauth2ApplicationClientURIKey, resp.Oauth2Application.Config.ClientUri)
	setData(&d, data, oauth2ApplicationLogoURIKey, resp.Oauth2Application.Config.LogoUri)
	setData(&d, data, oauth2ApplicationUserSupportEmailAddressKey,
		resp.Oauth2Application.Config.UserSupportEmailAddress)
	setData(&d, data, oauth2ApplicationAdditionalContactsKey, resp.Oauth2Application.Config.AdditionalContacts)

	subjectType, exists := OAuth2ClientSubjectTypesReverse[resp.Oauth2Application.Config.SubjectType]
	if !exists {
		d = append(d, buildPluginError(
			"BE send unsupported OAuth2 SubjectType: "+resp.Oauth2Application.Config.SubjectType.String(),
		))
	}
	setData(&d, data, oauth2ClientSubjectTypeKey, subjectType)

	setData(&d, data, oauth2ApplicationSectorIdentifierURIKey, resp.Oauth2Application.Config.SectorIdentifierUri)
	setData(&d, data, oauth2GrantTypeKey, grantTypeArrayToRawArray(&d, resp.Oauth2Application.Config.GrantTypes))
	setData(&d, data, oauth2ResponseTypeKey,
		responseTypeArrayToRawArray(&d, resp.Oauth2Application.Config.ResponseTypes))
	setData(&d, data, oauth2ScopesKey, resp.Oauth2Application.Config.Scopes)
	setData(&d, data, oauth2ApplicationAudiencesKey, resp.Oauth2Application.Config.Audiences)

	tokenEPAuthMethod := resp.Oauth2Application.Config.TokenEndpointAuthMethod
	tokenEPAuthMethodString, exists := OAuth2TokenEndpointAuthMethodsReverse[tokenEPAuthMethod]
	if !exists {
		d = append(d, buildPluginError(
			"BE send unsupported OAuth2 TokenEndpointAuthMethod: "+tokenEPAuthMethod.String(),
		))
	}
	setData(&d, data, oauth2ApplicationTokenEndpointAuthMethodKey, tokenEPAuthMethodString)

	setData(&d, data, oauth2ApplicationTokenEndpointAuthSigningAlgKey,
		resp.Oauth2Application.Config.TokenEndpointAuthSigningAlg)
	setData(&d, data, oauth2ApplicationUserinfoSignedResponseAlgKey,
		resp.Oauth2Application.Config.UserinfoSignedResponseAlg)
	return d
}

func resOAuth2ApplicationUpdateContext(ctx context.Context,
	data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}

	// If only change in plan is delete_protection, just ignore the request
	if !data.HasChangeExcept(deletionProtectionKey) {
		return d
	}

	req := &config.UpdateOAuth2ApplicationRequest{
		Id:          data.Id(),
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
		Config: &config.OAuth2ApplicationConfig{
			DisplayName:             data.Get(oauth2ApplicationDisplayNameKey).(string),
			Description:             data.Get(oauth2ApplicationDescriptionKey).(string),
			RedirectUris:            rawArrayToStringArray(data.Get(oauth2ApplicationRedirectUrisKey)),
			Owner:                   data.Get(oauth2ApplicationOwnerKey).(string),
			PolicyUri:               data.Get(oauth2ApplicationPolicyURIKey).(string),
			AllowedCorsOrigins:      rawArrayToStringArray(data.Get(oauth2ApplicationAllowedCorsOriginsKey)),
			TermsOfServiceUri:       data.Get(oauth2ApplicationTermsOfServiceURIKey).(string),
			ClientUri:               data.Get(oauth2ApplicationClientURIKey).(string),
			LogoUri:                 data.Get(oauth2ApplicationLogoURIKey).(string),
			UserSupportEmailAddress: data.Get(oauth2ApplicationUserSupportEmailAddressKey).(string),
			AdditionalContacts:      rawArrayToStringArray(data.Get(oauth2ApplicationAdditionalContactsKey)),
			SubjectType:             OAuth2ClientSubjectTypes[data.Get(oauth2ClientSubjectTypeKey).(string)],
			SectorIdentifierUri:     data.Get(oauth2ApplicationSectorIdentifierURIKey).(string),
			GrantTypes:              rawArrayToGrantTypeArray(&d, data.Get(oauth2GrantTypeKey)),
			ResponseTypes:           rawArrayToResponseTypeArray(&d, data.Get(oauth2ResponseTypeKey)),
			Scopes:                  rawArrayToStringArray(data.Get(oauth2ScopesKey)),
			Audiences:               rawArrayToStringArray(data.Get(oauth2ApplicationAudiencesKey)),
			//nolint:lll
			TokenEndpointAuthMethod:     OAuth2TokenEndpointAuthMethods[data.Get(oauth2ApplicationTokenEndpointAuthMethodKey).(string)],
			TokenEndpointAuthSigningAlg: data.Get(oauth2ApplicationTokenEndpointAuthSigningAlgKey).(string),
			UserinfoSignedResponseAlg:   data.Get(oauth2ApplicationUserinfoSignedResponseAlgKey).(string),
		},
	}

	_, err := client.getClient().UpdateOAuth2Application(ctx, req)
	if hasFailed(&d, err) {
		return d
	}
	return resOAuth2ApplicationReadContext(ctx, data, meta)
}

func resOAuth2ApplicationDeleteContext(ctx context.Context,
	data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	if hasDeleteProtection(&d, data) {
		return d
	}
	_, err := client.getClient().DeleteOAuth2Application(ctx, &config.DeleteOAuth2ApplicationRequest{
		Id: data.Id(),
	})
	hasFailed(&d, err)
	return d
}

func oauth2ApplicationClientIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: `The oauth2 application client_id`,
	}
}

func oauth2ApplicationClientSecretSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
		Description: `The oauth2 application client_secret`,
	}
}

func oauth2ApplicationDisplayNameSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(2, 254),
		Description:  `The oauth2 application display_name`,
	}
}

func oauth2ApplicationDescriptionSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(2, 254),
		Description:  `The oauth2 application description`,
	}
}

func oauth2ApplicationRedirectUrisSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.IsURLWithScheme([]string{"https"}),
			Description:  `The oauth2 application redirect_uris`,
		},
	}
}

func oauth2ApplicationOwnerSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(2, 254),
		Description:  `The oauth2 application owner`,
	}
}

func oauth2ApplicationPolicyURISchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: validation.All(
			validation.IsURLWithScheme([]string{"https"}),
			validation.StringLenBetween(1, 254),
		),
		Description: `The oauth2 application policy_uri`,
	}
}

func oauth2ApplicationAllowedCorsOriginsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		Elem: &schema.Schema{
			Type: schema.TypeString,
			ValidateFunc: validation.All(
				validation.IsURLWithScheme([]string{"https"}),
				validation.StringLenBetween(1, 254),
			),
			Description: `The oauth2 application redirect_uris`,
		},
	}
}

func oauth2ApplicationTermsOfServiceURISchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: validation.All(
			validation.IsURLWithScheme([]string{"https"}),
			validation.StringLenBetween(1, 254),
		),
		Description: `The oauth2 application terms_of_service_uri`,
	}
}

func oauth2ApplicationClientURISchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: validation.All(
			validation.IsURLWithScheme([]string{"https"}),
			validation.StringLenBetween(1, 254),
		),
		Description: `The oauth2 application client_uri`,
	}
}

func oauth2ApplicationLogoURISchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: validation.All(
			validation.IsURLWithScheme([]string{"https"}),
			validation.StringLenBetween(1, 254),
		),
		Description: `The oauth2 application logo_uri`,
	}
}

func oauth2ApplicationUserSupportEmailAddressSchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validateApplicationUserSupportEmailAddress("user_support_email_address"),
		Description:      `The oauth2 application user_support_email_address`,
	}
}

func oauth2ApplicationAdditionalContactsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type:        schema.TypeString,
			Description: `The oauth2 application additional_contacts`,
		},
	}
}

func oauth2ClientSubjectTypeSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(getMapStringKeys(OAuth2ClientSubjectTypes), false),
		Description:  `The oauth2 client_subject_type`,
	}
}

func oauth2ApplicationSectorIdentifierURISchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		ValidateFunc: validation.All(
			validation.IsURLWithScheme([]string{"https"}),
			validation.StringLenBetween(1, 254),
		),
		Description: `The oauth2 application sector_identifier_uri`,
	}
}

func oauth2ApplicationGrantTypeSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(getMapStringKeys(OAuth2GrantTypes), false),
			Description:  `The oauth2 grant_type`,
		},
	}
}

func oauth2ApplicationResponseTypeSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(getMapStringKeys(OAuth2ResponseTypes), false),
			Description:  `The oauth2 response_type`,
		},
	}
}

func oauth2ApplicationAudiencesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type:             schema.TypeString,
			ValidateDiagFunc: validateIsUUID("Audiences"),
			Description:      `The oauth2 application audiences`,
		},
	}
}

func oauth2ApplicationTokenEndpointAuthMethodSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(getMapStringKeys(OAuth2TokenEndpointAuthMethods), false),
		Description:  `The oauth2 application token_endpoint_auth_method`,
	}
}
func oauth2ApplicationTokenEndpointAuthSigningAlgSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(supportedSigningAlgs, false),
		Description:  `The oauth2 application token_endpoint_auth_signing_alg`,
	}
}

func oauth2ApplicationUserinfoSignedResponseAlgSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice([]string{"RS256"}, false),
		Description:  `The oauth2 application userinfo_signed_response_alg`,
	}
}

func validateIsUUID(key string) schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		_, errors := validation.IsUUID(v, key)

		for _, err := range errors {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "IsUUID validation failed",
				Detail:        err.Error(),
				AttributePath: append(path, cty.IndexStep{Key: cty.StringVal("Audiences")}),
			})
		}

		return diags
	}
}

func validateApplicationUserSupportEmailAddress(key string) schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		v, ok := i.(string)
		if !ok {
			return append(diags, buildPluginErrorWithAttrName(
				fmt.Sprintf("validateApplicationUserSupportEmailAddress failed, expected type of %s to be string", key),
				key,
			))
		}

		if _, err := mail.ParseAddress(v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Email validation failed",
				Detail:        err.Error(),
				AttributePath: append(path, cty.IndexStep{Key: cty.StringVal(key)}),
			})
		}

		stringLenBetweenDiagFunc := validation.ToDiagFunc(validation.StringLenBetween(1, 254))
		res := stringLenBetweenDiagFunc(v, cty.GetAttrPath(key))
		diags = append(diags, res...)

		return diags
	}
}
