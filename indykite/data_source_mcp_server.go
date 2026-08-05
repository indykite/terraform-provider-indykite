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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMCPServerList() *schema.Resource {
	entrySchema := listEntryCommonSchema(true)
	entrySchema[appAgentIDKey] = computedStringSchema(appAgentIDDescription)
	entrySchema[mcpServerTokenIntrospectIDKey] = computedStringSchema(
		"Identifier of the Token Introspect configuration")
	entrySchema[mcpServerScopesSupportedKey] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "Scopes supported by the MCP server",
	}
	entrySchema[mcpServerEnabledKey] = &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the MCP server is enabled",
	}

	list := &restListDataSource[MCPServerResponse]{
		path:            "/mcp-servers",
		scopeQueryParam: "project_id",
		scopeKey:        appSpaceIDKey,
		listAttrName:    "mcp_servers",
		fullFetch:       true,
		withFilter:      true,
		description:     "List MCP Server configurations in the given Application Space, filtered by exact name match.",
		entrySchema:     entrySchema,
		flatten: func(server *MCPServerResponse) map[string]any {
			return map[string]any{
				"id":                          server.ID,
				customerIDKey:                 server.CustomerID,
				appSpaceIDKey:                 server.AppSpaceID,
				nameKey:                       server.Name,
				displayNameKey:                server.DisplayName,
				descriptionKey:                server.Description,
				appAgentIDKey:                 server.AppAgentID,
				mcpServerTokenIntrospectIDKey: server.TokenIntrospectID,
				mcpServerScopesSupportedKey:   server.ScopesSupported,
				mcpServerEnabledKey:           server.Enabled,
			}
		},
	}
	return list.dataSource()
}
