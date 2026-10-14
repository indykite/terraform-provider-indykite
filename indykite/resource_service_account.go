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
		CustomizeDiff: roleForceNew,
	}
}

func roleSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: `Role assigned to the service account. Changing it replaces the service account.
		Valid values are: all_editor, all_viewer.
		The API does not return the role: after import, the first apply records the configured role
		without verifying it.`,
		ValidateFunc: validation.StringInSlice([]string{
			"all_editor", "all_viewer",
		}, false),
	}
}

// roleForceNew replaces the service account when its role changes, as the API cannot change it.
// After import the role is not in state, because the API does not return it; the configured role
// is then recorded in place instead.
func roleForceNew(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if d.Id() == "" || !d.HasChange(roleKey) {
		return nil
	}
	if oldRole, _ := d.GetChange(roleKey); oldRole == "" {
		return nil
	}
	return d.ForceNew(roleKey)
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
	// The read response has no role; keep the value from state unless the API returns one.
	if resp.Role != "" {
		setData(&d, data, roleKey, resp.Role)
	}

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

	if hasNameChanged(&d, data) {
		return d
	}

	// deletion_protection is not sent to the API, and the role cannot be updated: it only changes
	// in place when it is recorded after import.
	if !data.HasChangesExcept(deletionProtectionKey, roleKey) {
		return d
	}

	req := UpdateServiceAccountRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
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
