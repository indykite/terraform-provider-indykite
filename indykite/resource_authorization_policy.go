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
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	authzJSONConfigKey = "json"
	authzTagsKey       = "tags"
	authzStatusKey     = "status"
)

func resourceAuthorizationPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "KBAC leverages the IndyKite Knowledge Graph to express the relationships and  " +
			"context present in the real-world, digitally and deliver context-aware, " +
			"fine-grained authorization decisions.",

		CreateContext: resAuthorizationPolicyCreate,
		ReadContext:   resAuthorizationPolicyRead,
		UpdateContext: resAuthorizationPolicyUpdate,
		DeleteContext: resAuthorizationPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:    locationSchema(),
			customerIDKey:  setComputed(customerIDSchema()),
			appSpaceIDKey:  setComputed(appSpaceIDSchema()),
			nameKey:        nameSchema(),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),

			authzJSONConfigKey: {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
					validation.StringIsJSON,
				),
				Description: "Configuration of Authorization Policy in JSON format, the same one exported by The Hub.",
			},
			authzStatusKey: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(getMapStringKeys(AuthorizationPolicyStatusTypes), false),
				Description: "Status of the Authorization Policy. Possible values are: " +
					strings.Join(getMapStringKeys(AuthorizationPolicyStatusTypes), ", ") + ".",
			},
			authzTagsKey: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Tags of the Authorization Policy.",
			},
		},
	}
}

func resAuthorizationPolicyCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	// Map status from Terraform format to API format
	statusValue := data.Get(authzStatusKey).(string)
	apiStatus := AuthorizationPolicyStatusToAPI[statusValue]

	req := CreateAuthorizationPolicyRequest{
		ProjectID:   data.Get(locationKey).(string),
		Name:        data.Get(nameKey).(string),
		DisplayName: stringValue(optionalString(data, displayNameKey)),
		Description: stringValue(optionalString(data, descriptionKey)),
		Policy:      data.Get(authzJSONConfigKey).(string),
		Status:      apiStatus,
		Tags:        rawArrayToTypedArray[string](data.Get(authzTagsKey).([]any)),
	}

	var resp AuthorizationPolicyResponse
	err := clientCtx.GetClient().Post(ctx, "/authorization-policies", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

	return resAuthorizationPolicyRead(ctx, data, meta)
}

func resAuthorizationPolicyRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp AuthorizationPolicyResponse
	// Support both ID and name?location=parent_id formats
	path := buildReadPath("/authorization-policies", data)
	err := clientCtx.GetClient().Get(ctx, path, &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceID)

	// Set location based on which is present
	if resp.AppSpaceID != "" {
		setData(&d, data, locationKey, resp.AppSpaceID)
	} else if resp.CustomerID != "" {
		setData(&d, data, locationKey, resp.CustomerID)
	}

	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, authzJSONConfigKey, resp.Policy)

	// Map status from API format to Terraform format
	terraformStatus := AuthorizationPolicyStatusFromAPI[resp.Status]
	if terraformStatus == "" {
		terraformStatus = resp.Status // Fallback to original value if not found
	}
	setData(&d, data, authzStatusKey, terraformStatus)
	setData(&d, data, authzTagsKey, resp.Tags)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)

	return d
}

func resAuthorizationPolicyUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := UpdateAuthorizationPolicyRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	if data.HasChange(authzJSONConfigKey) {
		policy := data.Get(authzJSONConfigKey).(string)
		req.Policy = &policy
	}

	if data.HasChange(authzStatusKey) {
		statusValue := data.Get(authzStatusKey).(string)
		apiStatus := AuthorizationPolicyStatusToAPI[statusValue]
		req.Status = &apiStatus
	}

	if data.HasChange(authzTagsKey) {
		req.Tags = rawArrayToTypedArray[string](data.Get(authzTagsKey).([]any))
	}

	var resp AuthorizationPolicyResponse
	err := clientCtx.GetClient().Put(ctx, "/authorization-policies/"+data.Id(), req, &resp)
	if HasFailed(&d, err) {
		return d
	}

	return resAuthorizationPolicyRead(ctx, data, meta)
}

func resAuthorizationPolicyDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	err := clientCtx.GetClient().Delete(ctx, "/authorization-policies/"+data.Id())
	HasFailed(&d, err)
	return d
}
