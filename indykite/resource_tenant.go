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
	config "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

func resourceTenant() *schema.Resource {
	return &schema.Resource{
		CreateContext: resTenantCreate,
		ReadContext:   resTenantRead,
		UpdateContext: resTenantUpdate,
		DeleteContext: resTenantDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			customerIDKey:         setComputed(customerIDSchema()),
			appSpaceIDKey:         setComputed(appSpaceIDSchema()),
			issuerIDKey:           issuerIDSchema(),
			nameKey:               nameSchema(),
			displayNameKey:        displayNameSchema(),
			descriptionKey:        descriptionSchema(),
			createTimeKey:         createTimeSchema(),
			updateTimeKey:         updateTimeSchema(),
			deletionProtectionKey: deletionProtectionSchema(),
		},
	}
}

func resTenantCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	name := data.Get(nameKey).(string)
	resp, err := client.getClient().CreateTenant(ctx, &config.CreateTenantRequest{
		IssuerId:    data.Get(issuerIDKey).(string),
		Name:        name,
		DisplayName: optionalString(data, displayNameKey),
		Description: optionalString(data, descriptionKey),
	})
	if hasFailed(&d, err) {
		return d
	}
	data.SetId(resp.Id)

	return resTenantRead(ctx, data, meta)
}

func resTenantRead(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := client.getClient().ReadTenant(ctx, &config.ReadTenantRequest{
		Identifier: &config.ReadTenantRequest_Id{
			Id: data.Id(),
		}})
	if hasFailed(&d, err) {
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

func resTenantUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	// If only change in plan is delete_protection, just ignore the request
	if !data.HasChangeExcept(deletionProtectionKey) {
		return d
	}

	req := &config.UpdateTenantRequest{
		Id:          data.Id(),
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	_, err := client.getClient().UpdateTenant(ctx, req)
	if hasFailed(&d, err) {
		return d
	}
	return resTenantRead(ctx, data, meta)
}

func resTenantDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()
	if hasDeleteProtection(&d, data) {
		return d
	}
	_, err := client.getClient().DeleteTenant(ctx, &config.DeleteTenantRequest{
		Id: data.Id(),
	})
	hasFailed(&d, err)
	return d
}
