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

func dataApplicationReadContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}

	req := &configpb.ReadApplicationRequest{}
	if name, exists := data.GetOk(nameKey); exists {
		req.Identifier = &configpb.ReadApplicationRequest_Name{
			Name: &configpb.UniqueNameIdentifier{
				Name:     name.(string),
				Location: data.Get(appSpaceIDKey).(string),
			},
		}
	} else if id, ok := data.GetOk(applicationIDKey); ok {
		req.Identifier = &configpb.ReadApplicationRequest_Id{
			Id: id.(string),
		}
	}

	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadApplication(ctx, req)
	if readHasFailed(&d, err, data) {
		return d
	}

	if resp.GetApplication() == nil {
		return diag.Diagnostics{buildPluginError("empty Application response")}
	}

	data.SetId(resp.GetApplication().GetId())
	setData(&d, data, customerIDKey, resp.GetApplication().GetCustomerId())
	setData(&d, data, appSpaceIDKey, resp.GetApplication().GetAppSpaceId())
	setData(&d, data, nameKey, resp.GetApplication().GetName())
	setData(&d, data, displayNameKey, resp.GetApplication().GetDisplayName())
	setData(&d, data, descriptionKey, resp.GetApplication().GetDescription())
	setData(&d, data, createTimeKey, resp.GetApplication().GetCreateTime())
	setData(&d, data, updateTimeKey, resp.GetApplication().GetUpdateTime())
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
	resp, err := clientCtx.GetClient().ListApplications(ctx, &configpb.ListApplicationsRequest{
		AppSpaceId: data.Get(appSpaceIDKey).(string),
		Match:      match,
	})
	if HasFailed(&d, err) {
		return d
	}

	var allApplications []map[string]any
	for {
		app, err := resp.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			HasFailed(&d, err)
			return d
		}
		allApplications = append(allApplications, map[string]any{
			customerIDKey:  app.GetApplication().GetCustomerId(),
			appSpaceIDKey:  app.GetApplication().GetAppSpaceId(),
			"id":           app.GetApplication().GetId(),
			nameKey:        app.GetApplication().GetName(),
			displayNameKey: app.GetApplication().GetDisplayName(),
			descriptionKey: flattenOptionalString(app.GetApplication().GetDescription()),
		})
	}
	setData(&d, data, "applications", allApplications)

	id := data.Get(appSpaceIDKey).(string) + "/apps/" + strings.Join(match, ",")
	data.SetId(id)
	return d
}
