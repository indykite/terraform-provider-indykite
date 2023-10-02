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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

func dataSourceOAuth2Provider() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataOAuth2ProviderReadContext,
		Schema: map[string]*schema.Schema{
			oauth2ProviderIDKey:                      oauth2ProviderIDSchema(),
			appSpaceIDKey:                            setComputed(appSpaceIDSchema()),
			customerIDKey:                            setComputed(customerIDSchema()),
			nameKey:                                  setComputed(nameSchema()),
			displayNameKey:                           setComputed(displayNameSchema()),
			descriptionKey:                           setComputed(descriptionSchema()),
			createTimeKey:                            setComputed(createTimeSchema()),
			updateTimeKey:                            setComputed(updateTimeSchema()),
			oauth2GrantTypeKey:                       setComputed(oauth2ProviderGrantTypeSchema()),
			oauth2ResponseTypeKey:                    setComputed(oauth2ProviderResponseTypeSchema()),
			oauth2ScopesKey:                          setComputed(oauth2ScopesSchema()),
			oauth2ProviderTokenEndpointAuthMethodKey: setComputed(oauth2ProviderTokenEndpointAuthMethodSchema()),
			oauth2ProviderTokenEndpointAuthSigningAlgKey: setComputed(
				oauth2ProviderTokenEndpointAuthSigningAlgSchema()),
			oauth2ProviderFrontChannelLoginURIKey:    setComputed(frontChannelLoginURISchema()),
			oauth2ProviderFrontChannelConsentURIKey:  setComputed(frontChannelConsentURISchema()),
			oauth2ProviderRequestUrisKey:             setComputed(requestUrisSchema()),
			oauth2ProviderRequestObjectSigningAlgKey: setComputed(requestObjectSigningAlgSchema()),
			deletionProtectionKey:                    deletionProtectionSchema(),
		},
		Timeouts: defaultDataTimeouts(),
	}
}

func dataOAuth2ProviderReadContext(ctx context.Context,
	data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadOAuth2Provider(ctx, &configpb.ReadOAuth2ProviderRequest{
		Id:        data.Get(oauth2ProviderIDKey).(string),
		Bookmarks: clientCtx.GetBookmarks(),
	})
	if hasFailed(&d, err) {
		return d
	}

	return dataOAuth2ProviderFlatten(data, resp.GetOauth2Provider())
}

func dataOAuth2ProviderFlatten(data *schema.ResourceData, resp *configpb.OAuth2Provider) diag.Diagnostics {
	var d diag.Diagnostics
	if resp.GetConfig() == nil {
		return diag.Diagnostics{buildPluginError("empty OAuth2Provider response")}
	}

	data.SetId(resp.Id)
	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, customerIDKey, resp.CustomerId)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceId)
	setData(&d, data, oauth2GrantTypeKey, grantTypeArrayToRawArray(&d, resp.Config.GrantTypes))
	setData(&d, data, oauth2ResponseTypeKey, responseTypeArrayToRawArray(&d, resp.Config.ResponseTypes))
	setData(&d, data, oauth2ScopesKey, resp.Config.Scopes)
	setData(&d, data, oauth2ProviderTokenEndpointAuthMethodKey,
		tokenEndpointAuthMethodArrayToRawArray(&d, resp.Config.TokenEndpointAuthMethod))
	setData(&d, data, oauth2ProviderTokenEndpointAuthSigningAlgKey, resp.Config.TokenEndpointAuthSigningAlg)
	setData(&d, data, oauth2ProviderRequestUrisKey, resp.Config.RequestUris)
	setData(&d, data, oauth2ProviderRequestObjectSigningAlgKey, resp.Config.RequestObjectSigningAlg)
	setData(&d, data, oauth2ProviderFrontChannelLoginURIKey, resp.Config.FrontChannelLoginUri)
	setData(&d, data, oauth2ProviderFrontChannelConsentURIKey, resp.Config.FrontChannelConsentUri)
	return d
}
