// Copyright (c) 2024 IndyKite
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
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

//nolint:gosec // there are no secrets
const (
	ingestPipelineSourcesTypeKey       = "sources"
	ingestPipelineOperationsTypeKey    = "operations"
	ingestPipelineAppAgentTokenTypeKey = "app_agent_token"
)

var (
	ingestPipelineAppAgentTokenRegex = regexp.MustCompile(`^[A-Za-z0-9-_]+?\.[A-Za-z0-9-_]+?\.[A-Za-z0-9-_]+?$`)
)

func resourceIngestPipeline() *schema.Resource {
	return &schema.Resource{
		Description: `Ingest pipeline configuration adds support for 3rd party data sources, which can be used to ingest data to the IndyKite system. The configuration also allows you to define the allowed operations in the pipeline.`,

		CreateContext: resIngestPipelineCreate,
		ReadContext:   resIngestPipelineRead,
		UpdateContext: resIngestPipelineUpdate,
		DeleteContext: resIngestPipelineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:   locationSchema(),
			customerIDKey: setComputed(customerIDSchema()),
			appSpaceIDKey: setComputed(appSpaceIDSchema()),

			nameKey:        nameSchema(),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),

			ingestPipelineSourcesTypeKey: {
				Type:        schema.TypeList,
				Description: "List of sources to be used in the ingest pipeline.",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
				MaxItems: 10,
			},
			ingestPipelineOperationsTypeKey: {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Deprecated:  "This field is deprecated and is not used anymore.",
				Description: "List of operations is no longer used. All previously saved values are ignored.",
			},
			ingestPipelineAppAgentTokenTypeKey: {
				Type:        schema.TypeString,
				Description: "Application agent token is used to identify the application space in IndyKite APIs.",
				Sensitive:   true,
				Required:    true,
				ValidateFunc: validation.StringMatch(
					ingestPipelineAppAgentTokenRegex, "must be valid application agent token.",
				),
			},
		},
	}
}

func resIngestPipelineCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	req := CreateIngestPipelineRequest{
		ProjectID:     data.Get(locationKey).(string),
		Name:          data.Get(nameKey).(string),
		DisplayName:   stringValue(optionalString(data, displayNameKey)),
		Description:   stringValue(optionalString(data, descriptionKey)),
		Sources:       rawArrayToTypedArray[string](data.Get(ingestPipelineSourcesTypeKey).([]any)),
		AppAgentToken: data.Get(ingestPipelineAppAgentTokenTypeKey).(string),
	}

	var resp IngestPipelineResponse
	err := clientCtx.GetClient().Post(ctx, "/ingest-pipelines", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

	return resIngestPipelineRead(ctx, data, meta)
}

func resIngestPipelineRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp IngestPipelineResponse
	// Support both ID and name?location=parent_id formats
	path := buildReadPath("/ingest-pipelines", data)
	err := clientCtx.GetClient().Get(ctx, path, &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceID)

	// Set location based on which is present
	if resp.AppSpaceID != "" {
		setData(&d, data, locationKey, resp.AppSpaceID)
	} else if resp.CustomerID != "" {
		setData(&d, data, locationKey, resp.CustomerID)
	}

	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, ingestPipelineSourcesTypeKey, resp.Sources)

	return d
}

func resIngestPipelineUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := UpdateIngestPipelineRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	if data.HasChange(ingestPipelineSourcesTypeKey) {
		req.Sources = rawArrayToTypedArray[string](data.Get(ingestPipelineSourcesTypeKey).([]any))
	}

	if data.HasChange(ingestPipelineAppAgentTokenTypeKey) {
		appAgentToken := data.Get(ingestPipelineAppAgentTokenTypeKey).(string)
		req.AppAgentToken = &appAgentToken
	}

	var resp IngestPipelineResponse
	err := clientCtx.GetClient().Put(ctx, "/ingest-pipelines/"+data.Id(), req, &resp)
	if HasFailed(&d, err) {
		return d
	}

	return resIngestPipelineRead(ctx, data, meta)
}

func resIngestPipelineDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	err := clientCtx.GetClient().Delete(ctx, "/ingest-pipelines/"+data.Id())
	HasFailed(&d, err)
	return d
}
