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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	serviceAccountIDKey     = "service_account_id"
	serviceAccountConfigKey = "service_account_config"
)

func resourceServiceAccountCredential() *schema.Resource {
	return &schema.Resource{
		Description:   "Service Account Credential is a JSON configuration file that contains a secret key or token for authenticating to IndyKite APIs.",
		CreateContext: resServiceAccountCredCreate,
		ReadContext:   resServiceAccountCredRead,
		DeleteContext: resServiceAccountCredDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts("update"),
		Schema: map[string]*schema.Schema{
			customerIDKey: setComputed(customerIDSchema()),
			serviceAccountIDKey: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateGID,
				Description:      "Identifier of Service Account to which the credential will be registered.",
			},
			displayNameKey: {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				DiffSuppressFunc: DisplayNameCredentialDiffSuppress,
				Description:      "Optional human readable name of the credential.",
			},
			// Computed: when omitted, the API sets a far-future expiry that must not force replacement.
			expireTimeKey: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateFunc:     validation.IsRFC3339Time,
				DiffSuppressFunc: ExpireTimeDiffSuppress,
				Description:      "Optional date-time when credentials are going to expire in RFC3339 format.",
			},
			kidKey: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key identifier of the credential.",
			},
			serviceAccountConfigKey: {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "JSON configuration of the created credential. This is only available after creation.",
			},
			createTimeKey: createTimeSchema(),
		},
	}
}

func resServiceAccountCredCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	req := CreateServiceAccountCredentialRequest{
		ServiceAccountID: data.Get(serviceAccountIDKey).(string),
		DisplayName:      data.Get(displayNameKey).(string),
	}

	if expRaw, ok := data.GetOk(expireTimeKey); ok {
		exp, err := time.Parse(time.RFC3339, expRaw.(string))
		if err != nil {
			return append(d, buildPluginErrorWithAttrName(err.Error(), expireTimeKey))
		}
		req.ExpireTime = exp.Format(time.RFC3339)
	}

	var resp ServiceAccountCredentialResponse
	err := clientCtx.GetClient().Post(ctx, "/service-account-credentials", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)
	setData(&d, data, serviceAccountIDKey, resp.ServiceAccountID)

	// The configuration is only returned on create, so it is set here and never on read.
	config, err := serviceAccountConfigString(resp.ServiceAccountConfig)
	if HasFailed(&d, err) {
		return d
	}
	setData(&d, data, serviceAccountConfigKey, config)

	return resServiceAccountCredRead(ctx, data, meta)
}

// serviceAccountConfigString converts the service_account_config of the create response to the
// JSON text stored in state. The API returns a JSON object; a JSON string is accepted as well.
func serviceAccountConfigString(raw json.RawMessage) (string, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return "", nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return "", fmt.Errorf("failed to decode service_account_config: %w", err)
		}
		return s, nil
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, trimmed); err != nil {
		return "", fmt.Errorf("failed to decode service_account_config: %w", err)
	}
	return buf.String(), nil
}

func resServiceAccountCredRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp ServiceAccountCredentialResponse
	// Credentials don't have names, only IDs
	err := clientCtx.GetClient().Get(ctx, "/service-account-credentials/"+data.Id(), &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.OrganizationID)

	setData(&d, data, serviceAccountIDKey, resp.ServiceAccountID)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, kidKey, resp.Kid)
	setData(&d, data, createTimeKey, resp.CreateTime)
	if resp.ExpireTime != "" {
		setData(&d, data, expireTimeKey, resp.ExpireTime)
	}

	return d
}

func resServiceAccountCredDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	if hasDeleteProtection(&d, data) {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()
	err := clientCtx.GetClient().Delete(ctx, "/service-account-credentials/"+data.Id())
	HasFailed(&d, err)
	return d
}
