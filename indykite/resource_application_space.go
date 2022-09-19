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
	config "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
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

	name := data.Get(nameKey).(string)
	resp, err := client.Client().CreateApplicationSpace(ctx, &config.CreateApplicationSpaceRequest{
		CustomerId:  data.Get(customerIDKey).(string),
		Name:        name,
		DisplayName: optionalString(data, displayNameKey),
		Description: optionalString(data, descriptionKey),
	})
	if hasFailed(&d, err, "error creating application space for %q", name) {
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
	resp, err := client.Client().ReadApplicationSpace(ctx, &config.ReadApplicationSpaceRequest{
		Identifier: &config.ReadApplicationSpaceRequest_Id{
			Id: data.Id(),
		}})
	if err != nil {
		return diag.FromErr(err)
	}
	if resp == nil {
		return diag.Errorf("empty ApplicationSpace response")
	}

	data.SetId(resp.AppSpace.Id)
	Set(&d, data, customerIDKey, resp.AppSpace.CustomerId)
	Set(&d, data, nameKey, resp.AppSpace.Name)
	Set(&d, data, displayNameKey, resp.AppSpace.DisplayName)
	Set(&d, data, descriptionKey, resp.AppSpace.Description)
	Set(&d, data, issuerIDKey, resp.AppSpace.IssuerId)
	Set(&d, data, createTimeKey, resp.AppSpace.CreateTime)
	Set(&d, data, updateTimeKey, resp.AppSpace.UpdateTime)
	return d
}

func resAppSpaceUpdateContext(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)

	// If only change in plan is delete_protection, just ignore the request
	if !data.HasChangeExcept(deletionProtectionKey) {
		return d
	}

	req := &config.UpdateApplicationSpaceRequest{
		Id:          data.Id(),
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	if client == nil {
		return d
	}
	_, err := client.Client().UpdateApplicationSpace(ctx, req)
	if hasFailed(&d, err, "Error while updating application space #%s", data.Id()) {
		return d
	}
	return resAppSpaceReadContext(ctx, data, meta)
}

func resAppSpaceDeleteContext(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	if hasDeleteProtection(&d, data) {
		return d
	}
	_, err := client.Client().DeleteApplicationSpace(ctx, &config.DeleteApplicationSpaceRequest{
		Id: data.Id(),
	})
	if hasFailed(&d, err, "Error while deleting application space #%s", data.Id()) {
		return d
	}
	return d
}
