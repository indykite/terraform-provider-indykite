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
	configpb "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
)

func dataSourceOAuth2Application() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataOAuth2ApplicationReadContext,
		Schema: map[string]*schema.Schema{
			oauth2ApplicationIDKey:                      oauth2ApplicationIDSchema(),
			oauth2ProviderIDKey:                         setComputed(oauth2ProviderIDSchema()),
			appSpaceIDKey:                               setComputed(appSpaceIDSchema()),
			customerIDKey:                               setComputed(customerIDSchema()),
			nameKey:                                     setComputed(nameSchema()),
			displayNameKey:                              setComputed(displayNameSchema()),
			descriptionKey:                              setComputed(descriptionSchema()),
			createTimeKey:                               setComputed(createTimeSchema()),
			updateTimeKey:                               setComputed(updateTimeSchema()),
			oauth2ApplicationClientIDKey:                setComputed(oauth2ApplicationClientIDSchema()),
			oauth2ApplicationDisplayNameKey:             setComputed(oauth2ApplicationDisplayNameSchema()),
			oauth2ApplicationDescriptionKey:             setComputed(oauth2ApplicationDescriptionSchema()),
			oauth2ApplicationRedirectUrisKey:            setComputed(oauth2ApplicationRedirectUrisSchema()),
			oauth2ApplicationOwnerKey:                   setComputed(oauth2ApplicationOwnerSchema()),
			oauth2ApplicationPolicyURIKey:               setComputed(oauth2ApplicationPolicyURISchema()),
			oauth2ApplicationAllowedCorsOriginsKey:      setComputed(oauth2ApplicationAllowedCorsOriginsSchema()),
			oauth2ApplicationTermsOfServiceURIKey:       setComputed(oauth2ApplicationTermsOfServiceURISchema()),
			oauth2ApplicationClientURIKey:               setComputed(oauth2ApplicationClientURISchema()),
			oauth2ApplicationLogoURIKey:                 setComputed(oauth2ApplicationLogoURISchema()),
			oauth2ApplicationUserSupportEmailAddressKey: setComputed(oauth2ApplicationUserSupportEmailAddressSchema()),
			oauth2ApplicationAdditionalContactsKey:      setComputed(oauth2ApplicationAdditionalContactsSchema()),
			oauth2ClientSubjectTypeKey:                  setComputed(oauth2ClientSubjectTypeSchema()),
			oauth2ApplicationSectorIdentifierURIKey:     setComputed(oauth2ApplicationSectorIdentifierURISchema()),
			oauth2GrantTypeKey:                          setComputed(oauth2ApplicationGrantTypeSchema()),
			oauth2ResponseTypeKey:                       setComputed(oauth2ApplicationResponseTypeSchema()),
			oauth2ScopesKey:                             setComputed(oauth2ScopesSchema()),
			oauth2ApplicationAudiencesKey:               setComputed(oauth2ApplicationAudiencesSchema()),
			oauth2ApplicationTokenEndpointAuthMethodKey: setComputed(oauth2ApplicationTokenEndpointAuthMethodSchema()),
			oauth2ApplicationTokenEndpointAuthSigningAlgKey: setComputed(
				oauth2ApplicationTokenEndpointAuthSigningAlgSchema()),
			oauth2ApplicationUserinfoSignedResponseAlgKey: setComputed(
				oauth2ApplicationUserinfoSignedResponseAlgSchema()),
		},
		Timeouts: defaultDataTimeouts(),
	}
}

func dataOAuth2ApplicationReadContext(ctx context.Context,
	data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if d.HasError() {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := client.getClient().ReadOAuth2Application(ctx, &configpb.ReadOAuth2ApplicationRequest{
		Id: data.Get(oauth2ApplicationIDKey).(string),
	})
	if hasFailed(&d, err) {
		return d
	}

	return dataOAuth2ApplicationFlatten(data, resp.GetOauth2Application())
}

func dataOAuth2ApplicationFlatten(data *schema.ResourceData, resp *configpb.OAuth2Application) (d diag.Diagnostics) {
	if resp.GetConfig() == nil {
		return diag.Diagnostics{buildPluginError("empty OAuth2Application response")}
	}

	data.SetId(resp.Id)
	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, customerIDKey, resp.CustomerId)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceId)
	setData(&d, data, oauth2ProviderIDKey, resp.Oauth2ProviderId)

	setData(&d, data, oauth2ApplicationClientIDKey, resp.Config.ClientId)
	setData(&d, data, oauth2ApplicationDisplayNameKey, resp.Config.DisplayName)
	setData(&d, data, oauth2ApplicationDescriptionKey, resp.Config.Description)
	setData(&d, data, oauth2ApplicationRedirectUrisKey, resp.Config.RedirectUris)
	setData(&d, data, oauth2ApplicationOwnerKey, resp.Config.Owner)
	setData(&d, data, oauth2ApplicationPolicyURIKey, resp.Config.PolicyUri)
	setData(&d, data, oauth2ApplicationAllowedCorsOriginsKey, resp.Config.AllowedCorsOrigins)
	setData(&d, data, oauth2ApplicationTermsOfServiceURIKey, resp.Config.TermsOfServiceUri)
	setData(&d, data, oauth2ApplicationClientURIKey, resp.Config.ClientUri)
	setData(&d, data, oauth2ApplicationLogoURIKey, resp.Config.LogoUri)
	setData(&d, data, oauth2ApplicationUserSupportEmailAddressKey, resp.Config.UserSupportEmailAddress)
	setData(&d, data, oauth2ApplicationAdditionalContactsKey, resp.Config.AdditionalContacts)

	subjectType, exists := OAuth2ClientSubjectTypesReverse[resp.Config.SubjectType]
	if !exists {
		d = append(d, buildPluginError("BE send unsupported OAuth2 SubjectType: "+resp.Config.SubjectType.String()))
	}
	setData(&d, data, oauth2ClientSubjectTypeKey, subjectType)
	setData(&d, data, oauth2ApplicationSectorIdentifierURIKey, resp.Config.SectorIdentifierUri)
	setData(&d, data, oauth2GrantTypeKey, grantTypeArrayToRawArray(&d, resp.Config.GrantTypes))
	setData(&d, data, oauth2ResponseTypeKey, responseTypeArrayToRawArray(&d, resp.Config.ResponseTypes))
	setData(&d, data, oauth2ScopesKey, resp.Config.Scopes)
	setData(&d, data, oauth2ApplicationAudiencesKey, resp.Config.Audiences)

	tokenEPAuthMethod, exists := OAuth2TokenEndpointAuthMethodsReverse[resp.Config.TokenEndpointAuthMethod]
	if !exists {
		d = append(d, buildPluginError(
			"BE send unsupported OAuth2 TokenEndpointAuthMethod: "+resp.Config.TokenEndpointAuthMethod.String(),
		))
	}
	setData(&d, data, oauth2ApplicationTokenEndpointAuthMethodKey, tokenEPAuthMethod)

	setData(&d, data, oauth2ApplicationTokenEndpointAuthSigningAlgKey, resp.Config.TokenEndpointAuthSigningAlg)
	setData(&d, data, oauth2ApplicationUserinfoSignedResponseAlgKey, resp.Config.UserinfoSignedResponseAlg)
	return d
}
