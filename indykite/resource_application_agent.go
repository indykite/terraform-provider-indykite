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

	apiPermissions := rawArrayToTypedArray[string](data.Get(apiPermissionsKey).([]any))
	req := CreateApplicationAgentRequest{
		ApplicationID:  data.Get(applicationIDKey).(string),
		Name:           data.Get(nameKey).(string),
		DisplayName:    stringValue(optionalString(data, displayNameKey)),
		Description:    stringValue(optionalString(data, descriptionKey)),
		APIPermissions: apiPermissions,
	}

	var resp ApplicationAgentResponse
	err := clientCtx.GetClient().Post(ctx, "/application-agents", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

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

	var resp ApplicationAgentResponse
	// Support both ID and name?location=parent_id formats
	path := buildReadPath("/application-agents", data)
	err := clientCtx.GetClient().Get(ctx, path, &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceID)
	setData(&d, data, applicationIDKey, resp.ApplicationID)
	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, apiPermissionsKey, resp.APIPermissions)
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
	req := UpdateApplicationAgentRequest{
		DisplayName:    updateOptionalString(data, displayNameKey),
		Description:    updateOptionalString(data, descriptionKey),
		APIPermissions: apiPermissions,
	}

	var resp ApplicationAgentResponse
	err := clientCtx.GetClient().Put(ctx, "/application-agents/"+data.Id(), req, &resp)
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
	err := clientCtx.GetClient().Delete(ctx, "/application-agents/"+data.Id())
	HasFailed(&d, err)
	return d
}
