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

func resourceApplication() *schema.Resource {
	return &schema.Resource{
		Description: "An application represents the center of the solution, " +
			"and is also the legal entity users legally interact with. " +
			"Each application is created in an ApplicationSpace and has a profile, " +
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

	req := CreateApplicationRequest{
		ProjectID:   data.Get(appSpaceIDKey).(string),
		Name:        data.Get(nameKey).(string),
		DisplayName: stringValue(optionalString(data, displayNameKey)),
		Description: stringValue(optionalString(data, descriptionKey)),
	}

	var resp ApplicationResponse
	err := clientCtx.GetClient().Post(ctx, "/applications", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

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

	var resp ApplicationResponse
	// Support both ID and name?location=parent_id formats
	path := buildReadPath("/applications", data)
	err := clientCtx.GetClient().Get(ctx, path, &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceID)
	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
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

	req := UpdateApplicationRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	var resp ApplicationResponse
	err := clientCtx.GetClient().Put(ctx, "/applications/"+data.Id(), req, &resp)
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
	err := clientCtx.GetClient().Delete(ctx, "/applications/"+data.Id())
	HasFailed(&d, err)
	return d
}

// stringValue converts a string pointer to a string value.
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
