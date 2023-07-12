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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	config "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

const (
	idPrefix                 = "container:"
	defaultTenantIDKey       = "default_tenant_id"
	defaultAuthFlowIDKey     = "default_auth_flow_id"
	defaultEmailServiceIDKey = "default_email_service_id"
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
			// Add username policy and Property uniqueness when supported by BE
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
			// Add username policy when supported by BE
		},
	}
}

func noopDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) (d diag.Diagnostics) {
	return d
}

func resCustomerConfigCreateUpdateContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta interface{},
) (d diag.Diagnostics) {
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
	if hasFailed(&d, err) {
		return d
	}
	data.SetId(buildContainerID(resp.Id))
	clientCtx.AddBookmarks(resp.GetBookmark())

	return resCustomerConfigReadContext(ctx, data, meta)
}

func resCustomerConfigReadContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta interface{},
) (d diag.Diagnostics) {
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
	if hasFailed(&d, err) {
		return d
	}

	setData(&d, data, defaultAuthFlowIDKey, resp.GetConfig().GetDefaultAuthFlowId())
	setData(&d, data, defaultEmailServiceIDKey, resp.GetConfig().GetDefaultEmailServiceId())

	return d
}

func resApplicationSpaceConfigCreateUpdateContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta interface{},
) (d diag.Diagnostics) {
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	resp, err := clientCtx.GetClient().UpdateApplicationSpaceConfig(ctx, &config.UpdateApplicationSpaceConfigRequest{
		Id: data.Get(appSpaceIDKey).(string),
		Config: &config.ApplicationSpaceConfig{
			DefaultAuthFlowId:     stringOrEmpty(data, defaultAuthFlowIDKey),
			DefaultEmailServiceId: stringOrEmpty(data, defaultEmailServiceIDKey),
			DefaultTenantId:       stringOrEmpty(data, defaultTenantIDKey),
		},
		Bookmarks: clientCtx.GetBookmarks(),
	})
	if hasFailed(&d, err) {
		return d
	}
	data.SetId(buildContainerID(resp.Id))
	clientCtx.AddBookmarks(resp.GetBookmark())

	return resApplicationSpaceConfigReadContext(ctx, data, meta)
}

func resApplicationSpaceConfigReadContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta interface{},
) (d diag.Diagnostics) {
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
	if hasFailed(&d, err) {
		return d
	}

	setData(&d, data, defaultAuthFlowIDKey, resp.GetConfig().GetDefaultAuthFlowId())
	setData(&d, data, defaultEmailServiceIDKey, resp.GetConfig().GetDefaultEmailServiceId())
	setData(&d, data, defaultTenantIDKey, resp.GetConfig().GetDefaultTenantId())

	return d
}

func resTenantConfigCreateUpdateContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta interface{},
) (d diag.Diagnostics) {
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	resp, err := clientCtx.GetClient().UpdateTenantConfig(ctx, &config.UpdateTenantConfigRequest{
		Id: data.Get(tenantIDKey).(string),
		Config: &config.TenantConfig{
			DefaultAuthFlowId:     stringOrEmpty(data, defaultAuthFlowIDKey),
			DefaultEmailServiceId: stringOrEmpty(data, defaultEmailServiceIDKey),
		},
		Bookmarks: clientCtx.GetBookmarks(),
	})
	if hasFailed(&d, err) {
		return d
	}
	data.SetId(buildContainerID(resp.Id))
	clientCtx.AddBookmarks(resp.GetBookmark())

	return resTenantConfigReadContext(ctx, data, meta)
}

func resTenantConfigReadContext(
	ctx context.Context,
	data *schema.ResourceData,
	meta interface{},
) (d diag.Diagnostics) {
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
	if hasFailed(&d, err) {
		return d
	}

	setData(&d, data, defaultAuthFlowIDKey, resp.GetConfig().GetDefaultAuthFlowId())
	setData(&d, data, defaultEmailServiceIDKey, resp.GetConfig().GetDefaultEmailServiceId())

	return d
}
