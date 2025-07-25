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
	"errors"
	"io"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
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

func dataAppAgentReadContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}

	req := &configpb.ReadApplicationAgentRequest{}
	if name, exists := data.GetOk(nameKey); exists {
		req.Identifier = &configpb.ReadApplicationAgentRequest_Name{
			Name: &configpb.UniqueNameIdentifier{
				Name:     name.(string),
				Location: data.Get(appSpaceIDKey).(string),
			},
		}
	} else if id, ok := data.GetOk(appAgentIDKey); ok {
		req.Identifier = &configpb.ReadApplicationAgentRequest_Id{
			Id: id.(string),
		}
	}

	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadApplicationAgent(ctx, req)
	if readHasFailed(&d, err, data) {
		return d
	}

	if resp.GetApplicationAgent() == nil {
		return diag.Diagnostics{buildPluginError("empty ApplicationAgent response")}
	}

	data.SetId(resp.GetApplicationAgent().GetId())
	setData(&d, data, customerIDKey, resp.GetApplicationAgent().GetCustomerId())
	setData(&d, data, appSpaceIDKey, resp.GetApplicationAgent().GetAppSpaceId())
	setData(&d, data, applicationIDKey, resp.GetApplicationAgent().GetApplicationId())
	setData(&d, data, nameKey, resp.GetApplicationAgent().GetName())
	setData(&d, data, displayNameKey, resp.GetApplicationAgent().GetDisplayName())
	setData(&d, data, descriptionKey, resp.GetApplicationAgent().GetDescription())
	setData(&d, data, createTimeKey, resp.GetApplicationAgent().GetCreateTime())
	setData(&d, data, updateTimeKey, resp.GetApplicationAgent().GetUpdateTime())
	setData(&d, data, apiPermissionsKey, resp.GetApplicationAgent().GetApiAccessRestriction())
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
	resp, err := clientCtx.GetClient().ListApplicationAgents(ctx, &configpb.ListApplicationAgentsRequest{
		AppSpaceId: data.Get(appSpaceIDKey).(string),
		Match:      match,
	})
	if HasFailed(&d, err) {
		return d
	}

	var allApplicationAgents []map[string]any
	for {
		app, err := resp.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			HasFailed(&d, err)
			return d
		}
		allApplicationAgents = append(allApplicationAgents, map[string]any{
			customerIDKey:     app.GetApplicationAgent().GetCustomerId(),
			appSpaceIDKey:     app.GetApplicationAgent().GetAppSpaceId(),
			applicationIDKey:  app.GetApplicationAgent().GetApplicationId(),
			"id":              app.GetApplicationAgent().GetId(),
			nameKey:           app.GetApplicationAgent().GetName(),
			displayNameKey:    app.GetApplicationAgent().GetDisplayName(),
			descriptionKey:    flattenOptionalString(app.GetApplicationAgent().GetDescription()),
			apiPermissionsKey: app.GetApplicationAgent().GetApiAccessRestriction(),
		})
	}
	setData(&d, data, "app_agents", allApplicationAgents)

	id := data.Get(appSpaceIDKey).(string) + "/app_agents/" + strings.Join(match, ",")
	data.SetId(id)
	return d
}
