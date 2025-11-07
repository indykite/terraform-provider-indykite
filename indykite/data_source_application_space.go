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

func dataSourceAppSpace() *schema.Resource {
	appSpaceIdentifier := []string{nameKey, appSpaceIDKey}
	return &schema.Resource{
		Description: "It is workspace or environment for your applications.  ",
		ReadContext: dataAppSpaceReadContext,
		Schema: map[string]*schema.Schema{
			appSpaceIDKey: setExactlyOneOf(appSpaceIDSchema(), appSpaceIDKey, appSpaceIdentifier),
			nameKey: setRequiredWith(
				setExactlyOneOf(
					nameSchema(),
					nameKey,
					appSpaceIdentifier),
				customerIDKey),
			customerIDKey:    setRequiredWith(customerIDSchema(), nameKey),
			displayNameKey:   displayNameSchema(),
			descriptionKey:   descriptionSchema(),
			createTimeKey:    createTimeSchema(),
			updateTimeKey:    updateTimeSchema(),
			regionKey:        setComputed(regionSchema()),
			ikgSizeKey:       ikgSizeComputedSchema(),
			replicaRegionKey: setComputed(replicaRegionSchema()),
			dbConnectionKey:  dbConnectionComputedSchema(),
		},
		Timeouts: defaultDataTimeouts(),
	}
}

func dataSourceAppSpaceList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataAppSpaceListContext,
		Schema: map[string]*schema.Schema{
			customerIDKey: customerIDSchema(),
			filterKey:     exactNameFilterSchema(),
			"app_spaces": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						customerIDKey:    setComputed(customerIDSchema()),
						"id":             setComputed(appSpaceIDSchema()),
						nameKey:          nameSchema(),
						displayNameKey:   displayNameSchema(),
						descriptionKey:   descriptionSchema(),
						regionKey:        setComputed(regionSchema()),
						ikgSizeKey:       ikgSizeComputedSchema(),
						replicaRegionKey: setComputed(replicaRegionSchema()),
						// Note: db_connection is intentionally omitted from list view for security
						// Users should query individual app spaces to get db connection details
					},
				},
			},
		},
		Timeouts: defaultDataTimeouts(),
	}
}

// lookupApplicationSpaceByName finds an application space by name within a customer.
func lookupApplicationSpaceByName(
	ctx context.Context,
	clientCtx *ClientContext,
	data *schema.ResourceData,
	name string,
) (*ApplicationSpaceResponse, diag.Diagnostic) {
	customerID := data.Get(customerIDKey).(string)
	var listResp ListApplicationSpacesResponse
	err := clientCtx.GetClient().Get(ctx, "/projects?organization_id="+customerID, &listResp)
	if err != nil {
		return nil, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("failed to list application spaces: %v", err),
		}
	}

	// Find app space by name
	for i := range listResp.AppSpaces {
		if listResp.AppSpaces[i].Name == name {
			return &listResp.AppSpaces[i], diag.Diagnostic{}
		}
	}

	return nil, diag.Diagnostic{
		Severity: diag.Error,
		Summary: fmt.Sprintf(
			"application space with name '%s' not found in customer '%s'",
			name, customerID),
	}
}

func dataAppSpaceReadContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}

	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp *ApplicationSpaceResponse
	var err error

	// Check if we have an ID or need to look up by name
	if id, ok := data.GetOk(appSpaceIDKey); ok {
		// Direct lookup by ID
		resp = &ApplicationSpaceResponse{}
		err = clientCtx.GetClient().Get(ctx, "/projects/"+id.(string), resp)
	} else if name, exists := data.GetOk(nameKey); exists {
		// Look up by name within customer
		var diagErr diag.Diagnostic
		resp, diagErr = lookupApplicationSpaceByName(ctx, clientCtx, data, name.(string))
		// Only return if there's an actual error (non-zero severity with summary)
		if diagErr.Summary != "" {
			return append(d, diagErr)
		}
	}

	if readHasFailed(&d, err, data) {
		return d
	}

	return dataAppSpaceFlatten(data, resp)
}

func dataAppSpaceFlatten(data *schema.ResourceData, resp *ApplicationSpaceResponse) diag.Diagnostics {
	var d diag.Diagnostics
	if resp == nil {
		return diag.Diagnostics{buildPluginError("empty ApplicationSpace response")}
	}
	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, regionKey, resp.Region)
	setData(&d, data, ikgSizeKey, resp.IKGSize)
	setData(&d, data, replicaRegionKey, resp.ReplicaRegion)
	setDBConnectionData(&d, data, resp.DBConnection)
	return d
}

func dataAppSpaceListContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
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

	customerID := data.Get(customerIDKey).(string)
	var resp ListApplicationSpacesResponse
	err := clientCtx.GetClient().Get(ctx, "/projects?organization_id="+customerID, &resp)
	if HasFailed(&d, err) {
		return d
	}

	allAppSpaces := make([]map[string]any, 0, len(resp.AppSpaces))
	for i := range resp.AppSpaces {
		appSpace := &resp.AppSpaces[i]
		// Apply exact name match filter (MinItems: 1 ensures filter is always present)
		matchFound := false
		for _, filter := range match {
			if appSpace.Name == filter {
				matchFound = true
				break
			}
		}
		if !matchFound {
			continue
		}

		allAppSpaces = append(allAppSpaces, map[string]any{
			customerIDKey:    appSpace.CustomerID,
			"id":             appSpace.ID,
			nameKey:          appSpace.Name,
			displayNameKey:   appSpace.DisplayName,
			descriptionKey:   appSpace.Description,
			regionKey:        appSpace.Region,
			ikgSizeKey:       appSpace.IKGSize,
			replicaRegionKey: appSpace.ReplicaRegion,
			// db_connection intentionally omitted from list view for security
		})
	}
	setData(&d, data, "app_spaces", allAppSpaces)

	id := customerID + "/appSpaces/" + strings.Join(match, ",")
	data.SetId(id)
	return d
}
