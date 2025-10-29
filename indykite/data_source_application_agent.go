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
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAppAgent() *schema.Resource {
	oneOfAppID := []string{nameKey, appAgentIDKey}
	return &schema.Resource{
		Description: "Application agents are the profiles that contain the credentials " +
			"used by applications to connect to the backend.  " +
			"They represent the apps you develop or support, " +
			"and need to integrate. ",
		ReadContext: dataAppAgentReadContext,
		Schema: map[string]*schema.Schema{
			customerIDKey:     setComputed(customerIDSchema()),
			applicationIDKey:  setComputed(applicationIDSchema()),
			appAgentIDKey:     setExactlyOneOf(appAgentIDSchema(), appAgentIDKey, oneOfAppID),
			nameKey:           setRequiredWith(setExactlyOneOf(nameSchema(), nameKey, oneOfAppID), appSpaceIDKey),
			appSpaceIDKey:     setRequiredWith(appSpaceIDSchema(), nameKey),
			displayNameKey:    displayNameSchema(),
			descriptionKey:    descriptionSchema(),
			apiPermissionsKey: apiPermissionsSchema(),
			createTimeKey:     createTimeSchema(),
			updateTimeKey:     updateTimeSchema(),
		},
		Timeouts: defaultDataTimeouts(),
	}
}

func dataSourceAppAgentList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataAppAgentListContext,
		Schema: map[string]*schema.Schema{
			appSpaceIDKey: appSpaceIDSchema(),
			filterKey:     exactNameFilterSchema(),
			"app_agents": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						customerIDKey:     setComputed(customerIDSchema()),
						appSpaceIDKey:     setComputed(appSpaceIDSchema()),
						applicationIDKey:  setComputed(applicationIDSchema()),
						"id":              setComputed(appAgentIDSchema()),
						nameKey:           nameSchema(),
						displayNameKey:    displayNameSchema(),
						descriptionKey:    descriptionSchema(),
						apiPermissionsKey: apiPermissionsSchema(),
					},
				},
			},
		},
		Timeouts: defaultDataTimeouts(),
	}
}

// lookupApplicationAgentByName finds an application agent by name within an app space.
func lookupApplicationAgentByName(
	ctx context.Context,
	clientCtx *ClientContext,
	data *schema.ResourceData,
	name string,
) (ApplicationAgentResponse, diag.Diagnostic) {
	appSpaceID := data.Get(appSpaceIDKey).(string)
	var listResp ListApplicationAgentsResponse
	err := clientCtx.GetClient().Get(ctx, "/application-agents?appSpaceId="+appSpaceID, &listResp)
	if err != nil {
		return ApplicationAgentResponse{}, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("failed to list application agents: %v", err),
		}
	}

	// Find application agent by name
	for i := range listResp.Agents {
		if listResp.Agents[i].Name == name {
			return listResp.Agents[i], diag.Diagnostic{}
		}
	}

	return ApplicationAgentResponse{}, diag.Diagnostic{
		Severity: diag.Error,
		Summary: fmt.Sprintf(
			"application agent with name '%s' not found in app space '%s'",
			name, appSpaceID),
	}
}

func dataAppAgentReadContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}

	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp ApplicationAgentResponse
	var err error

	// Check if we have an ID or need to look up by name
	if id, ok := data.GetOk(appAgentIDKey); ok {
		// Direct lookup by ID
		err = clientCtx.GetClient().Get(ctx, "/application-agents/"+id.(string), &resp)
	} else if name, exists := data.GetOk(nameKey); exists {
		// Look up by name within app space
		var diagErr diag.Diagnostic
		resp, diagErr = lookupApplicationAgentByName(ctx, clientCtx, data, name.(string))
		// Only return if there's an actual error (non-zero severity with summary)
		if diagErr.Summary != "" {
			return append(d, diagErr)
		}
	}

	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceID)
	setData(&d, data, applicationIDKey, resp.ApplicationID)
	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, apiPermissionsKey, resp.APIPermissions)
	return d
}

func dataAppAgentListContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	rawFilter := data.Get(filterKey).([]any)
	match := make([]string, len(rawFilter))
	for i, v := range rawFilter {
		match[i] = v.(string)
	}

	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	appSpaceID := data.Get(appSpaceIDKey).(string)
	var resp ListApplicationAgentsResponse
	err := clientCtx.GetClient().Get(ctx, "/application-agents?appSpaceId="+appSpaceID, &resp)
	if HasFailed(&d, err) {
		return d
	}

	allApplicationAgents := make([]map[string]any, 0, len(resp.Agents))
	for i := range resp.Agents {
		agent := &resp.Agents[i]
		// Apply filter if specified
		if len(match) > 0 {
			matchFound := false
			for _, filter := range match {
				if strings.Contains(agent.Name, filter) {
					matchFound = true
					break
				}
			}
			if !matchFound {
				continue
			}
		}

		allApplicationAgents = append(allApplicationAgents, map[string]any{
			customerIDKey:     agent.CustomerID,
			appSpaceIDKey:     agent.AppSpaceID,
			applicationIDKey:  agent.ApplicationID,
			"id":              agent.ID,
			nameKey:           agent.Name,
			displayNameKey:    agent.DisplayName,
			descriptionKey:    agent.Description,
			apiPermissionsKey: agent.APIPermissions,
		})
	}
	setData(&d, data, "app_agents", allApplicationAgents)

	id := appSpaceID + "/app_agents/" + strings.Join(match, ",")
	data.SetId(id)
	return d
}
