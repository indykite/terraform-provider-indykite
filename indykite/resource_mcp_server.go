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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	mcpServerTokenIntrospectIDKey = "token_introspect_id"
	mcpServerScopesSupportedKey   = "scopes_supported"
	mcpServerEnabledKey           = "enabled"
)

func resourceMCPServer() *schema.Resource {
	return &schema.Resource{
		Description: `MCP Server configuration registers a Model Context Protocol server with the IndyKite platform.
		It links an Application Agent and a Token Introspect configuration and advertises the OAuth scopes the
		MCP server supports.`,
		CreateContext: resMCPServerCreate,
		ReadContext:   resMCPServerRead,
		UpdateContext: resMCPServerUpdate,
		DeleteContext: resMCPServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},

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

			appAgentIDKey: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateGID,
				Description:      "Identifier of Application Agent used by the MCP server, in GID format.",
			},
			mcpServerTokenIntrospectIDKey: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateGID,
				Description:      "Identifier of Token Introspect configuration used by the MCP server, in GID format.",
			},
			mcpServerScopesSupportedKey: {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "List of OAuth scopes supported by the MCP server. Must contain at least one scope.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 256),
				},
			},
			mcpServerEnabledKey: {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the MCP server is enabled.",
			},
		},
	}
}

func resMCPServerCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	req := CreateMCPServerRequest{
		ProjectID:         data.Get(locationKey).(string),
		Name:              data.Get(nameKey).(string),
		DisplayName:       stringValue(optionalString(data, displayNameKey)),
		Description:       stringValue(optionalString(data, descriptionKey)),
		AppAgentID:        data.Get(appAgentIDKey).(string),
		TokenIntrospectID: data.Get(mcpServerTokenIntrospectIDKey).(string),
		ScopesSupported:   rawArrayToTypedArray[string](data.Get(mcpServerScopesSupportedKey)),
		Enabled:           data.Get(mcpServerEnabledKey).(bool),
	}

	var resp MCPServerResponse
	err := clientCtx.GetClient().Post(ctx, "/mcp-servers", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

	return resMCPServerRead(ctx, data, meta)
}

func resMCPServerRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp MCPServerResponse
	path := buildReadPath("/mcp-servers", data)
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

	setData(&d, data, appAgentIDKey, resp.AppAgentID)
	setData(&d, data, mcpServerTokenIntrospectIDKey, resp.TokenIntrospectID)
	setData(&d, data, mcpServerScopesSupportedKey, resp.ScopesSupported)
	setData(&d, data, mcpServerEnabledKey, resp.Enabled)

	return d
}

func resMCPServerUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := UpdateMCPServerRequest{
		DisplayName:       updateOptionalString(data, displayNameKey),
		Description:       updateOptionalString(data, descriptionKey),
		AppAgentID:        data.Get(appAgentIDKey).(string),
		TokenIntrospectID: data.Get(mcpServerTokenIntrospectIDKey).(string),
		ScopesSupported:   rawArrayToTypedArray[string](data.Get(mcpServerScopesSupportedKey)),
		Enabled:           data.Get(mcpServerEnabledKey).(bool),
	}

	var resp MCPServerResponse
	err := clientCtx.GetClient().Put(ctx, "/mcp-servers/"+data.Id(), req, &resp)
	if HasFailed(&d, err) {
		return d
	}

	return resMCPServerRead(ctx, data, meta)
}

func resMCPServerDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	err := clientCtx.GetClient().Delete(ctx, "/mcp-servers/"+data.Id())
	HasFailed(&d, err)
	return d
}
