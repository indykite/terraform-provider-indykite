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

func resourceApplicationSpace() *schema.Resource {
	return &schema.Resource{
		Description:   "It is workspace or environment for your applications.  ",
		CreateContext: resAppSpaceCreateContext,
		ReadContext:   resAppSpaceReadContext,
		UpdateContext: resAppSpaceUpdateContext,
		DeleteContext: resAppSpaceDeleteContext,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			customerIDKey:         customerIDSchema(),
			nameKey:               nameSchema(),
			displayNameKey:        displayNameSchema(),
			descriptionKey:        descriptionSchema(),
			createTimeKey:         createTimeSchema(),
			updateTimeKey:         updateTimeSchema(),
			deletionProtectionKey: deletionProtectionSchema(),
			regionKey:             regionSchema(),
		},
	}
}

func resAppSpaceCreateContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	name := data.Get(nameKey).(string)
	resp, err := clientCtx.GetClient().CreateApplicationSpace(ctx, &config.CreateApplicationSpaceRequest{
		CustomerId:  data.Get(customerIDKey).(string),
		Name:        name,
		DisplayName: optionalString(data, displayNameKey),
		Description: optionalString(data, descriptionKey),
		Region:      data.Get(regionKey).(string),
	})
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.GetId())

	return resAppSpaceReadContext(ctx, data, meta)
}

func resAppSpaceReadContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadApplicationSpace(ctx, &config.ReadApplicationSpaceRequest{
		Identifier: &config.ReadApplicationSpaceRequest_Id{
			Id: data.Id(),
		},
	})
	if readHasFailed(&d, err, data) {
		return d
	}

	if resp.GetAppSpace() == nil {
		return diag.Diagnostics{buildPluginError("empty ApplicationSpace response")}
	}

	data.SetId(resp.GetAppSpace().GetId())
	setData(&d, data, customerIDKey, resp.GetAppSpace().GetCustomerId())
	setData(&d, data, nameKey, resp.GetAppSpace().GetName())
	setData(&d, data, displayNameKey, resp.GetAppSpace().GetDisplayName())
	setData(&d, data, descriptionKey, resp.GetAppSpace().GetDescription())
	setData(&d, data, createTimeKey, resp.GetAppSpace().GetCreateTime())
	setData(&d, data, updateTimeKey, resp.GetAppSpace().GetUpdateTime())
	setData(&d, data, regionKey, resp.GetAppSpace().GetRegion())
	return d
}

func resAppSpaceUpdateContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
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

	req := &config.UpdateApplicationSpaceRequest{
		Id:          data.Id(),
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	_, err := clientCtx.GetClient().UpdateApplicationSpace(ctx, req)
	if HasFailed(&d, err) {
		return d
	}
	return resAppSpaceReadContext(ctx, data, meta)
}

func resAppSpaceDeleteContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
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
	_, err := clientCtx.GetClient().DeleteApplicationSpace(ctx, &config.DeleteApplicationSpaceRequest{
		Id: data.Id(),
	})
	HasFailed(&d, err)
	return d
}
