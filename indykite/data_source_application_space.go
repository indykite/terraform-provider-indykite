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
	"io"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	configpb "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
)

func dataSourceAppSpace() *schema.Resource {
	appSpaceIdentifier := []string{nameKey, appSpaceIDKey}
	return &schema.Resource{
		ReadContext: dataAppSpaceReadContext,
		Schema: map[string]*schema.Schema{
			appSpaceIDKey:  setExactlyOneOf(appSpaceIDSchema(), appSpaceIDKey, appSpaceIdentifier),
			nameKey:        setRequiredWith(setExactlyOneOf(nameSchema(), nameKey, appSpaceIdentifier), customerIDKey),
			customerIDKey:  setRequiredWith(customerIDSchema(), nameKey),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
			issuerIDKey:    setComputed(issuerIDSchema()),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),
		},
		Timeouts: defaultTimeouts(),
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
						issuerIDKey:    setComputed(issuerIDSchema()),
					},
				},
			},
		},
		Timeouts: defaultTimeouts(),
	}
}

func dataAppSpaceReadContext(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	req := new(configpb.ReadApplicationSpaceRequest)
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

	client := fromMeta(&d, meta)
	if d.HasError() {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := client.Client().ReadApplicationSpace(ctx, req)
	if hasFailed(&d, err, "") {
		return d
	}

	return dataAppSpaceFlatten(data, resp.AppSpace)
}

func dataAppSpaceFlatten(data *schema.ResourceData, resp *configpb.ApplicationSpace) (d diag.Diagnostics) {
	if resp == nil {
		return diag.Errorf("empty ApplicationSpace response")
	}
	data.SetId(resp.Id)
	Set(&d, data, customerIDKey, resp.CustomerId)
	Set(&d, data, nameKey, resp.Name)
	Set(&d, data, displayNameKey, resp.DisplayName)
	Set(&d, data, descriptionKey, resp.Description)
	Set(&d, data, issuerIDKey, resp.IssuerId)
	Set(&d, data, createTimeKey, resp.CreateTime)
	Set(&d, data, updateTimeKey, resp.UpdateTime)
	return d
}

func dataAppSpaceListContext(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	rawFilter := data.Get(filterKey).([]interface{})
	match := make([]string, len(rawFilter))
	for i, v := range rawFilter {
		match[i] = v.(string)
	}

	client := fromMeta(&d, meta)
	if d.HasError() {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := client.Client().ListApplicationSpaces(ctx, &configpb.ListApplicationSpacesRequest{
		CustomerId: data.Get(customerIDKey).(string),
		Match:      match,
	})
	if hasFailed(&d, err, "") {
		return d
	}

	var allAppSpaces []map[string]interface{}
	for {
		appSpace, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return append(d, diag.FromErr(err)...)
		}
		allAppSpaces = append(allAppSpaces, map[string]interface{}{
			customerIDKey:  appSpace.GetAppSpace().GetCustomerId(),
			"id":           appSpace.GetAppSpace().GetId(),
			nameKey:        appSpace.GetAppSpace().GetName(),
			displayNameKey: appSpace.GetAppSpace().GetDisplayName(),
			descriptionKey: flattenOptionalString(appSpace.GetAppSpace().GetDescription()),
			issuerIDKey:    appSpace.GetAppSpace().GetIssuerId(),
		})
	}
	Set(&d, data, "app_spaces", allAppSpaces)

	id := data.Get(customerIDKey).(string) + "/appSpaces/" + strings.Join(match, ",")
	data.SetId(id)
	return d
}
