// Copyright (c) 2023 IndyKite
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
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	config "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	idPrefix                 = "container:"
	defaultTenantIDKey       = "default_tenant_id"
	defaultAuthFlowIDKey     = "default_auth_flow_id"
	defaultEmailServiceIDKey = "default_email_service_id" // #nosec G101

	usernamePolicyKey                       = "username_policy"
	usernamePolicyAllowedUsernameFormatsKey = "allowed_username_formats"
	usernamePolicyValidEmailKey             = "valid_email"
	usernamePolicyVerifyEmailKey            = "verify_email"
	usernamePolicyVerifyEmailGracePeriodKey = "verify_email_grace_period"
	usernamePolicyAllowedEmailDomainsKey    = "allowed_email_domains"
	usernamePolicyExclusiveEmailDomainsKey  = "exclusive_email_domains"

	uniquePropertyConstraintsKey = "unique_property_constraints"
)

func buildContainerID(id string) string {
	return idPrefix + id
}

func resourceCustomerConfiguration() *schema.Resource {
	// This is pasted into Markdown, so avoid tabs/spaces in the beginning of line as that has meaning in Markdown.
	desc := `Customer configuration resource manage defaults and other configuration for given Customer.

Most likely it will contain references to other configurations created under Customer,
so it must be managed as a separated resource to avoid circular dependencies.
But be careful, that only 1 configuration per Customer can be created.

This resource cannot be imported, because it is tight to Customer. It will be created automatically
when Customer is created, and deleted when Customer is deleted.`

	return &schema.Resource{
		Description:   desc,
		CreateContext: resCustomerConfigCreateUpdateContext,
		ReadContext:   resCustomerConfigReadContext,
		UpdateContext: resCustomerConfigCreateUpdateContext,
		DeleteContext: noopDelete,
		// This cannot be imported, as there is no ID in backend. Just ID of container.
		Timeouts: defaultTimeouts("create", "delete"),
		Schema: map[string]*schema.Schema{
			customerIDKey: customerIDSchema(),

			defaultAuthFlowIDKey: convertToOptional(baseIDSchema("ID of default Authentication flow", false)),
			defaultEmailServiceIDKey: convertToOptional(
				baseIDSchema("ID of default Email notification provider", false)),
		},
	}
}

func resourceApplicationSpaceConfiguration() *schema.Resource {
	// This is pasted into Markdown, so avoid tabs/spaces in the beginning of line as that has meaning in Markdown.
	desc := `Application space configuration resource manage defaults and other configuration for given AppSpace.

Most likely it will contain references to other configurations created under Application space,
so it must be managed as a separated resource to avoid circular dependencies.
But be careful, that only 1 configuration per AppSpace can be created.

This resource cannot be imported, because it is tight to AppSpace. It will be created automatically
when AppSpace is created, and deleted when AppSpace is deleted.`

	return &schema.Resource{
		Description: desc,

		CreateContext: resApplicationSpaceConfigCreateUpdateContext,
		ReadContext:   resApplicationSpaceConfigReadContext,
		UpdateContext: resApplicationSpaceConfigCreateUpdateContext,
		DeleteContext: noopDelete,
		// This cannot be imported, as there is no ID in backend. Just ID of container.
		Timeouts: defaultTimeouts("create", "delete"),
		Schema: map[string]*schema.Schema{
			appSpaceIDKey: appSpaceIDSchema(),

			defaultTenantIDKey:   convertToOptional(baseIDSchema("ID of default Tenant", false)),
			defaultAuthFlowIDKey: convertToOptional(baseIDSchema("ID of default Authentication flow", false)),
			defaultEmailServiceIDKey: convertToOptional(
				baseIDSchema("ID of default Email notification provider", false)),
			usernamePolicyKey: getUsernamePolicySchema(),
			uniquePropertyConstraintsKey: {
				Type:             schema.TypeMap,
				Optional:         true,
				ValidateDiagFunc: resContainerCfgUniqueConstraintValidation,
				DiffSuppressFunc: func(k, oldValue, newValue string, _ *schema.ResourceData) bool {
					// DiffSuppressFunc is called also with key ending with '.%'. That is length of map.
					if strings.HasSuffix(k, ".%") {
						return oldValue == newValue
					}
					if oldValue == newValue {
						return true
					}
					oldProto := new(config.UniquePropertyConstraint)
					newProto := new(config.UniquePropertyConstraint)
					if (protojson.UnmarshalOptions{AllowPartial: true}.Unmarshal([]byte(oldValue), oldProto) != nil) {
						return false
					}
					if (protojson.UnmarshalOptions{AllowPartial: true}.Unmarshal([]byte(newValue), newProto) != nil) {
						return false
					}
					return proto.Equal(oldProto, newProto)
				},
				Description: "The Unique Property Constraints define the list of identity property names for which the system enforce the unique constraint before storing them. Specify as JSON.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceTenantConfiguration() *schema.Resource {
	// This is pasted into Markdown, so avoid tabs/spaces in the beginning of line as that has meaning in Markdown.
	desc := `Tenant configuration resource manage defaults and other configuration for given Tenant.

Most likely it will contain references to other configurations created under Tenant,
so it must be managed as a separated resource to avoid circular dependencies.
But be careful, that only 1 configuration per Tenant can be created.

This resource cannot be imported, because it is tight to Tenant. It will be created automatically
when Tenant is created, and deleted when Tenant is deleted.`

	return &schema.Resource{
		Description: desc,

		CreateContext: resTenantConfigCreateUpdateContext,
		ReadContext:   resTenantConfigReadContext,
		UpdateContext: resTenantConfigCreateUpdateContext,
		DeleteContext: noopDelete,
		// This cannot be imported, as there is no ID in backend. Just ID of container.
		Timeouts: defaultTimeouts("create", "delete"),
		Schema: map[string]*schema.Schema{
			tenantIDKey: tenantIDSchema(),

			defaultAuthFlowIDKey: convertToOptional(baseIDSchema("ID of default Authentication flow", false)),
			defaultEmailServiceIDKey: convertToOptional(
				baseIDSchema("ID of default Email notification provider", false)),
			usernamePolicyKey: getUsernamePolicySchema(),
		},
	}
}

func noopDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	var d diag.Diagnostics
	return d
}

func resCustomerConfigCreateUpdateContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	resp, err := clientCtx.GetClient().UpdateCustomerConfig(ctx, &config.UpdateCustomerConfigRequest{
		Id: data.Get(customerIDKey).(string),
		Config: &config.CustomerConfig{
			DefaultAuthFlowId:     stringOrEmpty(data, defaultAuthFlowIDKey),
			DefaultEmailServiceId: stringOrEmpty(data, defaultEmailServiceIDKey),
		},
		Bookmarks: clientCtx.GetBookmarks(),
	})
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(buildContainerID(resp.Id))
	clientCtx.AddBookmarks(resp.GetBookmark())

	return resCustomerConfigReadContext(ctx, data, meta)
}

func resCustomerConfigReadContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	resp, err := clientCtx.GetClient().ReadCustomerConfig(ctx, &config.ReadCustomerConfigRequest{
		Id:        data.Get(customerIDKey).(string),
		Bookmarks: clientCtx.GetBookmarks(),
	})
	if readHasFailed(&d, err, data) {
		return d
	}

	setData(&d, data, defaultAuthFlowIDKey, resp.GetConfig().GetDefaultAuthFlowId())
	setData(&d, data, defaultEmailServiceIDKey, resp.GetConfig().GetDefaultEmailServiceId())

	return d
}

func resApplicationSpaceConfigCreateUpdateContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	var err error
	cfg := &config.ApplicationSpaceConfig{
		DefaultAuthFlowId:     stringOrEmpty(data, defaultAuthFlowIDKey),
		DefaultEmailServiceId: stringOrEmpty(data, defaultEmailServiceIDKey),
		DefaultTenantId:       stringOrEmpty(data, defaultTenantIDKey),
		UsernamePolicy:        resContainerCfgUsernamePolicyBuild(&d, data),
	}
	if propMap, ok := data.Get(uniquePropertyConstraintsKey).(map[string]any); ok && len(propMap) > 0 {
		cfg.UniquePropertyConstraints = make(map[string]*config.UniquePropertyConstraint, len(propMap))
		for k, v := range propMap {
			m := new(config.UniquePropertyConstraint)
			err = protojson.UnmarshalOptions{AllowPartial: true}.Unmarshal([]byte(v.(string)), m)
			if err != nil {
				d = append(d, diag.FromErr(err)...)
				continue
			}
			cfg.UniquePropertyConstraints[k] = m
		}
	}
	if d.HasError() {
		return d
	}

	resp, err := clientCtx.GetClient().UpdateApplicationSpaceConfig(ctx, &config.UpdateApplicationSpaceConfigRequest{
		Id:        data.Get(appSpaceIDKey).(string),
		Config:    cfg,
		Bookmarks: clientCtx.GetBookmarks(),
	})
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(buildContainerID(resp.Id))
	clientCtx.AddBookmarks(resp.GetBookmark())

	return resApplicationSpaceConfigReadContext(ctx, data, meta)
}

func resApplicationSpaceConfigReadContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	resp, err := clientCtx.GetClient().ReadApplicationSpaceConfig(ctx, &config.ReadApplicationSpaceConfigRequest{
		Id:        data.Get(appSpaceIDKey).(string),
		Bookmarks: clientCtx.GetBookmarks(),
	})
	if readHasFailed(&d, err, data) {
		return d
	}

	setData(&d, data, defaultAuthFlowIDKey, resp.GetConfig().GetDefaultAuthFlowId())
	setData(&d, data, defaultEmailServiceIDKey, resp.GetConfig().GetDefaultEmailServiceId())
	setData(&d, data, defaultTenantIDKey, resp.GetConfig().GetDefaultTenantId())
	resContainerCfgUsernamePolicyFlatten(&d, data, resp.GetConfig().GetUsernamePolicy())

	props := make(map[string]any, len(resp.GetConfig().GetUniquePropertyConstraints()))
	for k, v := range resp.GetConfig().GetUniquePropertyConstraints() {
		mv, _ := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(v)
		props[k] = string(mv)
	}
	setData(&d, data, uniquePropertyConstraintsKey, props)
	return d
}

func resTenantConfigCreateUpdateContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	var err error
	cfg := &config.TenantConfig{
		DefaultAuthFlowId:     stringOrEmpty(data, defaultAuthFlowIDKey),
		DefaultEmailServiceId: stringOrEmpty(data, defaultEmailServiceIDKey),
		UsernamePolicy:        resContainerCfgUsernamePolicyBuild(&d, data),
	}
	if d.HasError() {
		return d
	}

	resp, err := clientCtx.GetClient().UpdateTenantConfig(ctx, &config.UpdateTenantConfigRequest{
		Id:        data.Get(tenantIDKey).(string),
		Config:    cfg,
		Bookmarks: clientCtx.GetBookmarks(),
	})
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(buildContainerID(resp.Id))
	clientCtx.AddBookmarks(resp.GetBookmark())

	return resTenantConfigReadContext(ctx, data, meta)
}

func resTenantConfigReadContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta any,
) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	resp, err := clientCtx.GetClient().ReadTenantConfig(ctx, &config.ReadTenantConfigRequest{
		Id:        data.Get(tenantIDKey).(string),
		Bookmarks: clientCtx.GetBookmarks(),
	})
	if readHasFailed(&d, err, data) {
		return d
	}

	setData(&d, data, defaultAuthFlowIDKey, resp.GetConfig().GetDefaultAuthFlowId())
	setData(&d, data, defaultEmailServiceIDKey, resp.GetConfig().GetDefaultEmailServiceId())
	resContainerCfgUsernamePolicyFlatten(&d, data, resp.GetConfig().GetUsernamePolicy())
	return d
}

func getUsernamePolicySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: false,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				usernamePolicyAllowedUsernameFormatsKey: {
					Type:        schema.TypeList,
					Optional:    true,
					Description: `Which username format is allowed. Valid values are email, mobile and username`,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{"email", "mobile", "username"}, false),
					},
				},
				usernamePolicyValidEmailKey: {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "If email must be valid with MX record",
				},
				usernamePolicyVerifyEmailKey: {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "If email must be verified by a link sent to the owner",
				},
				usernamePolicyVerifyEmailGracePeriodKey: {
					Type:     schema.TypeString,
					Optional: true,
					ValidateDiagFunc: func(v any, path cty.Path) diag.Diagnostics {
						if _, err := time.ParseDuration(v.(string)); err != nil {
							return diag.Diagnostics{{
								Severity:      diag.Error,
								Summary:       err.Error(),
								AttributePath: path,
							}}
						}
						return nil
					},
					DiffSuppressFunc: SuppressDurationDiff,
				},
				usernamePolicyAllowedEmailDomainsKey: {
					Type:        schema.TypeList,
					Optional:    true,
					Description: `Allowed email domains to register. Can be shared among tenants.`,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				usernamePolicyExclusiveEmailDomainsKey: {
					Type:        schema.TypeList,
					Optional:    true,
					Description: `Allowed email domains to register. Can be shared among tenants.`,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func resContainerCfgUsernamePolicyBuild(d *diag.Diagnostics, data *schema.ResourceData) *config.UsernamePolicy {
	val, ok := data.GetOk(usernamePolicyKey)
	if !ok {
		return nil
	}
	mapVal := val.([]any)[0].(map[string]any)

	cfg := &config.UsernamePolicy{
		AllowedUsernameFormats: rawArrayToStringArray(mapVal[usernamePolicyAllowedUsernameFormatsKey]),
		ValidEmail:             mapVal[usernamePolicyValidEmailKey].(bool),
		VerifyEmail:            mapVal[usernamePolicyVerifyEmailKey].(bool),
		AllowedEmailDomains:    rawArrayToStringArray(mapVal[usernamePolicyAllowedEmailDomainsKey]),
		ExclusiveEmailDomains:  rawArrayToStringArray(mapVal[usernamePolicyExclusiveEmailDomainsKey]),
	}
	if g, ok := mapVal[usernamePolicyVerifyEmailGracePeriodKey].(string); ok && g != "" {
		gd, err := time.ParseDuration(g)
		if err != nil {
			*d = append(*d, diag.FromErr(err)...)
			return nil
		}
		cfg.VerifyEmailGracePeriod = durationpb.New(gd)
	}
	return cfg
}

func resContainerCfgUsernamePolicyFlatten(
	d *diag.Diagnostics,
	data *schema.ResourceData,
	policy *config.UsernamePolicy,
) {
	if policy == nil {
		return
	}
	var gracePeriod string
	if policy.VerifyEmailGracePeriod != nil {
		gracePeriod = policy.VerifyEmailGracePeriod.AsDuration().String()
	}
	setData(d, data, usernamePolicyKey, []map[string]any{{
		usernamePolicyAllowedUsernameFormatsKey: policy.AllowedUsernameFormats,
		usernamePolicyValidEmailKey:             policy.ValidEmail,
		usernamePolicyVerifyEmailKey:            policy.VerifyEmail,
		usernamePolicyVerifyEmailGracePeriodKey: gracePeriod,
		usernamePolicyAllowedEmailDomainsKey:    policy.AllowedEmailDomains,
		usernamePolicyExclusiveEmailDomainsKey:  policy.ExclusiveEmailDomains,
	}})
}

func resContainerCfgUniqueConstraintValidation(value any, path cty.Path) diag.Diagnostics {
	d := validation.MapKeyMatch(
		regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,253}$`),
		"Only A-Z, numbers and _ is allowed, and must start with letter",
	)(value, path)

	mapVal := value.(map[string]any)
	for key, v := range mapVal {
		dErr := diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Invalid map value",
			AttributePath: append(path, cty.IndexStep{Key: cty.StringVal(key)}),
		}

		if _, err := structure.NormalizeJsonString(v); err != nil {
			dErr.Detail = "value is not valid JSON: " + err.Error()
			d = append(d, dErr)
			continue
		}

		// Do not allow partial if we cannot PreserveDiff
		m := new(config.UniquePropertyConstraint)
		err := protojson.UnmarshalOptions{AllowPartial: true}.Unmarshal([]byte(v.(string)), m)
		if err != nil {
			dErr.Detail = err.Error()
			d = append(d, dErr)
			continue
		}
		if err = m.Validate(); err != nil {
			dErr.Detail = err.Error()
			d = append(d, dErr)
		}
	}
	return d
}
