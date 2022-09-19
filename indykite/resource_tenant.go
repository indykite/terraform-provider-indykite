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
	config "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
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

	name := data.Get(nameKey).(string)
	resp, err := client.Client().CreateTenant(ctx, &config.CreateTenantRequest{
		IssuerId:    data.Get(issuerIDKey).(string),
		Name:        name,
		DisplayName: optionalString(data, displayNameKey),
		Description: optionalString(data, descriptionKey),
	})
	if hasFailed(&d, err, "error creating Tenant for %q", name) {
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
	resp, err := client.Client().ReadTenant(ctx, &config.ReadTenantRequest{
		Identifier: &config.ReadTenantRequest_Id{
			Id: data.Id(),
		}})
	if err != nil {
		return diag.FromErr(err)
	}

	if resp == nil {
		return diag.Errorf("empty Tenant response")
	}

	data.SetId(resp.Tenant.Id)
	Set(&d, data, customerIDKey, resp.Tenant.CustomerId)
	Set(&d, data, appSpaceIDKey, resp.Tenant.AppSpaceId)
	Set(&d, data, issuerIDKey, resp.Tenant.IssuerId)
	Set(&d, data, nameKey, resp.Tenant.Name)
	Set(&d, data, displayNameKey, resp.Tenant.DisplayName)
	Set(&d, data, descriptionKey, resp.Tenant.Description)
	Set(&d, data, createTimeKey, resp.Tenant.CreateTime)
	Set(&d, data, updateTimeKey, resp.Tenant.UpdateTime)
	return d
}

func resTenantUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)

	// If only change in plan is delete_protection, just ignore the request
	if !data.HasChangeExcept(deletionProtectionKey) {
		return d
	}

	req := &config.UpdateTenantRequest{
		Id:          data.Id(),
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	if client == nil {
		return d
	}
	_, err := client.Client().UpdateTenant(ctx, req)
	if hasFailed(&d, err, "Error while updating Tenant #%s", data.Id()) {
		return d
	}
	return resTenantRead(ctx, data, meta)
}

func resTenantDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	if hasDeleteProtection(&d, data) {
		return d
	}
	_, err := client.Client().DeleteTenant(ctx, &config.DeleteTenantRequest{
		Id: data.Id(),
	})
	if hasFailed(&d, err, "Error while deleting Tenant #%s", data.Id()) {
		return d
	}
	return d
}
