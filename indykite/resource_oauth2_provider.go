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

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	configpb "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
)

const (
	oauth2ProviderRequestUrisKey                 = "request_uris"
	oauth2ProviderRequestObjectSigningAlgKey     = "request_object_signing_alg"
	oauth2ProviderFrontChannelLoginURIKey        = "front_channel_login_uri"
	oauth2ProviderFrontChannelConsentURIKey      = "front_channel_consent_uri"
	oauth2ProviderTokenEndpointAuthMethodKey     = "token_endpoint_auth_method"
	oauth2ProviderTokenEndpointAuthSigningAlgKey = "token_endpoint_auth_signing_alg"
)

func resourceOAuth2Provider() *schema.Resource {
	return &schema.Resource{
		CreateContext: resOAuth2ProviderCreateContext,
		ReadContext:   resOAuth2ProviderReadContext,
		UpdateContext: resOAuth2ProviderUpdateContext,
		DeleteContext: resOAuth2ProviderDeleteContext,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			appSpaceIDKey:                            appSpaceIDSchema(),
			customerIDKey:                            setComputed(customerIDSchema()),
			nameKey:                                  nameSchema(),
			displayNameKey:                           displayNameSchema(),
			descriptionKey:                           descriptionSchema(),
			createTimeKey:                            createTimeSchema(),
			updateTimeKey:                            updateTimeSchema(),
			oauth2GrantTypeKey:                       oauth2ProviderGrantTypeSchema(),
			oauth2ResponseTypeKey:                    oauth2ProviderResponseTypeSchema(),
			oauth2ScopesKey:                          oauth2ScopesSchema(),
			oauth2ProviderTokenEndpointAuthMethodKey: oauth2ProviderTokenEndpointAuthMethodSchema(),
			oauth2ProviderTokenEndpointAuthSigningAlgKey: oauth2ProviderTokenEndpointAuthSigningAlgSchema(),
			oauth2ProviderFrontChannelLoginURIKey:        frontChannelLoginURISchema(),
			oauth2ProviderFrontChannelConsentURIKey:      frontChannelConsentURISchema(),
			oauth2ProviderRequestUrisKey:                 requestUrisSchema(),
			oauth2ProviderRequestObjectSigningAlgKey:     requestObjectSigningAlgSchema(),
			deletionProtectionKey:                        deletionProtectionSchema(),
		},
	}
}

func resOAuth2ProviderCreateContext(ctx context.Context,
	data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}

	name := data.Get(nameKey).(string)
	request := &configpb.CreateOAuth2ProviderRequest{
		AppSpaceId:  data.Get(appSpaceIDKey).(string),
		Name:        name,
		DisplayName: optionalString(data, displayNameKey),
		Description: optionalString(data, descriptionKey),
		Config: &configpb.OAuth2ProviderConfig{
			GrantTypes:    rawSetToGrantTypeArray(data.Get(oauth2GrantTypeKey)),
			ResponseTypes: rawSetToResponseTypeArray(data.Get(oauth2ResponseTypeKey)),
			Scopes:        rawSetToStringArray(data.Get(oauth2ScopesKey)),
			TokenEndpointAuthMethod: rawSetToTokenEndpointAuthMethodArray(
				data.Get(oauth2ProviderTokenEndpointAuthMethodKey)),
			TokenEndpointAuthSigningAlg: rawSetToStringArray(data.Get(oauth2ProviderTokenEndpointAuthSigningAlgKey)),
			RequestUris:                 rawSetToStringArray(data.Get(oauth2ProviderRequestUrisKey)),
			RequestObjectSigningAlg:     data.Get(oauth2ProviderRequestObjectSigningAlgKey).(string),
			FrontChannelLoginUri:        rawMapToStringMap(data.Get(oauth2ProviderFrontChannelLoginURIKey)),
			FrontChannelConsentUri:      rawMapToStringMap(data.Get(oauth2ProviderFrontChannelConsentURIKey)),
		},
	}

	resp, err := client.Client().CreateOAuth2Provider(ctx, request)
	if hasFailed(&d, err, "error creating oauth2 provider for %q", name) {
		return d
	}
	data.SetId(resp.Id)

	return resOAuth2ProviderReadContext(ctx, data, meta)
}

func resOAuth2ProviderReadContext(ctx context.Context,
	data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	resp, err := client.Client().ReadOAuth2Provider(ctx, &configpb.ReadOAuth2ProviderRequest{
		Id: data.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if resp == nil {
		return diag.Errorf("empty OAuth2Provider response")
	}

	data.SetId(resp.Oauth2Provider.Id)
	Set(&d, data, nameKey, resp.Oauth2Provider.Name)
	Set(&d, data, displayNameKey, resp.Oauth2Provider.DisplayName)
	Set(&d, data, descriptionKey, resp.Oauth2Provider.Description)
	Set(&d, data, createTimeKey, resp.Oauth2Provider.CreateTime)
	Set(&d, data, updateTimeKey, resp.Oauth2Provider.UpdateTime)
	Set(&d, data, customerIDKey, resp.Oauth2Provider.CustomerId)
	Set(&d, data, appSpaceIDKey, resp.Oauth2Provider.AppSpaceId)
	Set(&d, data, oauth2GrantTypeKey, grantTypeArrayToRawArray(resp.Oauth2Provider.Config.GrantTypes))
	Set(&d, data, oauth2ResponseTypeKey, responseTypeArrayToRawArray(resp.Oauth2Provider.Config.ResponseTypes))
	Set(&d, data, oauth2ScopesKey, resp.Oauth2Provider.Config.Scopes)
	Set(&d, data, oauth2ProviderTokenEndpointAuthMethodKey,
		tokenEndpointAuthMethodArrayToRawArray(resp.Oauth2Provider.Config.TokenEndpointAuthMethod))
	Set(&d, data, oauth2ProviderTokenEndpointAuthSigningAlgKey, resp.Oauth2Provider.Config.TokenEndpointAuthSigningAlg)
	Set(&d, data, oauth2ProviderRequestUrisKey, resp.Oauth2Provider.Config.RequestUris)
	Set(&d, data, oauth2ProviderRequestObjectSigningAlgKey, resp.Oauth2Provider.Config.RequestObjectSigningAlg)
	Set(&d, data, oauth2ProviderFrontChannelLoginURIKey, resp.Oauth2Provider.Config.FrontChannelLoginUri)
	Set(&d, data, oauth2ProviderFrontChannelConsentURIKey, resp.Oauth2Provider.Config.FrontChannelConsentUri)
	return d
}

func resOAuth2ProviderUpdateContext(ctx context.Context,
	data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)

	// If only change in plan is delete_protection, just ignore the request
	if !data.HasChangeExcept(deletionProtectionKey) {
		return d
	}

	req := &configpb.UpdateOAuth2ProviderRequest{
		Id:          data.Id(),
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
		Config: &configpb.OAuth2ProviderConfig{
			GrantTypes:    rawSetToGrantTypeArray(data.Get(oauth2GrantTypeKey)),
			ResponseTypes: rawSetToResponseTypeArray(data.Get(oauth2ResponseTypeKey)),
			Scopes:        rawSetToStringArray(data.Get(oauth2ScopesKey)),
			TokenEndpointAuthMethod: rawSetToTokenEndpointAuthMethodArray(
				data.Get(oauth2ProviderTokenEndpointAuthMethodKey)),
			TokenEndpointAuthSigningAlg: rawSetToStringArray(data.Get(oauth2ProviderTokenEndpointAuthSigningAlgKey)),
			RequestUris:                 rawSetToStringArray(data.Get(oauth2ProviderRequestUrisKey)),
			RequestObjectSigningAlg:     data.Get(oauth2ProviderRequestObjectSigningAlgKey).(string),
			FrontChannelLoginUri:        rawMapToStringMap(data.Get(oauth2ProviderFrontChannelLoginURIKey)),
			FrontChannelConsentUri:      rawMapToStringMap(data.Get(oauth2ProviderFrontChannelConsentURIKey)),
		},
	}

	if client == nil {
		return d
	}
	_, err := client.Client().UpdateOAuth2Provider(ctx, req)
	if hasFailed(&d, err, "Error while updating oauth2 provider #%s", data.Id()) {
		return d
	}
	return resOAuth2ProviderReadContext(ctx, data, meta)
}

func resOAuth2ProviderDeleteContext(ctx context.Context,
	data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	if hasDeleteProtection(&d, data) {
		return d
	}
	_, err := client.Client().DeleteOAuth2Provider(ctx, &configpb.DeleteOAuth2ProviderRequest{
		Id: data.Id(),
	})
	if hasFailed(&d, err, "Error while deleting oauth2 provider #%s", data.Id()) {
		return d
	}
	return d
}

func oauth2ProviderGrantTypeSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		MinItems: 1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(getMapStringKeys(OAuth2GrantTypes), false),
			Description:  `The oauth2 grant_type`,
		},
	}
}

func oauth2ProviderResponseTypeSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		MinItems: 1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(getMapStringKeys(OAuth2ResponseTypes), false),
			Description:  `The oauth2 response_type`,
		},
	}
}

func requestObjectSigningAlgSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(supportedSigningAlgs, false),
		Description:  `The oauth2 provider request_object_signing_alg`,
	}
}

func requestUrisSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.IsURLWithScheme([]string{"https"}),
			Description:  `The oauth2 provider request_uris`,
		},
	}
}

func frontChannelConsentURISchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeMap,
		Required:         true,
		ValidateDiagFunc: validateMapOfURIs(1, 32),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func frontChannelLoginURISchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeMap,
		Required:         true,
		ValidateDiagFunc: validateMapOfURIs(1, 32),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func oauth2ProviderTokenEndpointAuthMethodSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		MinItems: 1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(getMapStringKeys(OAuth2TokenEndpointAuthMethods), false),
			Description:  `The oauth2 provider token_endpoint_auth_method`,
		},
	}
}
func oauth2ProviderTokenEndpointAuthSigningAlgSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		MinItems: 1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(supportedSigningAlgs, false),
			Description:  `The oauth2 provider token_endpoint_auth_signing_alg`,
		},
	}
}

func validateMapOfURIs(keyMinLength int, keyMaxLength int) schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics

		inputMap, _ := v.(map[string]interface{})
		if len(inputMap) == 0 {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Map item required",
				Detail:        fmt.Sprintf("provided map %v must have items", v),
				AttributePath: append(path, cty.IndexStep{Key: cty.StringVal("Map")}),
			})
		}

		isURLWithSchemeDiagFunc := validation.ToDiagFunc(validation.IsURLWithScheme([]string{"https"}))
		for key, value := range inputMap {
			res := isURLWithSchemeDiagFunc(value, cty.GetAttrPath(key))
			diags = append(diags, res...)
		}

		mapKeyLenBetweenDiagFunc := validation.MapKeyLenBetween(keyMinLength, keyMaxLength)
		res := mapKeyLenBetweenDiagFunc(v, path)
		diags = append(diags, res...)

		return diags
	}
}

func rawSetToGrantTypeArray(rawData interface{}) []configpb.GrantType {
	grantTypeSet, ok := rawData.(*schema.Set)
	if !ok {
		return nil
	}

	grantTypeList := grantTypeSet.List()

	arr := make([]configpb.GrantType, len(grantTypeList))
	if len(arr) == 0 {
		return nil
	}
	for i, val := range grantTypeList {
		grantType := OAuth2GrantTypes[val.(string)]
		arr[i] = grantType
	}

	return arr
}

func grantTypeArrayToRawArray(grantTypes []configpb.GrantType) []string {
	arr := make([]string, len(grantTypes))
	if len(arr) == 0 {
		return nil
	}
	oauth2GrantTypesReverse := reverseProtoEnumMap(OAuth2GrantTypes)
	for i, val := range grantTypes {
		grantType := oauth2GrantTypesReverse[val]
		arr[i] = grantType
	}

	return arr
}

func rawSetToResponseTypeArray(rawData interface{}) []configpb.ResponseType {
	responseTypeSet, ok := rawData.(*schema.Set)
	if !ok {
		return nil
	}

	responseTypeList := responseTypeSet.List()

	arr := make([]configpb.ResponseType, len(responseTypeList))
	if len(arr) == 0 {
		return nil
	}
	for i, val := range responseTypeList {
		reponseType := OAuth2ResponseTypes[val.(string)]
		arr[i] = reponseType
	}

	return arr
}

func responseTypeArrayToRawArray(responseTypes []configpb.ResponseType) []string {
	arr := make([]string, len(responseTypes))
	if len(arr) == 0 {
		return nil
	}
	oauth2ResponseTypesReverse := reverseProtoEnumMap(OAuth2ResponseTypes)
	for i, val := range responseTypes {
		grantType := oauth2ResponseTypesReverse[val]
		arr[i] = grantType
	}

	return arr
}

func rawSetToTokenEndpointAuthMethodArray(rawData interface{}) []configpb.TokenEndpointAuthMethod {
	responseTypeSet, ok := rawData.(*schema.Set)
	if !ok {
		return nil
	}

	responseTypeList := responseTypeSet.List()

	arr := make([]configpb.TokenEndpointAuthMethod, len(responseTypeList))
	if len(arr) == 0 {
		return nil
	}
	for i, val := range responseTypeList {
		authMethod := OAuth2TokenEndpointAuthMethods[val.(string)]
		arr[i] = authMethod
	}

	return arr
}

func tokenEndpointAuthMethodArrayToRawArray(
	tokenEndpointAuthMethods []configpb.TokenEndpointAuthMethod) []string {
	arr := make([]string, len(tokenEndpointAuthMethods))
	if len(arr) == 0 {
		return nil
	}
	for i, val := range tokenEndpointAuthMethods {
		grantType := oauth2TokenEndpointAuthMethodsReverse[val]
		arr[i] = grantType
	}

	return arr
}
