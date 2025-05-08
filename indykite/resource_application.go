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

func resourceApplication() *schema.Resource {
	return &schema.Resource{
		Description: "An application represents the center of the solution, " +
			"and is also the legal entity users legally interact with. " +
			"Each application is created in an ApplicationSpace or project, and has a profile, " +
			"an application agent and application agent credentials. ",
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

func resApplicationCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	name := data.Get(nameKey).(string)
	resp, err := clientCtx.GetClient().CreateApplication(ctx, &config.CreateApplicationRequest{
		AppSpaceId:  data.Get(appSpaceIDKey).(string),
		Name:        name,
		DisplayName: optionalString(data, displayNameKey),
		Description: optionalString(data, descriptionKey),
	})
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.GetId())

	return resApplicationRead(ctx, data, meta)
}

func resApplicationRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadApplication(ctx, &config.ReadApplicationRequest{
		Identifier: &config.ReadApplicationRequest_Id{
			Id: data.Id(),
		},
	})
	if readHasFailed(&d, err, data) {
		return d
	}

	if resp.GetApplication() == nil {
		return diag.Diagnostics{buildPluginError("empty Application response")}
	}

	data.SetId(resp.GetApplication().GetId())
	setData(&d, data, customerIDKey, resp.GetApplication().GetCustomerId())
	setData(&d, data, appSpaceIDKey, resp.GetApplication().GetAppSpaceId())
	setData(&d, data, nameKey, resp.GetApplication().GetName())
	setData(&d, data, displayNameKey, resp.GetApplication().GetDisplayName())
	setData(&d, data, descriptionKey, resp.GetApplication().GetDescription())
	setData(&d, data, createTimeKey, resp.GetApplication().GetCreateTime())
	setData(&d, data, updateTimeKey, resp.GetApplication().GetUpdateTime())
	return d
}

func resApplicationUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	// If only change in plan is delete_protection, just ignore the request
	if !data.HasChangeExcept(deletionProtectionKey) {
		return d
	}

	req := &config.UpdateApplicationRequest{
		Id:          data.Id(),
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	_, err := clientCtx.GetClient().UpdateApplication(ctx, req)
	if HasFailed(&d, err) {
		return d
	}
	return resApplicationRead(ctx, data, meta)
}

func resApplicationDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()
	if hasDeleteProtection(&d, data) {
		return d
	}
	_, err := clientCtx.GetClient().DeleteApplication(ctx, &config.DeleteApplicationRequest{
		Id: data.Id(),
	})
	HasFailed(&d, err)
	return d
}
