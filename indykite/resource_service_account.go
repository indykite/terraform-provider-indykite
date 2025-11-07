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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	roleKey = "role"
)

func resourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description:   "Service Account is used for authentication to IndyKite APIs.",
		CreateContext: resServiceAccountCreateContext,
		ReadContext:   resServiceAccountReadContext,
		UpdateContext: resServiceAccountUpdateContext,
		DeleteContext: resServiceAccountDeleteContext,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts("update"),
		Schema: map[string]*schema.Schema{
			customerIDKey:         customerIDSchema(),
			nameKey:               nameSchema(),
			displayNameKey:        displayNameSchema(),
			descriptionKey:        descriptionSchema(),
			createTimeKey:         createTimeSchema(),
			updateTimeKey:         updateTimeSchema(),
			deletionProtectionKey: deletionProtectionSchema(),
			roleKey:               roleSchema(),
		},
	}
}

func roleSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
		Description: `Role assigned to the service account.
		Valid values are: all_editor, all_viewer.`,
		ValidateFunc: validation.StringInSlice([]string{
			"all_editor", "all_viewer",
		}, false),
	}
}

func resServiceAccountCreateContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	req := CreateServiceAccountRequest{
		OrganizationID: data.Get(customerIDKey).(string),
		Name:           data.Get(nameKey).(string),
		DisplayName:    stringValue(optionalString(data, displayNameKey)),
		Description:    stringValue(optionalString(data, descriptionKey)),
		Role:           data.Get(roleKey).(string),
	}

	var resp ServiceAccountResponse
	err := clientCtx.GetClient().Post(ctx, "/service-accounts", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)
	return resServiceAccountReadContext(ctx, data, meta)
}

func resServiceAccountReadContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp ServiceAccountResponse
	// Support both ID and name?location=parent_id formats
	path := buildReadPath("/service-accounts", data)
	err := clientCtx.GetClient().Get(ctx, path, &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.OrganizationID)

	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, roleKey, resp.Role)

	return d
}

func resServiceAccountUpdateContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := UpdateServiceAccountRequest{
		DisplayName: optionalString(data, displayNameKey),
		Description: optionalString(data, descriptionKey),
	}

	var resp ServiceAccountResponse
	err := clientCtx.GetClient().Put(ctx, "/service-accounts/"+data.Id(), req, &resp)
	if HasFailed(&d, err) {
		return d
	}

	return resServiceAccountReadContext(ctx, data, meta)
}

func resServiceAccountDeleteContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
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

	err := clientCtx.GetClient().Delete(ctx, "/service-accounts/"+data.Id())
	HasFailed(&d, err)

	return d
}
