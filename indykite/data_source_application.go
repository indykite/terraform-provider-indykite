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

func dataSourceApplication() *schema.Resource {
	oneOfAppID := []string{nameKey, applicationIDKey}

	return &schema.Resource{
		Description: "An application represents the center of the solution, " +
			"and is also the legal entity users legally interact with. " +
			"Each application is created in an ApplicationSpace, and has a profile, " +
			"an application agent and application agent credentials. ",
		ReadContext: dataApplicationReadContext,
		Schema: map[string]*schema.Schema{
			customerIDKey:    setComputed(customerIDSchema()),
			applicationIDKey: setExactlyOneOf(applicationIDSchema(), applicationIDKey, oneOfAppID),
			nameKey:          setExactlyOneOf(nameSchema(), nameKey, oneOfAppID),
			appSpaceIDKey:    convertToOptional(appSpaceIDSchema()),
			displayNameKey:   displayNameSchema(),
			descriptionKey:   descriptionSchema(),
			createTimeKey:    createTimeSchema(),
			updateTimeKey:    updateTimeSchema(),
		},
		Timeouts: defaultDataTimeouts(),
	}
}

func dataSourceApplicationList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataApplicationListContext,
		Schema: map[string]*schema.Schema{
			appSpaceIDKey: appSpaceIDSchema(), // User-facing field - ONLY this should be in configs
			filterKey:     exactNameFilterSchema(),
			"applications": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						customerIDKey:  setComputed(customerIDSchema()),
						appSpaceIDKey:  setComputed(appSpaceIDSchema()),
						"id":           setComputed(applicationIDSchema()),
						nameKey:        nameSchema(),
						displayNameKey: displayNameSchema(),
						descriptionKey: descriptionSchema(),
					},
				},
			},
		},
		Timeouts: defaultDataTimeouts(),
	}
}

// lookupApplicationByName finds an application by name within an app space.
func lookupApplicationByName(
	ctx context.Context,
	clientCtx *ClientContext,
	data *schema.ResourceData,
	name string,
) (*ApplicationResponse, diag.Diagnostic) {
	appSpaceID := data.Get(appSpaceIDKey).(string)

	// Validate that app_space_id is provided when using name
	if appSpaceID == "" {
		return nil, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "\"name\": all of `app_space_id, name` must be specified",
		}
	}

	resp := &ApplicationResponse{}
	err := clientCtx.GetClient().Get(ctx, "/applications/"+name+"?location="+appSpaceID, resp)
	if err != nil {
		return nil, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("failed to get application by name: %v", err),
		}
	}

	return resp, diag.Diagnostic{}
}

func dataApplicationReadContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}

	var resp *ApplicationResponse
	var err error

	// Check if we have an ID or need to look up by name
	if id, ok := data.GetOk(applicationIDKey); ok {
		// Direct lookup by ID
		resp = &ApplicationResponse{}
		err = clientCtx.GetClient().Get(ctx, "/applications/"+id.(string), resp)
	} else if name, exists := data.GetOk(nameKey); exists {
		// Look up by name within app space
		var diagErr diag.Diagnostic
		resp, diagErr = lookupApplicationByName(ctx, clientCtx, data, name.(string))
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
	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	return d
}

func dataApplicationListContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
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

	// User provides app_space_id in config, we use project_id parameter for REST API
	appSpaceID := data.Get(appSpaceIDKey).(string)

	var resp ListApplicationsResponse
	err := clientCtx.GetClient().Get(ctx, "/applications?project_id="+appSpaceID+"&search=", &resp)
	if HasFailed(&d, err) {
		return d
	}

	allApplications := make([]map[string]any, 0, len(resp.Applications))
	for i := range resp.Applications {
		app := &resp.Applications[i]
		// Apply exact name match filter (MinItems: 1 ensures filter is always present)
		matchFound := false
		for _, filter := range match {
			if app.Name == filter {
				matchFound = true
				break
			}
		}
		if !matchFound {
			continue
		}

		allApplications = append(allApplications, map[string]any{
			customerIDKey:  app.CustomerID,
			appSpaceIDKey:  app.AppSpaceID,
			"id":           app.ID,
			nameKey:        app.Name,
			displayNameKey: app.DisplayName,
			descriptionKey: app.Description,
		})
	}
	setData(&d, data, "applications", allApplications)

	id := appSpaceID + "/apps/" + strings.Join(match, ",")
	data.SetId(id)
	return d
}
