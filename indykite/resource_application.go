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
		Bookmarks:   clientCtx.GetBookmarks(),
	})
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.Id)
	clientCtx.AddBookmarks(resp.GetBookmark())

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
		Bookmarks: clientCtx.GetBookmarks(),
	})
	if readHasFailed(&d, err, data) {
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
		Bookmarks:   clientCtx.GetBookmarks(),
	}

	resp, err := clientCtx.GetClient().UpdateApplication(ctx, req)
	if HasFailed(&d, err) {
		return d
	}
	clientCtx.AddBookmarks(resp.GetBookmark())
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
	resp, err := clientCtx.GetClient().DeleteApplication(ctx, &config.DeleteApplicationRequest{
		Id:        data.Id(),
		Bookmarks: clientCtx.GetBookmarks(),
	})
	HasFailed(&d, err)
	clientCtx.AddBookmarks(resp.GetBookmark())
	return d
}
