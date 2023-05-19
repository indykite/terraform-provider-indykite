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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	objects "github.com/indykite/indykite-sdk-go/gen/indykite/objects/v1beta1"
)

const (
	invitationMessageKey      = "invitation_message"
	resetPasswordMessageKey   = "reset_password_message"
	verificationMessageKey    = "email_verification_message"
	oneTimePasswordMessageKey = "one_time_password_message"

	mailTemplateKey       = "template"
	defaultFromAddressKey = "default_from_address"

	emailAddressKey = "address"
	emailNameKey    = "name"
	fromMailKey     = "from"
	replyToMailKey  = "reply_to"
	toMailKey       = "to"
	ccMailKey       = "cc"
	bccMailKey      = "bcc"
	subjectMailKey  = "subject"

	templateIDKey            = "id"
	templateVersionKey       = "version"
	templateHeadersKey       = "headers"
	templateCustomArgsKey    = "custom_arguments"
	templateDynamicValuesKey = "template_dynamic_values"
	templateCategoriesKey    = "categories"
	templateEventPayloadKey  = "event_payload"
	templateSESArnKey        = "ses_arn"

	providerSESKey      = "amazon_ses"
	providerSendgridKey = "sendgrid"

	sesAccessKey       = "access_key_id"
	sesSecretAccessKey = "secret_access_key"
	sesRegionKey       = "region"
	sesConfigSetKey    = "configuration_set_name"
	sesFeedbackAddrKey = "feedback_forwarding_email_address"
	sesReplyToAddrsKey = "reply_to_addresses"

	sendgridAPIKey     = "api_key"
	sendgridSandboxKey = "sandbox_mode"
	sendgridIPPoolKey  = "ip_pool_name"
	sendgridHostKey    = "host"
)

func resourceEmailNotification() *schema.Resource {
	readContext := configReadContextFunc(resourceEmailNotificationFlatten)

	oneOfProvider := []string{providerSESKey, providerSendgridKey /*, providerMailjetKey, providerMailgunKey */}
	return &schema.Resource{
		CreateContext: configCreateContextFunc(resourceEmailNotificationBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceEmailNotificationBuild, readContext),
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
			// config common schema ends here
			defaultFromAddressKey: emailAddressSchema(1),

			// Providers
			providerSESKey:      setExactlyOneOf(providerSESSchema(), providerSESKey, oneOfProvider),
			providerSendgridKey: setExactlyOneOf(providerSendgridSchema(), providerSendgridKey, oneOfProvider),
			// providerMailjetKey:  buildExactlyOneOf(nameSchema(), providerMailjetKey, oneOfProvider),
			// providerMailgunKey:  buildExactlyOneOf(nameSchema(), providerMailgunKey, oneOfProvider),

			// Email templates
			invitationMessageKey:      emailDefinitionSchema(),
			resetPasswordMessageKey:   emailDefinitionSchema(),
			verificationMessageKey:    emailDefinitionSchema(),
			oneTimePasswordMessageKey: emailDefinitionSchema(),
		},
	}
}

func emailAddressSchema(maxItem int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: false,
		Optional: true,
		MaxItems: maxItem,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				emailAddressKey: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: ValidateEmail,
					Description:      `The required email address`,
				},
				emailNameKey: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(0, 200),
					Description:  `Optional email name`,
				},
			},
		},
	}
}

func providerSESSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				sesAccessKey: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(25, 254),
				},
				sesSecretAccessKey: {
					Type:         schema.TypeString,
					Required:     true,
					Sensitive:    true,
					ValidateFunc: validation.StringLenBetween(25, 254),
				},
				sesRegionKey: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(2, 20),
				},
				sesConfigSetKey: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(0, 254),
				},
				defaultFromAddressKey: emailAddressSchema(1),
				sesFeedbackAddrKey: {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: ValidateEmail,
				},
				sesReplyToAddrsKey: {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString, ValidateDiagFunc: ValidateEmail},
				},
			},
		},
	}
}

func providerSendgridSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				sendgridAPIKey: {
					Type:         schema.TypeString,
					Required:     true,
					Sensitive:    true,
					ValidateFunc: validation.StringLenBetween(25, 254),
				},
				sendgridSandboxKey: {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				sendgridIPPoolKey: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 254),
				},
				sendgridHostKey: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				},
			},
		},
	}
}

func emailDefinitionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				fromMailKey:    emailAddressSchema(1),
				replyToMailKey: emailAddressSchema(1),
				toMailKey:      emailAddressSchema(0),
				ccMailKey:      emailAddressSchema(0),
				bccMailKey:     emailAddressSchema(0),
				subjectMailKey: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 503),
				},
				mailTemplateKey: emailTemplateSchema(),
			},
		},
	}
}

func emailTemplateSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				templateIDKey: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 254),
					Description:  "ID of the template taken from selected email provider",
				},
				templateVersionKey: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 254),
				},
				templateHeadersKey: {
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				templateCustomArgsKey: {
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				templateDynamicValuesKey: {
					Type:             schema.TypeString,
					Optional:         true,
					DiffSuppressFunc: structure.SuppressJsonDiff,
					ValidateFunc:     validation.StringIsJSON,
					Description:      `Dynamic template values must be valid JSON string`,
				},
				templateCategoriesKey: {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				templateEventPayloadKey: {
					Type:     schema.TypeString,
					Optional: true,
				},
				templateSESArnKey: {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func resourceEmailNotificationFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) (d diag.Diagnostics) {
	mailConf := resp.GetConfigNode().GetEmailServiceConfig()
	if mailConf == nil {
		return diag.Diagnostics{buildPluginError("config in the response is not valid EmailNotificationConfig")}
	}

	if mailConf.DefaultFromAddress != nil {
		setData(&d, data, defaultFromAddressKey, flattenEmailAddrList([]*configpb.Email{mailConf.DefaultFromAddress}))
	}

	switch provider := mailConf.Provider.(type) {
	case *configpb.EmailServiceConfig_Amazon:
		var oldAccessKey interface{}
		if val, ok := data.Get(providerSESKey).([]interface{}); ok && len(val) > 0 {
			dataMap, _ := val[0].(map[string]interface{})
			oldAccessKey = dataMap[sesSecretAccessKey]
		}
		setData(&d, data, providerSESKey, []map[string]interface{}{{
			sesAccessKey:          provider.Amazon.AccessKeyId,
			sesSecretAccessKey:    oldAccessKey,
			sesRegionKey:          provider.Amazon.Region,
			sesConfigSetKey:       provider.Amazon.ConfigurationSetName,
			defaultFromAddressKey: flattenEmailAddrList([]*configpb.Email{provider.Amazon.DefaultFromAddress}),
			sesReplyToAddrsKey:    provider.Amazon.ReplyToAddresses,
			sesFeedbackAddrKey:    provider.Amazon.FeedbackForwardingEmailAddress,
		}})
	case *configpb.EmailServiceConfig_Sendgrid:
		var oldAPIKey interface{}
		if val, ok := data.Get(providerSendgridKey).([]interface{}); ok && len(val) > 0 {
			dataMap, _ := val[0].(map[string]interface{})
			oldAPIKey = dataMap[sendgridAPIKey]
		}
		setData(&d, data, providerSendgridKey, []map[string]interface{}{{
			sendgridAPIKey:     oldAPIKey,
			sendgridSandboxKey: provider.Sendgrid.SandboxMode,
			sendgridIPPoolKey:  flattenOptionalString(provider.Sendgrid.IpPoolName),
			sendgridHostKey:    flattenOptionalString(provider.Sendgrid.Host),
		}})
	default:
		return append(d, buildPluginError(fmt.Sprintf("Email provider %T is not supported yet", provider)))
	}

	if val, err := flattenMessageDefinition(mailConf.InvitationMessage); err != nil {
		return append(d, buildPluginErrorWithAttrName(
			"Invalid InvitationMessage response"+err.Error(),
			invitationMessageKey,
		))
	} else if val != nil {
		setData(&d, data, invitationMessageKey, []map[string]interface{}{val})
	}

	if val, err := flattenMessageDefinition(mailConf.ResetPasswordMessage); err != nil {
		return append(d, buildPluginErrorWithAttrName(
			"Invalid ResetPasswordMessage response"+err.Error(),
			resetPasswordMessageKey,
		))
	} else if val != nil {
		setData(&d, data, resetPasswordMessageKey, []map[string]interface{}{val})
	}

	if val, err := flattenMessageDefinition(mailConf.VerificationMessage); err != nil {
		return append(d, buildPluginErrorWithAttrName(
			"Invalid VerificationMessage response"+err.Error(),
			verificationMessageKey,
		))
	} else if val != nil {
		setData(&d, data, verificationMessageKey, []map[string]interface{}{val})
	}

	if val, err := flattenMessageDefinition(mailConf.OneTimePasswordMessage); err != nil {
		return append(d, buildPluginErrorWithAttrName(
			"Invalid OneTimePasswordMessage response"+err.Error(),
			oneTimePasswordMessageKey,
		))
	} else if val != nil {
		setData(&d, data, oneTimePasswordMessageKey, []map[string]interface{}{val})
	}

	return d
}

func flattenEmailAddrList(resp []*configpb.Email) []map[string]string {
	// Return empty array rather than nil, this is what Terraform wants
	flatten := make([]map[string]string, 0, len(resp))
	for _, m := range resp {
		if m == nil {
			continue
		}
		flatten = append(flatten, map[string]string{
			emailAddressKey: m.Address,
			emailNameKey:    m.Name,
		})
	}
	return flatten
}

func flattenMessageDefinition(resp *configpb.EmailDefinition) (map[string]interface{}, error) {
	if resp == nil {
		return nil, nil
	}
	t, ok := resp.Email.(*configpb.EmailDefinition_Template)
	if !ok || t.Template == nil {
		return nil, errors.New("only Template email definition is supported now")
	}
	var jsonText []byte
	if t.Template.DynamicTemplateValues != nil {
		m, err := objects.ToMap(t.Template.DynamicTemplateValues)
		if err != nil {
			return nil, err
		}
		jsonText, err = json.Marshal(m)
		if err != nil {
			return nil, err
		}
	}
	return map[string]interface{}{
		fromMailKey:    flattenEmailAddrList([]*configpb.Email{t.Template.From}),
		replyToMailKey: flattenEmailAddrList([]*configpb.Email{t.Template.ReplyTo}),
		toMailKey:      flattenEmailAddrList(t.Template.To),
		ccMailKey:      flattenEmailAddrList(t.Template.Cc),
		bccMailKey:     flattenEmailAddrList(t.Template.Bcc),
		subjectMailKey: t.Template.Subject,
		mailTemplateKey: []map[string]interface{}{{
			templateIDKey:            t.Template.TemplateId,
			templateVersionKey:       flattenOptionalString(t.Template.TemplateVersion),
			templateHeadersKey:       flattenOptionalMap(t.Template.Headers),
			templateCustomArgsKey:    flattenOptionalMap(t.Template.CustomArgs),
			templateDynamicValuesKey: string(jsonText),
			templateCategoriesKey:    flattenOptionalArray(t.Template.Categories),
			templateEventPayloadKey:  flattenOptionalString(t.Template.EventPayload),
			templateSESArnKey:        t.Template.TemplateArn,
		}},
	}, nil
}

func resourceEmailNotificationBuild(
	d *diag.Diagnostics,
	data *schema.ResourceData,
	_ *metaContext,
	builder *config.NodeRequest,
) {
	configNode := &configpb.EmailServiceConfig{}

	if val, ok := data.GetOk(defaultFromAddressKey); ok {
		configNode.DefaultFromAddress = buildEmailAddress(val)
	}

	if val, ok := data.GetOk(providerSESKey); ok {
		mapVal := val.([]interface{})[0].(map[string]interface{})
		provider := &configpb.AmazonSESProviderConfig{
			AccessKeyId:                    mapVal[sesAccessKey].(string),
			SecretAccessKey:                mapVal[sesSecretAccessKey].(string),
			Region:                         mapVal[sesRegionKey].(string),
			ConfigurationSetName:           mapVal[sesConfigSetKey].(string),
			ReplyToAddresses:               rawArrayToStringArray(mapVal[sesReplyToAddrsKey]),
			FeedbackForwardingEmailAddress: mapVal[sesFeedbackAddrKey].(string),
		}

		if fromVal, has := mapVal[defaultFromAddressKey]; has {
			provider.DefaultFromAddress = buildEmailAddress(fromVal)
		}
		configNode.Provider = &configpb.EmailServiceConfig_Amazon{Amazon: provider}
	}

	if val, ok := data.GetOk(providerSendgridKey); ok {
		mapVal := val.([]interface{})[0].(map[string]interface{})
		configNode.Provider = &configpb.EmailServiceConfig_Sendgrid{Sendgrid: &configpb.SendGridProviderConfig{
			ApiKey:      mapVal[sendgridAPIKey].(string),
			SandboxMode: mapVal[sendgridSandboxKey].(bool),
			IpPoolName:  stringToOptionalStringWrapper(mapVal[sendgridIPPoolKey].(string)),
			Host:        stringToOptionalStringWrapper(mapVal[sendgridHostKey].(string)),
		}}
	}

	if val, ok := data.GetOk(verificationMessageKey); ok {
		configNode.VerificationMessage = buildEmailDefinition(val, d,
			cty.GetAttrPath(verificationMessageKey).IndexInt(0).GetAttr(mailTemplateKey),
		)
	}
	if val, ok := data.GetOk(invitationMessageKey); ok {
		configNode.InvitationMessage = buildEmailDefinition(val, d,
			cty.GetAttrPath(invitationMessageKey).IndexInt(0).GetAttr(mailTemplateKey),
		)
	}
	if val, ok := data.GetOk(resetPasswordMessageKey); ok {
		configNode.ResetPasswordMessage = buildEmailDefinition(val, d,
			cty.GetAttrPath(resetPasswordMessageKey).IndexInt(0).GetAttr(mailTemplateKey),
		)
	}
	if val, ok := data.GetOk(oneTimePasswordMessageKey); ok {
		configNode.OneTimePasswordMessage = buildEmailDefinition(val, d,
			cty.GetAttrPath(oneTimePasswordMessageKey).IndexInt(0).GetAttr(mailTemplateKey),
		)
	}
	builder.WithEmailNotificationConfig(configNode)
}

// buildEmailAddrList will cast step-by-step to []map[string]string.
func buildEmailAddrList(rawData interface{}) []*configpb.Email {
	emails := make([]*configpb.Email, len(rawData.([]interface{})))
	for i, v := range rawData.([]interface{}) {
		emails[i] = &configpb.Email{
			Address: v.(map[string]interface{})[emailAddressKey].(string),
			Name:    v.(map[string]interface{})[emailNameKey].(string),
		}
	}
	if len(emails) == 0 {
		return nil
	}
	return emails
}

// buildEmailAddress uses buildEmailAddrList and returns first element or nil.
func buildEmailAddress(rawData interface{}) *configpb.Email {
	if arr := buildEmailAddrList(rawData); len(arr) > 0 {
		return arr[0]
	}
	return nil
}

// buildEmailDefinition casts immediately to []interface{} and first element to map[string]interface{} without checks.
func buildEmailDefinition(rawData interface{}, d *diag.Diagnostics, path cty.Path) *configpb.EmailDefinition {
	data := rawData.([]interface{})[0].(map[string]interface{})
	if len(data[mailTemplateKey].([]interface{})) == 0 {
		*d = append(*d, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "email message must contain template definition",
			AttributePath: path,
		})
		return nil
	}
	templateData := data[mailTemplateKey].([]interface{})[0].(map[string]interface{})
	templateDef := &configpb.EmailTemplate{
		TemplateId:      templateData[templateIDKey].(string),
		TemplateVersion: stringToOptionalStringWrapper(templateData[templateVersionKey].(string)),

		// Following keys are not from template but from parent object
		From:    buildEmailAddress(data[fromMailKey]),
		ReplyTo: buildEmailAddress(data[replyToMailKey]),
		To:      buildEmailAddrList(data[toMailKey]),
		Cc:      buildEmailAddrList(data[ccMailKey]),
		Bcc:     buildEmailAddrList(data[bccMailKey]),
		Subject: data[subjectMailKey].(string),

		Headers:    rawMapToStringMap(templateData[templateHeadersKey]),
		CustomArgs: rawMapToStringMap(templateData[templateCustomArgsKey]),

		Categories:   rawArrayToStringArray(templateData[templateCategoriesKey]),
		EventPayload: stringToOptionalStringWrapper(templateData[templateEventPayloadKey].(string)),
		TemplateArn:  templateData[templateSESArnKey].(string),
	}
	var err error
	templateDef.DynamicTemplateValues, err = buildDynamicTemplateValues(templateData[templateDynamicValuesKey].(string))
	if err != nil {
		*d = append(*d, buildPluginErrorWithAttrName(
			"cannot build dynamic template values from JSON: "+err.Error(),
			err.Error(),
		))
		return nil
	}
	// Currently, only template is supported.
	return &configpb.EmailDefinition{Email: &configpb.EmailDefinition_Template{Template: templateDef}}
}

func buildDynamicTemplateValues(rawJSON string) (map[string]*objects.Value, error) {
	if rawJSON == "" {
		return nil, nil
	}
	jsonMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(rawJSON), &jsonMap); err != nil {
		return nil, err
	}

	return objects.ToMapValue(jsonMap)
}
