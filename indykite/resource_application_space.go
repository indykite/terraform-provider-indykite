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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	config "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

func resourceApplicationSpace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resAppSpaceCreateContext,
		ReadContext:   resAppSpaceReadContext,
		UpdateContext: resAppSpaceUpdateContext,
		DeleteContext: resAppSpaceDeleteContext,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			customerIDKey:         customerIDSchema(),
			nameKey:               nameSchema(),
			displayNameKey:        displayNameSchema(),
			descriptionKey:        descriptionSchema(),
			issuerIDKey:           setComputed(issuerIDSchema()),
			createTimeKey:         createTimeSchema(),
			updateTimeKey:         updateTimeSchema(),
			deletionProtectionKey: deletionProtectionSchema(),
		},
	}
}

func resAppSpaceCreateContext(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	name := data.Get(nameKey).(string)
	resp, err := client.getClient().CreateApplicationSpace(ctx, &config.CreateApplicationSpaceRequest{
		CustomerId:  data.Get(customerIDKey).(string),
		Name:        name,
		DisplayName: optionalString(data, displayNameKey),
		Description: optionalString(data, descriptionKey),
	})
	if hasFailed(&d, err) {
		return d
	}
	data.SetId(resp.Id)

	return resAppSpaceReadContext(ctx, data, meta)
}

func resAppSpaceReadContext(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := client.getClient().ReadApplicationSpace(ctx, &config.ReadApplicationSpaceRequest{
		Identifier: &config.ReadApplicationSpaceRequest_Id{
			Id: data.Id(),
		}})
	if hasFailed(&d, err) {
		return d
	}
	if resp.GetAppSpace() == nil {
		return diag.Diagnostics{buildPluginError("empty ApplicationSpace response")}
	}

	data.SetId(resp.AppSpace.Id)
	setData(&d, data, customerIDKey, resp.AppSpace.CustomerId)
	setData(&d, data, nameKey, resp.AppSpace.Name)
	setData(&d, data, displayNameKey, resp.AppSpace.DisplayName)
	setData(&d, data, descriptionKey, resp.AppSpace.Description)
	setData(&d, data, issuerIDKey, resp.AppSpace.IssuerId)
	setData(&d, data, createTimeKey, resp.AppSpace.CreateTime)
	setData(&d, data, updateTimeKey, resp.AppSpace.UpdateTime)
	return d
}

func resAppSpaceUpdateContext(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	// If only change in plan is delete_protection, just ignore the request
	if !data.HasChangeExcept(deletionProtectionKey) {
		return d
	}

	req := &config.UpdateApplicationSpaceRequest{
		Id:          data.Id(),
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	_, err := client.getClient().UpdateApplicationSpace(ctx, req)
	if hasFailed(&d, err) {
		return d
	}
	return resAppSpaceReadContext(ctx, data, meta)
}

func resAppSpaceDeleteContext(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()
	if hasDeleteProtection(&d, data) {
		return d
	}
	_, err := client.getClient().DeleteApplicationSpace(ctx, &config.DeleteApplicationSpaceRequest{
		Id: data.Id(),
	})
	hasFailed(&d, err)
	return d
}
