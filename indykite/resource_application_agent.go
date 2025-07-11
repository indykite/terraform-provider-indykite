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

func resourceApplicationAgent() *schema.Resource {
	return &schema.Resource{
		Description: "Application agents are the profiles that contain the credentials " +
			"used by applications to connect to the backend.  " +
			"They represent the apps you develop or support, " +
			"and need to integrate. ",
		CreateContext: resAppAgentCreate,
		ReadContext:   resAppAgentRead,
		UpdateContext: resAppAgentUpdate,
		DeleteContext: resAppAgentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			customerIDKey:         setComputed(customerIDSchema()),
			appSpaceIDKey:         setComputed(appSpaceIDSchema()),
			applicationIDKey:      applicationIDSchema(),
			nameKey:               nameSchema(),
			displayNameKey:        displayNameSchema(),
			descriptionKey:        descriptionSchema(),
			createTimeKey:         createTimeSchema(),
			updateTimeKey:         updateTimeSchema(),
			deletionProtectionKey: deletionProtectionSchema(),
			apiPermissionsKey:     apiPermissionsSchema(),
		},
	}
}

func resAppAgentCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	name := data.Get(nameKey).(string)
	apiPermissions := rawArrayToTypedArray[string](data.Get(apiPermissionsKey).([]any))
	resp, err := clientCtx.GetClient().CreateApplicationAgent(ctx, &config.CreateApplicationAgentRequest{
		ApplicationId:  data.Get(applicationIDKey).(string),
		Name:           name,
		DisplayName:    optionalString(data, displayNameKey),
		Description:    optionalString(data, descriptionKey),
		ApiPermissions: apiPermissions,
	})
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.GetId())

	return resAppAgentRead(ctx, data, meta)
}

func resAppAgentRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadApplicationAgent(ctx, &config.ReadApplicationAgentRequest{
		Identifier: &config.ReadApplicationAgentRequest_Id{
			Id: data.Id(),
		},
	})
	if readHasFailed(&d, err, data) {
		return d
	}

	if resp.GetApplicationAgent() == nil {
		return diag.Diagnostics{buildPluginError("empty ApplicationAgent response")}
	}

	data.SetId(resp.GetApplicationAgent().GetId())
	setData(&d, data, customerIDKey, resp.GetApplicationAgent().GetCustomerId())
	setData(&d, data, appSpaceIDKey, resp.GetApplicationAgent().GetAppSpaceId())
	setData(&d, data, applicationIDKey, resp.GetApplicationAgent().GetApplicationId())
	setData(&d, data, nameKey, resp.GetApplicationAgent().GetName())
	setData(&d, data, displayNameKey, resp.GetApplicationAgent().GetDisplayName())
	setData(&d, data, descriptionKey, resp.GetApplicationAgent().GetDescription())
	setData(&d, data, createTimeKey, resp.GetApplicationAgent().GetCreateTime())
	setData(&d, data, updateTimeKey, resp.GetApplicationAgent().GetUpdateTime())
	setData(&d, data, apiPermissionsKey, resp.GetApplicationAgent().GetApiAccessRestriction())
	return d
}

func resAppAgentUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
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

	apiPermissions := rawArrayToTypedArray[string](data.Get(apiPermissionsKey).([]any))
	req := &config.UpdateApplicationAgentRequest{
		Id:             data.Id(),
		DisplayName:    updateOptionalString(data, displayNameKey),
		Description:    updateOptionalString(data, descriptionKey),
		ApiPermissions: apiPermissions,
	}

	_, err := clientCtx.GetClient().UpdateApplicationAgent(ctx, req)
	if HasFailed(&d, err) {
		return d
	}
	return resAppAgentRead(ctx, data, meta)
}

func resAppAgentDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
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
	_, err := clientCtx.GetClient().DeleteApplicationAgent(ctx, &config.DeleteApplicationAgentRequest{
		Id: data.Id(),
	})
	HasFailed(&d, err)
	return d
}
