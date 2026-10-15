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
	entityMatchingPropertyMappingStatusKey         = "property_mapping_status"
	entityMatchingPropertyMappingMessageKey        = "property_mapping_message"
	entityMatchingStatusKey                        = "entity_matching_status"
	entityMatchingMessageKey                       = "entity_matching_message"
	entityMatchingLastRunTimeKey                   = "last_run_time"
	entityMatchingMatchedEntitiesKey               = "matched_entities"
	entityMatchingReportURLKey                     = "report_url"
	entityMatchingReportTypeKey                    = "report_type"
	entityMatchingPropertyMappingsKey              = "property_mappings"
	entityMatchingSourceNodeTypeKey                = "source_node_type"
	entityMatchingSourceNodePropertyKey            = "source_node_property"
	entityMatchingTargetNodeTypeKey                = "target_node_type"
	entityMatchingTargetNodePropertyKey            = "target_node_property"
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
			StateContext: projectStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:   locationSchema(),
			projectIDKey:  projectIDSchema(),
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
			entityMatchingPropertyMappingStatusKey: computedStringSchema(
				"Status of the pipeline step that maps the properties of the node types."),
			entityMatchingPropertyMappingMessageKey: computedStringSchema(
				"Error message of the property mapping step, explaining its status."),
			entityMatchingStatusKey: computedStringSchema(
				"Status of the pipeline step that matches node entities."),
			entityMatchingMessageKey: computedStringSchema(
				"Error message of the entity matching step, explaining its status."),
			entityMatchingLastRunTimeKey: computedStringSchema("Time when the pipeline was last run."),
			entityMatchingMatchedEntitiesKey: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of entities the pipeline matched in its last run.",
			},
			entityMatchingReportURLKey:  computedStringSchema("Location of the analysis report."),
			entityMatchingReportTypeKey: computedStringSchema("Format in which the analysis report is stored."),
			entityMatchingPropertyMappingsKey: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Rules the pipeline uses to match source nodes with target nodes.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					entityMatchingSourceNodeTypeKey:     computedStringSchema("Type of the source nodes."),
					entityMatchingSourceNodePropertyKey: computedStringSchema("Property of the source nodes compared."),
					entityMatchingTargetNodeTypeKey:     computedStringSchema("Type of the target nodes."),
					entityMatchingTargetNodePropertyKey: computedStringSchema("Property of the target nodes compared."),
					entityMatchingPipelineSimilarityScoreCutOffKey: {
						Type:        schema.TypeFloat,
						Computed:    true,
						Description: "Threshold above which entities are matched by this rule.",
					},
				}},
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
		ProjectID:   projectIDFromData(data),
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
	setProjectIDData(&d, data, resp.AppSpaceID, resp.CustomerID)

	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)

	if resp.NodeFilter != nil {
		setData(&d, data, entityMatchingPipelineSourceNodeFilterKey, resp.NodeFilter.SourceNodeTypes)
		setData(&d, data, entityMatchingPipelineTargetNodeFilterKey, resp.NodeFilter.TargetNodeTypes)
	}

	setData(&d, data, entityMatchingPipelineSimilarityScoreCutOffKey, roundScore(resp.SimilarityScoreCutoff))

	setData(&d, data, entityMatchingPipelineRerunInterval, resp.RerunInterval)

	setData(&d, data, entityMatchingPropertyMappingStatusKey, resp.PropertyMappingStatus)
	setData(&d, data, entityMatchingPropertyMappingMessageKey, resp.PropertyMappingMessage)
	setData(&d, data, entityMatchingStatusKey, resp.EntityMatchingStatus)
	setData(&d, data, entityMatchingMessageKey, resp.EntityMatchingMessage)
	setData(&d, data, entityMatchingLastRunTimeKey, resp.LastRunTime)
	setData(&d, data, entityMatchingMatchedEntitiesKey, resp.MatchedEntities)
	setData(&d, data, entityMatchingReportURLKey, resp.ReportURL)
	setData(&d, data, entityMatchingReportTypeKey, resp.ReportType)
	mappings := make([]map[string]any, 0, len(resp.PropertyMappings))
	for _, m := range resp.PropertyMappings {
		mappings = append(mappings, map[string]any{
			entityMatchingSourceNodeTypeKey:                m.SourceNodeType,
			entityMatchingSourceNodePropertyKey:            m.SourceNodeProperty,
			entityMatchingTargetNodeTypeKey:                m.TargetNodeType,
			entityMatchingTargetNodePropertyKey:            m.TargetNodeProperty,
			entityMatchingPipelineSimilarityScoreCutOffKey: roundScore(m.SimilarityScoreCutoff),
		})
	}
	setData(&d, data, entityMatchingPropertyMappingsKey, mappings)

	return d
}

// roundScore rounds a float32 score to 4 decimal places for float64 compatibility.
func roundScore(score float32) float64 {
	const ratio = 10000
	return math.Round(float64(score)*ratio) / ratio
}

func resEntityMatchingPipelineUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	if hasNameChanged(&d, data) {
		return d
	}

	// Rerun interval is always sent: the API clears it when it is omitted from an update.
	interval := data.Get(entityMatchingPipelineRerunInterval).(string)
	req := UpdateEntityMatchingPipelineRequest{
		DisplayName:           updateOptionalString(data, displayNameKey),
		Description:           updateOptionalString(data, descriptionKey),
		RerunInterval:         &interval,
		SimilarityScoreCutoff: float32(data.Get(entityMatchingPipelineSimilarityScoreCutOffKey).(float64)),
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
