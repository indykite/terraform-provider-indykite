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

func resourceApplicationAgent() *schema.Resource {
	return &schema.Resource{
		CreateContext: resAppAgentCreate,
		ReadContext:   resAppAgentRead,
		UpdateContext: resAppAgentUpdate,
		DeleteContext: resAppAgentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			customerIDKey:         setComputed(customerIDSchema()),
			appSpaceIDKey:         setComputed(appSpaceIDSchema()),
			applicationIDKey:      applicationIDSchema(),
			nameKey:               nameSchema(),
			displayNameKey:        displayNameSchema(),
			descriptionKey:        descriptionSchema(),
			createTimeKey:         createTimeSchema(),
			updateTimeKey:         updateTimeSchema(),
			deletionProtectionKey: deletionProtectionSchema(),
		},
	}
}

func resAppAgentCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}

	name := data.Get(nameKey).(string)
	resp, err := client.Client().CreateApplicationAgent(ctx, &config.CreateApplicationAgentRequest{
		ApplicationId: data.Get(applicationIDKey).(string),
		Name:          name,
		DisplayName:   optionalString(data, displayNameKey),
		Description:   optionalString(data, descriptionKey),
	})
	if hasFailed(&d, err, "error creating ApplicationAgent for %q", name) {
		return d
	}
	data.SetId(resp.Id)

	return resAppAgentRead(ctx, data, meta)
}

func resAppAgentRead(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	resp, err := client.Client().ReadApplicationAgent(ctx, &config.ReadApplicationAgentRequest{
		Identifier: &config.ReadApplicationAgentRequest_Id{
			Id: data.Id(),
		}})
	if err != nil {
		return diag.FromErr(err)
	}

	if resp == nil {
		return diag.Errorf("empty ApplicationAgent response")
	}

	data.SetId(resp.ApplicationAgent.Id)
	Set(&d, data, customerIDKey, resp.ApplicationAgent.CustomerId)
	Set(&d, data, appSpaceIDKey, resp.ApplicationAgent.AppSpaceId)
	Set(&d, data, applicationIDKey, resp.ApplicationAgent.ApplicationId)
	Set(&d, data, nameKey, resp.ApplicationAgent.Name)
	Set(&d, data, displayNameKey, resp.ApplicationAgent.DisplayName)
	Set(&d, data, descriptionKey, resp.ApplicationAgent.Description)
	Set(&d, data, createTimeKey, resp.ApplicationAgent.CreateTime)
	Set(&d, data, updateTimeKey, resp.ApplicationAgent.UpdateTime)
	return d
}

func resAppAgentUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)

	// If only change in plan is delete_protection, just ignore the request
	if !data.HasChangeExcept(deletionProtectionKey) {
		return d
	}

	req := &config.UpdateApplicationAgentRequest{
		Id:          data.Id(),
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	if client == nil {
		return d
	}
	_, err := client.Client().UpdateApplicationAgent(ctx, req)
	if hasFailed(&d, err, "Error while updating ApplicationAgent #%s", data.Id()) {
		return d
	}
	return resAppAgentRead(ctx, data, meta)
}

func resAppAgentDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
	client := fromMeta(&d, meta)
	if client == nil {
		return d
	}
	if hasDeleteProtection(&d, data) {
		return d
	}
	_, err := client.Client().DeleteApplicationAgent(ctx, &config.DeleteApplicationAgentRequest{
		Id: data.Id(),
	})
	if hasFailed(&d, err, "Error while deleting ApplicationAgent #%s", data.Id()) {
		return d
	}
	return d
}
