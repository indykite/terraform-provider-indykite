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

func resourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resApplicationCreate,
		ReadContext:   resApplicationRead,
		UpdateContext: resApplicationUpdate,
		DeleteContext: resApplicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			customerIDKey:         setComputed(customerIDSchema()),
			appSpaceIDKey:         appSpaceIDSchema(),
			nameKey:               nameSchema(),
			displayNameKey:        displayNameSchema(),
			descriptionKey:        descriptionSchema(),
			createTimeKey:         createTimeSchema(),
			updateTimeKey:         updateTimeSchema(),
			deletionProtectionKey: deletionProtectionSchema(),
		},
	}
}

func resApplicationCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}

	name := data.Get(nameKey).(string)
	resp, err := client.getClient().CreateApplication(ctx, &config.CreateApplicationRequest{
		AppSpaceId:  data.Get(appSpaceIDKey).(string),
		Name:        name,
		DisplayName: optionalString(data, displayNameKey),
		Description: optionalString(data, descriptionKey),
	})
	if hasFailed(&d, err) {
		return d
	}
	data.SetId(resp.Id)

	return resApplicationRead(ctx, data, meta)
}

func resApplicationRead(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	resp, err := client.getClient().ReadApplication(ctx, &config.ReadApplicationRequest{
		Identifier: &config.ReadApplicationRequest_Id{
			Id: data.Id(),
		}})
	if hasFailed(&d, err) {
		return d
	}

	if resp.GetApplication() == nil {
		return diag.Diagnostics{buildPluginError("empty Application response")}
	}

	data.SetId(resp.Application.Id)
	setData(&d, data, customerIDKey, resp.Application.CustomerId)
	setData(&d, data, appSpaceIDKey, resp.Application.AppSpaceId)
	setData(&d, data, nameKey, resp.Application.Name)
	setData(&d, data, displayNameKey, resp.Application.DisplayName)
	setData(&d, data, descriptionKey, resp.Application.Description)
	setData(&d, data, createTimeKey, resp.Application.CreateTime)
	setData(&d, data, updateTimeKey, resp.Application.UpdateTime)
	return d
}

func resApplicationUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}

	// If only change in plan is delete_protection, just ignore the request
	if !data.HasChangeExcept(deletionProtectionKey) {
		return d
	}

	req := &config.UpdateApplicationRequest{
		Id:          data.Id(),
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	_, err := client.getClient().UpdateApplication(ctx, req)
	if hasFailed(&d, err) {
		return d
	}
	return resApplicationRead(ctx, data, meta)
}

func resApplicationDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	if hasDeleteProtection(&d, data) {
		return d
	}
	_, err := client.getClient().DeleteApplication(ctx, &config.DeleteApplicationRequest{
		Id: data.Id(),
	})
	hasFailed(&d, err)
	return d
}
