// Copyright (c) 2026 IndyKite
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
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	auditSigningProviderKey    = "key_provider"
	auditSigningKeyResourceKey = "key_resource"
	auditSigningKidKey         = "kid"
	auditSigningAuthParamsKey  = "auth_params"

	auditSigningProviderPlatformManaged = "PLATFORM_MANAGED"

	auditSigningAuthParamsMaxPairs = 32
	auditSigningAuthParamsMaxLen   = 256
)

// AuditSigningProviders lists the supported audit signing key providers.
var AuditSigningProviders = []string{
	auditSigningProviderPlatformManaged,
	"CUSTOMER_GCP_KMS",
	"CUSTOMER_AWS_KMS",
	"CUSTOMER_AZURE_KEY_VAULT",
}

func resourceAuditSigning() *schema.Resource {
	return &schema.Resource{
		Description: "Audit Signing configuration defines which key is used to sign audit log records of a Project. " +
			"The key is either managed by the IndyKite platform, or provided by the customer through a cloud KMS " +
			"(Google Cloud KMS, AWS KMS or Azure Key Vault), in which case the key resource, key ID and " +
			"authentication parameters must be supplied.",
		CreateContext: resAuditSigningCreate,
		ReadContext:   resAuditSigningRead,
		UpdateContext: resAuditSigningUpdate,
		DeleteContext: resAuditSigningDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		CustomizeDiff: validateAuditSigningCustomerManagedKey,

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:   locationSchema(),
			customerIDKey: setComputed(customerIDSchema()),
			appSpaceIDKey: setComputed(appSpaceIDSchema()),

			nameKey:        nameSchema(),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),
			createdByKey:   createdBySchema(),
			updatedByKey:   updatedBySchema(),

			auditSigningProviderKey: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      auditSigningProviderPlatformManaged,
				ValidateFunc: validation.StringInSlice(AuditSigningProviders, false),
				Description: "Key provider identifies who manages the signing key. " +
					"One of: PLATFORM_MANAGED, CUSTOMER_GCP_KMS, CUSTOMER_AWS_KMS, CUSTOMER_AZURE_KEY_VAULT. " +
					"Defaults to PLATFORM_MANAGED.",
			},
			auditSigningKeyResourceKey: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
				Description: "Resource identifier of the customer managed signing key in the provider's KMS, " +
					"e.g. the Cloud KMS key version name, the AWS KMS key ARN or the Azure Key Vault key identifier.",
			},
			auditSigningKidKey: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
				Description:  "Key ID (kid) published with signed audit records so verifiers can locate the key.",
			},
			auditSigningAuthParamsKey: {
				Type:             schema.TypeMap,
				Optional:         true,
				Sensitive:        true,
				Elem:             &schema.Schema{Type: schema.TypeString},
				ValidateDiagFunc: validateAuditSigningAuthParams,
				Description: "Authentication parameters used to access the customer managed key, " +
					"e.g. service account credentials or access keys. At most 32 entries. " +
					"Values are write-only: the API never returns them, so Terraform keeps the values " +
					"from the configuration and only reconciles the set of keys.",
			},
		},
	}
}

func resAuditSigningCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	req := CreateAuditSigningRequest{
		ProjectID:   data.Get(locationKey).(string),
		Name:        data.Get(nameKey).(string),
		DisplayName: stringValue(optionalString(data, displayNameKey)),
		Description: stringValue(optionalString(data, descriptionKey)),
		Provider:    data.Get(auditSigningProviderKey).(string),
		KeyResource: data.Get(auditSigningKeyResourceKey).(string),
		Kid:         data.Get(auditSigningKidKey).(string),
		AuthParams:  auditSigningAuthParamsFromData(data),
	}

	var resp AuditSigningResponse
	err := clientCtx.GetClient().Post(ctx, "/audit-signings", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

	return resAuditSigningRead(ctx, data, meta)
}

func resAuditSigningRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp AuditSigningResponse
	path := buildReadPath("/audit-signings", data)
	err := clientCtx.GetClient().Get(ctx, path, &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceID)

	if resp.AppSpaceID != "" {
		setData(&d, data, locationKey, resp.AppSpaceID)
	} else if resp.CustomerID != "" {
		setData(&d, data, locationKey, resp.CustomerID)
	}

	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, createdByKey, resp.CreatedBy)
	setData(&d, data, updatedByKey, resp.UpdatedBy)

	setData(&d, data, auditSigningProviderKey, resp.Provider)
	setData(&d, data, auditSigningKeyResourceKey, stringValue(resp.KeyResource))
	setData(&d, data, auditSigningKidKey, stringValue(resp.Kid))
	setData(&d, data, auditSigningAuthParamsKey, mergeAuditSigningAuthParams(data, resp.AuthParams))

	return d
}

func resAuditSigningUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := UpdateAuditSigningRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
		Provider:    data.Get(auditSigningProviderKey).(string),
		KeyResource: data.Get(auditSigningKeyResourceKey).(string),
		Kid:         data.Get(auditSigningKidKey).(string),
		AuthParams:  auditSigningAuthParamsFromData(data),
	}

	var resp AuditSigningResponse
	err := clientCtx.GetClient().Put(ctx, "/audit-signings/"+data.Id(), req, &resp)
	if HasFailed(&d, err) {
		return d
	}

	return resAuditSigningRead(ctx, data, meta)
}

func resAuditSigningDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	err := clientCtx.GetClient().Delete(ctx, "/audit-signings/"+data.Id())
	HasFailed(&d, err)
	return d
}

// validateAuditSigningCustomerManagedKey rule that a customer managed
// provider must name the key: key_resource and kid are required unless the key is
// PLATFORM_MANAGED. Checking it at plan time avoids a failed apply.
func validateAuditSigningCustomerManagedKey(_ context.Context, d *schema.ResourceDiff, _ any) error {
	provider, _ := d.Get(auditSigningProviderKey).(string)
	if provider == auditSigningProviderPlatformManaged {
		return nil
	}
	for _, key := range []string{auditSigningKeyResourceKey, auditSigningKidKey} {
		if v, _ := d.Get(key).(string); v == "" {
			return fmt.Errorf("%q is required when %s is %s", key, auditSigningProviderKey, provider)
		}
	}
	return nil
}

// validateAuditSigningAuthParams enforces the API limits on auth_params: at most 32 entries,
// every key and value between 1 and 256 characters. For schema.TypeMap the SDK only coerces
// values to the Elem type and never runs an Elem ValidateFunc on them, and keys have no schema
// at all, so both checks have to live on the map attribute itself.
func validateAuditSigningAuthParams(i any, path cty.Path) diag.Diagnostics {
	params, ok := i.(map[string]any)
	if !ok {
		return diag.Diagnostics{buildPluginErrorWithPath(
			fmt.Sprintf("validateAuditSigningAuthParams failed, expected map, got %T", i), path)}
	}
	var d diag.Diagnostics
	if len(params) > auditSigningAuthParamsMaxPairs {
		d = append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary: fmt.Sprintf("expected at most %d auth_params entries, got %d",
				auditSigningAuthParamsMaxPairs, len(params)),
			AttributePath: path,
		})
	}
	for _, key := range getMapStringKeys(params) {
		if l := len(key); l < 1 || l > auditSigningAuthParamsMaxLen {
			d = append(d, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf("expected length of auth_params key %q to be in the range (1 - %d), got %d",
					key, auditSigningAuthParamsMaxLen, l),
				AttributePath: path,
			})
		}
		value, _ := params[key].(string)
		if l := len(value); l < 1 || l > auditSigningAuthParamsMaxLen {
			d = append(d, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf("expected length of auth_params.%s to be in the range (1 - %d), got %d",
					key, auditSigningAuthParamsMaxLen, l),
				AttributePath: path.IndexString(key),
			})
		}
	}
	return d
}

// auditSigningAuthParamsFromData reads the auth_params map from Terraform data
// as map[string]string, or nil when unset.
func auditSigningAuthParamsFromData(data *schema.ResourceData) map[string]string {
	raw, _ := data.Get(auditSigningAuthParamsKey).(map[string]any)
	if len(raw) == 0 {
		return nil
	}
	params := make(map[string]string, len(raw))
	for k, v := range raw {
		params[k], _ = v.(string)
	}
	return params
}

// mergeAuditSigningAuthParams reconciles the masked auth_params returned by the API
// (keys present, values blanked) with the values already held in state, so the
// secret material is never lost on refresh while added or removed keys still
// surface as a diff. Keys unknown to the state (e.g. after import) keep the masked value.
func mergeAuditSigningAuthParams(data *schema.ResourceData, apiParams map[string]string) map[string]any {
	if apiParams == nil {
		return nil
	}
	stateParams, _ := data.Get(auditSigningAuthParamsKey).(map[string]any)
	merged := make(map[string]any, len(apiParams))
	for key, masked := range apiParams {
		if v, ok := stateParams[key]; ok {
			merged[key] = v
			continue
		}
		merged[key] = masked
	}
	return merged
}
