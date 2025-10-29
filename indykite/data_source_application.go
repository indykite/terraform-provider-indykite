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
			"Each application is created in an ApplicationSpace or project, and has a profile, " +
			"an application agent and application agent credentials. ",
		ReadContext: dataApplicationReadContext,
		Schema: map[string]*schema.Schema{
			customerIDKey:    setComputed(customerIDSchema()),
			applicationIDKey: setExactlyOneOf(applicationIDSchema(), applicationIDKey, oneOfAppID),
			nameKey:          setRequiredWith(setExactlyOneOf(nameSchema(), nameKey, oneOfAppID), appSpaceIDKey),
			appSpaceIDKey:    setRequiredWith(appSpaceIDSchema(), nameKey),
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
			appSpaceIDKey: appSpaceIDSchema(),
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
) (ApplicationResponse, diag.Diagnostic) {
	appSpaceID := data.Get(appSpaceIDKey).(string)
	var listResp ListApplicationsResponse
	err := clientCtx.GetClient().Get(ctx, "/applications?appSpaceId="+appSpaceID, &listResp)
	if err != nil {
		return ApplicationResponse{}, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("failed to list applications: %v", err),
		}
	}

	// Find application by name
	for i := range listResp.Applications {
		if listResp.Applications[i].Name == name {
			return listResp.Applications[i], diag.Diagnostic{}
		}
	}

	return ApplicationResponse{}, diag.Diagnostic{
		Severity: diag.Error,
		Summary: fmt.Sprintf(
			"application with name '%s' not found in app space '%s'",
			name, appSpaceID),
	}
}

func dataApplicationReadContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}

	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp ApplicationResponse
	var err error

	// Check if we have an ID or need to look up by name
	if id, ok := data.GetOk(applicationIDKey); ok {
		// Direct lookup by ID
		err = clientCtx.GetClient().Get(ctx, "/applications/"+id.(string), &resp)
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

	appSpaceID := data.Get(appSpaceIDKey).(string)
	var resp ListApplicationsResponse
	err := clientCtx.GetClient().Get(ctx, "/applications?appSpaceId="+appSpaceID, &resp)
	if HasFailed(&d, err) {
		return d
	}

	allApplications := make([]map[string]any, 0, len(resp.Applications))
	for i := range resp.Applications {
		app := &resp.Applications[i]
		// Apply filter if specified
		if len(match) > 0 {
			matchFound := false
			for _, filter := range match {
				if strings.Contains(app.Name, filter) {
					matchFound = true
					break
				}
			}
			if !matchFound {
				continue
			}
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
