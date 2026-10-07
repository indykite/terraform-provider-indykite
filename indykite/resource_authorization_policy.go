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
			StateContext: projectStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:    locationSchema(),
			projectIDKey:   projectIDSchema(),
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
				Description: "Configuration of Authorization Policy in JSON format, the same one exported by The Hub. " +
					"The claims of the request tokens are available to the policy under the reserved names " +
					"`$token` (the end-user access token from the Authorization header) and `$ik_token` " +
					"(the IndyKite delegated token from the X-IK-Token header, including its RFC 8693 `act` " +
					"delegation chain). A KBAC condition cypher reads them as parameters, e.g. " +
					"`resource.delegated_to = $ik_token.act.sub`, and a CIQ policy filter reads them as " +
					"attribute or value, e.g. `\"attribute\": \"$ik_token.act.sub\"`. The platform binds both names " +
					"on every request, so they must not be listed as input params or supplied by callers; a token " +
					"that did not arrive binds an empty claim set and the condition fails closed.",
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
		ProjectID:   projectIDFromData(data),
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
	setProjectIDData(&d, data, resp.AppSpaceID, resp.CustomerID)

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

	policy := data.Get(authzJSONConfigKey).(string)
	statusValue := data.Get(authzStatusKey).(string)
	apiStatus := AuthorizationPolicyStatusToAPI[statusValue]

	req := UpdateAuthorizationPolicyRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
		Policy:      &policy,
		Status:      &apiStatus,
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
