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
	"math"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	entityMatchingPipelineSourceNodeFilterKey      = "source_node_filter"
	entityMatchingPipelineTargetNodeFilterKey      = "target_node_filter"
	entityMatchingPipelineSimilarityScoreCutOffKey = "similarity_score_cutoff"
	entityMatchingPipelineRerunInterval            = "rerun_interval"
)

func resourceEntityMatchingPipeline() *schema.Resource {
	return &schema.Resource{
		Description: "The EntityMatchingPipeline facilitates the setup of a configuration to detect " +
			"and match identical nodes in the Identity Knowledge Graph. ",

		CreateContext: resEntityMatchingPipelineCreate,
		ReadContext:   resEntityMatchingPipelineRead,
		UpdateContext: resEntityMatchingPipelineUpdate,
		DeleteContext: resEntityMatchingPipelineDelete,
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

			entityMatchingPipelineSourceNodeFilterKey: {
				Type:        schema.TypeList,
				Description: "List of source node types to be used in the entity matching pipeline.",
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
			},
			entityMatchingPipelineTargetNodeFilterKey: {
				Type:        schema.TypeList,
				Description: "List of target node types to be used in the entity matching pipeline.",
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
			},
			entityMatchingPipelineSimilarityScoreCutOffKey: {
				Type:         schema.TypeFloat,
				Description:  "Similarity score cutoff to be used in the entity matching pipeline. Defaults to 0.5 if not specified.",
				Optional:     true,
				Default:      0.5,
				ValidateFunc: validation.FloatBetween(0, 1),
			},
			entityMatchingPipelineRerunInterval: {
				Type:        schema.TypeString,
				Description: "RerunInterval is the time between scheduled re-runs.",
				Optional:    true,
			},
		},
	}
}

func resEntityMatchingPipelineCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	req := CreateEntityMatchingPipelineRequest{
		ProjectID:   data.Get(locationKey).(string),
		Name:        data.Get(nameKey).(string),
		DisplayName: stringValue(optionalString(data, displayNameKey)),
		Description: stringValue(optionalString(data, descriptionKey)),
		NodeFilter: &EntityMatchingNodeFilter{
			SourceNodeTypes: rawArrayToTypedArray[string](data.Get(entityMatchingPipelineSourceNodeFilterKey)),
			TargetNodeTypes: rawArrayToTypedArray[string](data.Get(entityMatchingPipelineTargetNodeFilterKey)),
		},
		SimilarityScoreCutoff: float32(data.Get(entityMatchingPipelineSimilarityScoreCutOffKey).(float64)),
	}

	if rerunInterval, ok := data.GetOk(entityMatchingPipelineRerunInterval); ok {
		req.RerunInterval = rerunInterval.(string)
	}

	var resp EntityMatchingPipelineResponse
	err := clientCtx.GetClient().Post(ctx, "/entity-matching-pipelines", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

	return resEntityMatchingPipelineRead(ctx, data, meta)
}

func resEntityMatchingPipelineRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp EntityMatchingPipelineResponse
	// Support both ID and name?location=parent_id formats
	path := buildReadPath("/entity-matching-pipelines", data)
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

	if resp.NodeFilter != nil {
		setData(&d, data, entityMatchingPipelineSourceNodeFilterKey, resp.NodeFilter.SourceNodeTypes)
		setData(&d, data, entityMatchingPipelineTargetNodeFilterKey, resp.NodeFilter.TargetNodeTypes)
	}

	// Round float32 to 4 decimal places for float64 compatibility
	var ratio float64 = 10000
	setData(&d, data, entityMatchingPipelineSimilarityScoreCutOffKey,
		math.Round(float64(resp.SimilarityScoreCutoff)*ratio)/ratio)

	setData(&d, data, entityMatchingPipelineRerunInterval, resp.RerunInterval)

	return d
}

func resEntityMatchingPipelineUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := UpdateEntityMatchingPipelineRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	if data.HasChange(entityMatchingPipelineSimilarityScoreCutOffKey) {
		score := float32(data.Get(entityMatchingPipelineSimilarityScoreCutOffKey).(float64))
		req.SimilarityScoreCutoff = &score
	}

	if data.HasChange(entityMatchingPipelineRerunInterval) {
		interval := data.Get(entityMatchingPipelineRerunInterval).(string)
		req.RerunInterval = &interval
	}

	var resp EntityMatchingPipelineResponse
	err := clientCtx.GetClient().Put(ctx, "/entity-matching-pipelines/"+data.Id(), req, &resp)
	if HasFailed(&d, err) {
		return d
	}

	return resEntityMatchingPipelineRead(ctx, data, meta)
}

func resEntityMatchingPipelineDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	err := clientCtx.GetClient().Delete(ctx, "/entity-matching-pipelines/"+data.Id())
	HasFailed(&d, err)
	return d
}
