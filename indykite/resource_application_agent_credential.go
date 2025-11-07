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
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	publicKeyPEMKey = "public_key_pem"
	publicKeyJWKKey = "public_key_jwk"
	expireTimeKey   = "expire_time"
	kidKey          = "kid"
	agentConfigKey  = "agent_config"
)

func resourceApplicationAgentCredential() *schema.Resource {
	return &schema.Resource{
		Description:   "App agent credentials is a JSON configuration file that contains a secret key or token. ",
		CreateContext: resAppAgentCredCreate,
		ReadContext:   resAppAgentCredRead,
		DeleteContext: resAppAgentCredDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts("update"),
		Schema: map[string]*schema.Schema{
			customerIDKey:    setComputed(customerIDSchema()),
			appSpaceIDKey:    setComputed(appSpaceIDSchema()),
			applicationIDKey: setComputed(applicationIDSchema()),
			appAgentIDKey: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: ValidateGID,
				Description:      appAgentIDDescription,
			},

			displayNameKey: {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				DiffSuppressFunc: DisplayNameCredentialDiffSuppress,
			},
			publicKeyPEMKey: {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Deprecated:    "This field is deprecated.",
				ConflictsWith: []string{publicKeyJWKKey},
				ValidateFunc: validation.All(
					validation.StringMatch(
						// \s* is required, because Terraform does not trim inputs before validation
						regexp.MustCompile(`(?s)^\s*-----BEGIN PUBLIC KEY-----\n.*\n-----END PUBLIC KEY-----\s*$`),
						"key must starts with '-----BEGIN PUBLIC KEY-----' and ends with '-----END PUBLIC KEY-----'",
					),
					validation.StringLenBetween(256, 8192),
				),
				Description: "Provide your onw Public key in PEM format, otherwise new pair is generated",
			},
			publicKeyJWKKey: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Deprecated:       "This field is deprecated.",
				ConflictsWith:    []string{publicKeyPEMKey},
				DiffSuppressFunc: structure.SuppressJsonDiff,
				ValidateFunc:     validation.All(validation.StringIsJSON, validation.StringLenBetween(96, 8192)),
				Description:      "Provide your onw Public key in JWK format, otherwise new pair is generated",
			},
			expireTimeKey: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description:  "Optional date-time when credentials are going to expire",
			},
			kidKey:         {Type: schema.TypeString, Computed: true},
			agentConfigKey: {Type: schema.TypeString, Computed: true, Sensitive: true},
			createTimeKey:  createTimeSchema(),
		},
	}
}

func resAppAgentCredCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	req := CreateApplicationAgentCredentialRequest{
		ApplicationAgentID: data.Get(appAgentIDKey).(string),
		DisplayName:        data.Get(displayNameKey).(string),
	}

	if expRaw, ok := data.GetOk(expireTimeKey); ok {
		exp, err := time.Parse(time.RFC3339, expRaw.(string))
		if err != nil {
			return append(d, buildPluginErrorWithAttrName(err.Error(), expireTimeKey))
		}
		req.ExpireTime = exp.Format(time.RFC3339)
	}

	if key, ok := data.GetOk(publicKeyPEMKey); ok {
		req.PublicKeyPEM = strings.TrimSpace(key.(string))
	} else if key, ok := data.GetOk(publicKeyJWKKey); ok {
		req.PublicKeyJWK = strings.TrimSpace(key.(string))
	}

	var resp ApplicationAgentCredentialResponse
	err := clientCtx.GetClient().Post(ctx, "/application-agent-credentials", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)
	setData(&d, data, appAgentIDKey, resp.ApplicationAgentID)
	setData(&d, data, agentConfigKey, resp.AgentConfig)

	return resAppAgentCredRead(ctx, data, meta)
}

func resAppAgentCredRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp ApplicationAgentCredentialResponse
	// Credentials don't have names, only IDs
	err := clientCtx.GetClient().Get(ctx, "/application-agent-credentials/"+data.Id(), &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceID)
	setData(&d, data, applicationIDKey, resp.ApplicationID)
	setData(&d, data, appAgentIDKey, resp.ApplicationAgentID)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, kidKey, resp.Kid)
	setData(&d, data, agentConfigKey, resp.AgentConfig)
	setData(&d, data, createTimeKey, resp.CreateTime)

	if !resp.ExpireTime.IsZero() {
		setData(&d, data, expireTimeKey, resp.ExpireTime)
	}

	return d
}

func resAppAgentCredDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
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
	err := clientCtx.GetClient().Delete(ctx, "/application-agent-credentials/"+data.Id())
	HasFailed(&d, err)
	return d
}
