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
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	config "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		UpdateContext: resAppAgentCredUpdate,
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
			appAgentIDKey:    appAgentIDSchema(),

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

	req := &config.RegisterApplicationAgentCredentialRequest{
		ApplicationAgentId: data.Get(appAgentIDKey).(string),
		DisplayName:        data.Get(displayNameKey).(string),
	}

	if expRaw, ok := data.GetOk(expireTimeKey); ok {
		exp, err := time.Parse(time.RFC3339, expRaw.(string))
		if err != nil {
			return append(d, buildPluginErrorWithAttrName(err.Error(), expireTimeKey))
		}
		req.ExpireTime = timestamppb.New(exp)
	}

	if key, ok := data.GetOk(publicKeyPEMKey); ok {
		req.PublicKey = &config.RegisterApplicationAgentCredentialRequest_Pem{
			Pem: []byte(strings.TrimSpace(key.(string))),
		}
	} else if key, ok = data.GetOk(publicKeyJWKKey); ok {
		req.PublicKey = &config.RegisterApplicationAgentCredentialRequest_Jwk{
			Jwk: []byte(strings.TrimSpace(key.(string))),
		}
	}

	resp, err := clientCtx.GetClient().RegisterApplicationAgentCredential(ctx, req)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.GetId())
	setData(&d, data, agentConfigKey, string(resp.GetAgentConfig()))

	return resAppAgentCredRead(ctx, data, meta)
}

func resAppAgentCredUpdate(_ context.Context, data *schema.ResourceData, _ any) diag.Diagnostics {
	var d diag.Diagnostics

	agentConfig, ok := data.Get(agentConfigKey).(string)
	// AgentConfig is empty (nil) and cannot be casted to string
	// It is fine, do not create any error here
	if !ok {
		return d
	}

	mapCfg := map[string]any{}
	err := json.Unmarshal([]byte(agentConfig), &mapCfg)
	if err != nil {
		return append(d, buildPluginErrorWithAttrName(err.Error(), agentConfigKey))
	}

	var byteCfg []byte
	byteCfg, err = json.Marshal(mapCfg)
	if err != nil {
		return append(d, buildPluginErrorWithAttrName(err.Error(), agentConfigKey))
	}

	setData(&d, data, agentConfigKey, string(byteCfg))
	return d
}

func resAppAgentCredRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().
		ReadApplicationAgentCredential(ctx, &config.ReadApplicationAgentCredentialRequest{
			Id: data.Id(),
		})
	if readHasFailed(&d, err, data) {
		return d
	}

	if resp.GetApplicationAgentCredential() == nil {
		return diag.Diagnostics{buildPluginError("empty ApplicationAgentCredential response")}
	}

	data.SetId(resp.GetApplicationAgentCredential().GetId())
	setData(&d, data, customerIDKey, resp.GetApplicationAgentCredential().GetCustomerId())
	setData(&d, data, appSpaceIDKey, resp.GetApplicationAgentCredential().GetAppSpaceId())
	setData(&d, data, applicationIDKey, resp.GetApplicationAgentCredential().GetApplicationId())
	setData(&d, data, appAgentIDKey, resp.GetApplicationAgentCredential().GetApplicationAgentId())

	setData(&d, data, displayNameKey, resp.GetApplicationAgentCredential().GetDisplayName())
	setData(&d, data, kidKey, resp.GetApplicationAgentCredential().GetKid())
	setData(&d, data, createTimeKey, resp.GetApplicationAgentCredential().GetCreateTime())

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
	_, err := clientCtx.GetClient().DeleteApplicationAgentCredential(ctx,
		&config.DeleteApplicationAgentCredentialRequest{
			Id: data.Id(),
		},
	)
	HasFailed(&d, err)
	return d
}
