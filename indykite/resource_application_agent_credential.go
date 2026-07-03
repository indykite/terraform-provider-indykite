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

// A freshly created application agent may not be immediately usable by the
// credential endpoint: after the agent is created the backend still performs
// follow-up actions (node creation) with no way for us to observe completion,
// and the agent may not be visible to the credential endpoint yet. So the
// credential create always waits credCreateInitialWait up front, then retries
// not-found responses via hashicorp/go-retryablehttp with exponential backoff
// (waits doubling from credCreateRetryWaitMin up to credCreateRetryWaitMax),
// giving roughly 2 + (2+4+8+16+30) ≈ 60 seconds of total wait before giving up.
const credCreateMaxRetries = 5

// The wait bounds are vars (not consts) so tests can shorten them via the seam
// in export_test.go instead of waiting the full production backoff.
var (
	credCreateInitialWait  = 2 * time.Second
	credCreateRetryWaitMin = 2 * time.Second
	credCreateRetryWaitMax = 30 * time.Second
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateFunc:     validation.IsRFC3339Time,
				DiffSuppressFunc: ExpireTimeDiffSuppress,
				Description:      "Optional date-time when credentials are going to expire",
			},
			kidKey:         {Type: schema.TypeString, Computed: true},
			agentConfigKey: {Type: schema.TypeString, Computed: true, Sensitive: true},
			createTimeKey:  createTimeSchema(),
		},
	}
}

// postAppAgentCredentialWithRetry issues the credential create request, retrying on
// not-found responses to ride out the delay before a freshly created application
// agent becomes usable by the credential endpoint. It always pauses
// credCreateInitialWait before the first attempt, because agent creation triggers
// backend follow-up actions (node creation) whose completion cannot be observed.
// The final (possibly not-found) error is returned for the caller to classify.
func postAppAgentCredentialWithRetry(
	ctx context.Context,
	client *RestClient,
	req *CreateApplicationAgentCredentialRequest,
) (ApplicationAgentCredentialResponse, error) {
	var resp ApplicationAgentCredentialResponse
	select {
	case <-time.After(credCreateInitialWait):
	case <-ctx.Done():
		return resp, ctx.Err()
	}
	err := client.PostWithRetryOnNotFound(ctx, "/application-agent-credentials", req, &resp,
		credCreateMaxRetries, credCreateRetryWaitMin, credCreateRetryWaitMax)
	return resp, err
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

	resp, err := postAppAgentCredentialWithRetry(ctx, clientCtx.GetClient(), &req)
	if IsNotFoundError(err) {
		// Retries are exhausted and the referenced application agent is still not
		// visible. Surface a hard error: the not-found warning HasFailed would emit
		// leaves Terraform with an inconsistent (created-but-unset) result.
		return append(d, diag.Errorf(
			"application agent %q was not found after %d attempts; it may not have "+
				"finished propagating on the backend",
			req.ApplicationAgentID, credCreateMaxRetries+1)...)
	}
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)
	setData(&d, data, appAgentIDKey, resp.ApplicationAgentID)

	// Save agent_config from the create response — the API only returns
	// the secret at creation time; the GET endpoint never includes it.
	agentConfig := string(resp.AgentConfig)

	readDiags := resAppAgentCredRead(ctx, data, meta)
	d = append(d, readDiags...)

	// Restore the write-once secret so it persists in state, but only if
	// Read succeeded — no point setting it on an invalid resource.
	if agentConfig != "" && !readDiags.HasError() {
		setData(&d, data, agentConfigKey, agentConfig)
	}

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
	if len(resp.AgentConfig) > 0 {
		setData(&d, data, agentConfigKey, string(resp.AgentConfig))
	}
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
