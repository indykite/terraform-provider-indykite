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

func dataSourceAppSpace() *schema.Resource {
	appSpaceIdentifier := []string{nameKey, appSpaceIDKey}
	return &schema.Resource{
		Description: "It is workspace or environment for your applications.  ",
		ReadContext: dataAppSpaceReadContext,
		Schema: map[string]*schema.Schema{
			appSpaceIDKey:  setExactlyOneOf(appSpaceIDSchema(), appSpaceIDKey, appSpaceIdentifier),
			nameKey:        setRequiredWith(setExactlyOneOf(nameSchema(), nameKey, appSpaceIdentifier), customerIDKey),
			customerIDKey:  setRequiredWith(customerIDSchema(), nameKey),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),
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
						customerIDKey:  setComputed(customerIDSchema()),
						"id":           setComputed(appSpaceIDSchema()),
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

func dataAppSpaceReadContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}

	req := &configpb.ReadApplicationSpaceRequest{}
	if name, exists := data.GetOk(nameKey); exists {
		req.Identifier = &configpb.ReadApplicationSpaceRequest_Name{
			Name: &configpb.UniqueNameIdentifier{
				Name:     name.(string),
				Location: data.Get(customerIDKey).(string),
			},
		}
	} else if id, ok := data.GetOk(appSpaceIDKey); ok {
		req.Identifier = &configpb.ReadApplicationSpaceRequest_Id{
			Id: id.(string),
		}
	}

	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadApplicationSpace(ctx, req)
	if readHasFailed(&d, err, data) {
		return d
	}

	return dataAppSpaceFlatten(data, resp.GetAppSpace())
}

func dataAppSpaceFlatten(data *schema.ResourceData, resp *configpb.ApplicationSpace) diag.Diagnostics {
	var d diag.Diagnostics
	if resp == nil {
		return diag.Diagnostics{buildPluginError("empty ApplicationSpace response")}
	}
	data.SetId(resp.GetId())
	setData(&d, data, customerIDKey, resp.GetCustomerId())
	setData(&d, data, nameKey, resp.GetName())
	setData(&d, data, displayNameKey, resp.GetDisplayName())
	setData(&d, data, descriptionKey, resp.GetDescription())
	setData(&d, data, createTimeKey, resp.GetCreateTime())
	setData(&d, data, updateTimeKey, resp.GetUpdateTime())
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
	resp, err := clientCtx.GetClient().ListApplicationSpaces(ctx, &configpb.ListApplicationSpacesRequest{
		CustomerId: data.Get(customerIDKey).(string),
		Match:      match,
	})
	if HasFailed(&d, err) {
		return d
	}

	var allAppSpaces []map[string]any
	for {
		appSpace, err := resp.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			HasFailed(&d, err)
			return d
		}
		allAppSpaces = append(allAppSpaces, map[string]any{
			customerIDKey:  appSpace.GetAppSpace().GetCustomerId(),
			"id":           appSpace.GetAppSpace().GetId(),
			nameKey:        appSpace.GetAppSpace().GetName(),
			displayNameKey: appSpace.GetAppSpace().GetDisplayName(),
			descriptionKey: flattenOptionalString(appSpace.GetAppSpace().GetDescription()),
		})
	}
	setData(&d, data, "app_spaces", allAppSpaces)

	id := data.Get(customerIDKey).(string) + "/appSpaces/" + strings.Join(match, ",")
	data.SetId(id)
	return d
}
