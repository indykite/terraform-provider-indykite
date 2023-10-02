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

func resAppAgentCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	name := data.Get(nameKey).(string)
	resp, err := clientCtx.GetClient().CreateApplicationAgent(ctx, &config.CreateApplicationAgentRequest{
		ApplicationId: data.Get(applicationIDKey).(string),
		Name:          name,
		DisplayName:   optionalString(data, displayNameKey),
		Description:   optionalString(data, descriptionKey),
		Bookmarks:     clientCtx.GetBookmarks(),
	})
	if hasFailed(&d, err) {
		return d
	}
	data.SetId(resp.Id)
	clientCtx.AddBookmarks(resp.GetBookmark())

	return resAppAgentRead(ctx, data, meta)
}

func resAppAgentRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadApplicationAgent(ctx, &config.ReadApplicationAgentRequest{
		Identifier: &config.ReadApplicationAgentRequest_Id{
			Id: data.Id(),
		},
		Bookmarks: clientCtx.GetBookmarks(),
	})
	if hasFailed(&d, err) {
		return d
	}

	if resp.GetApplicationAgent() == nil {
		return diag.Diagnostics{buildPluginError("empty ApplicationAgent response")}
	}

	data.SetId(resp.ApplicationAgent.Id)
	setData(&d, data, customerIDKey, resp.ApplicationAgent.CustomerId)
	setData(&d, data, appSpaceIDKey, resp.ApplicationAgent.AppSpaceId)
	setData(&d, data, applicationIDKey, resp.ApplicationAgent.ApplicationId)
	setData(&d, data, nameKey, resp.ApplicationAgent.Name)
	setData(&d, data, displayNameKey, resp.ApplicationAgent.DisplayName)
	setData(&d, data, descriptionKey, resp.ApplicationAgent.Description)
	setData(&d, data, createTimeKey, resp.ApplicationAgent.CreateTime)
	setData(&d, data, updateTimeKey, resp.ApplicationAgent.UpdateTime)
	return d
}

func resAppAgentUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	// If only change in plan is delete_protection, just ignore the request
	if !data.HasChangeExcept(deletionProtectionKey) {
		return d
	}

	req := &config.UpdateApplicationAgentRequest{
		Id:          data.Id(),
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
		Bookmarks:   clientCtx.GetBookmarks(),
	}

	resp, err := clientCtx.GetClient().UpdateApplicationAgent(ctx, req)
	if hasFailed(&d, err) {
		return d
	}
	clientCtx.AddBookmarks(resp.GetBookmark())
	return resAppAgentRead(ctx, data, meta)
}

func resAppAgentDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()
	if hasDeleteProtection(&d, data) {
		return d
	}
	resp, err := clientCtx.GetClient().DeleteApplicationAgent(ctx, &config.DeleteApplicationAgentRequest{
		Id:        data.Id(),
		Bookmarks: clientCtx.GetBookmarks(),
	})
	hasFailed(&d, err)
	clientCtx.AddBookmarks(resp.GetBookmark())
	return d
}
