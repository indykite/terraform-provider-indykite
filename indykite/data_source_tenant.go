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
	"errors"
	"io"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

func dataSourceTenant() *schema.Resource {
	oneOfTenantID := []string{nameKey, tenantIDKey}
	return &schema.Resource{
		ReadContext: dataTenantReadContext,
		Schema: map[string]*schema.Schema{
			customerIDKey:  setComputed(customerIDSchema()),
			issuerIDKey:    setComputed(issuerIDSchema()),
			appSpaceIDKey:  setRequiredWith(appSpaceIDSchema(), nameKey),
			tenantIDKey:    setExactlyOneOf(tenantIDSchema(), tenantIDKey, oneOfTenantID),
			nameKey:        setRequiredWith(setExactlyOneOf(nameSchema(), nameKey, oneOfTenantID), appSpaceIDKey),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),
		},
		Timeouts: defaultDataTimeouts(),
	}
}

func dataSourceTenantList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataTenantListContext,
		Schema: map[string]*schema.Schema{
			appSpaceIDKey: appSpaceIDSchema(),
			filterKey:     exactNameFilterSchema(),
			"tenants": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						customerIDKey:  setComputed(customerIDSchema()),
						appSpaceIDKey:  setComputed(appSpaceIDSchema()),
						issuerIDKey:    setComputed(issuerIDSchema()),
						"id":           setComputed(tenantIDSchema()),
						nameKey:        nameSchema(),
						displayNameKey: displayNameSchema(),
						descriptionKey: descriptionSchema(),
					},
				},
			},
		},
		Timeouts: defaultDataTimeouts(),
	}
}

func dataTenantReadContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}

	req := &configpb.ReadTenantRequest{
		Bookmarks: clientCtx.GetBookmarks(),
	}
	if name, exists := data.GetOk(nameKey); exists {
		req.Identifier = &configpb.ReadTenantRequest_Name{
			Name: &configpb.UniqueNameIdentifier{
				Name:     name.(string),
				Location: data.Get(appSpaceIDKey).(string),
			},
		}
	} else if id, ok := data.GetOk(tenantIDKey); ok {
		req.Identifier = &configpb.ReadTenantRequest_Id{
			Id: id.(string),
		}
	}

	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadTenant(ctx, req)
	if readHasFailed(&d, err, data) {
		return d
	}

	if resp.GetTenant() == nil {
		return diag.Diagnostics{buildPluginError("empty Tenant response")}
	}
	data.SetId(resp.Tenant.Id)
	setData(&d, data, customerIDKey, resp.Tenant.CustomerId)
	setData(&d, data, appSpaceIDKey, resp.Tenant.AppSpaceId)
	setData(&d, data, issuerIDKey, resp.Tenant.IssuerId)
	setData(&d, data, nameKey, resp.Tenant.Name)
	setData(&d, data, displayNameKey, resp.Tenant.DisplayName)
	setData(&d, data, descriptionKey, resp.Tenant.Description)
	setData(&d, data, createTimeKey, resp.Tenant.CreateTime)
	setData(&d, data, updateTimeKey, resp.Tenant.UpdateTime)
	return d
}

func dataTenantListContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	rawFilter := data.Get(filterKey).([]any)
	match := make([]string, len(rawFilter))
	for i, v := range rawFilter {
		match[i] = v.(string)
	}

	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ListTenants(ctx, &configpb.ListTenantsRequest{
		AppSpaceId: data.Get(appSpaceIDKey).(string),
		Match:      match,
		Bookmarks:  clientCtx.GetBookmarks(),
	})
	if HasFailed(&d, err) {
		return d
	}

	var allTenants []map[string]any
	for {
		app, err := resp.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			HasFailed(&d, err)
			return d
		}
		allTenants = append(allTenants, map[string]any{
			customerIDKey:  app.GetTenant().GetCustomerId(),
			appSpaceIDKey:  app.GetTenant().GetAppSpaceId(),
			issuerIDKey:    app.GetTenant().GetIssuerId(),
			"id":           app.GetTenant().GetId(),
			nameKey:        app.GetTenant().GetName(),
			displayNameKey: app.GetTenant().GetDisplayName(),
			descriptionKey: flattenOptionalString(app.GetTenant().GetDescription()),
		})
	}
	setData(&d, data, "tenants", allTenants)

	id := data.Get(appSpaceIDKey).(string) + "/tenants/" + strings.Join(match, ",")
	data.SetId(id)
	return d
}
