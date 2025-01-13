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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

const (
	entityMatchingPipelineSourceNodeFilterKey      = "source_node_filter"
	entityMatchingPipelineTargetNodeFilterKey      = "target_node_filter"
	entityMatchingPipelineSimilarityScoreCutOffKey = "similarity_score_cutoff"
	entityMatchingPipelineRerunInterval            = "rerun_interval"
)

func resourceEntityMatchingPipeline() *schema.Resource {
	readContext := configReadContextFunc(resourceEntityMatchingPipelineFlatten)

	return &schema.Resource{
		Description: "The EntityMatchingPipeline facilitates the setup of a configuration to detect " +
			"and match identical nodes in the Identity Knowledge Graph. ",

		CreateContext: configCreateContextFunc(resourceEntityMatchingPipelineBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceEntityMatchingPipelineBuild, readContext),
		DeleteContext: configDeleteContextFunc(),
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
				Description:  "Similarity score cutoff to be used in the entity matching pipeline.",
				Optional:     true,
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

func resourceEntityMatchingPipelineFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) diag.Diagnostics {
	var d diag.Diagnostics
	entitymatching := resp.GetConfigNode().GetEntityMatchingPipelineConfig()

	sourceTypes := make([]string, len(entitymatching.GetNodeFilter().GetSourceNodeTypes()))
	for i, source := range entitymatching.GetNodeFilter().GetSourceNodeTypes() {
		sourceTypes[i] = source
	}
	setData(&d, data, entityMatchingPipelineSourceNodeFilterKey, sourceTypes)

	targetTypes := make([]string, len(entitymatching.GetNodeFilter().GetTargetNodeTypes()))
	for i, target := range entitymatching.GetNodeFilter().GetTargetNodeTypes() {
		targetTypes[i] = target
	}
	setData(&d, data, entityMatchingPipelineTargetNodeFilterKey, targetTypes)

	similarityScoreCutoff := entitymatching.GetSimilarityScoreCutoff()
	setData(&d, data, entityMatchingPipelineSimilarityScoreCutOffKey, float32(similarityScoreCutoff))

	rerunInterval := entitymatching.GetRerunInterval()
	setData(&d, data, entityMatchingPipelineRerunInterval, rerunInterval)

	return d
}

func resourceEntityMatchingPipelineBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	cfg := &configpb.EntityMatchingPipelineConfig{}

	sourceNodeTypes := data.Get(entityMatchingPipelineSourceNodeFilterKey)
	targetNodeTypes := data.Get(entityMatchingPipelineTargetNodeFilterKey)

	cfg.NodeFilter = &configpb.EntityMatchingPipelineConfig_NodeFilter{
		SourceNodeTypes: rawArrayToTypedArray[string](sourceNodeTypes),
		TargetNodeTypes: rawArrayToTypedArray[string](targetNodeTypes),
	}
	// Check if SimilarityScoreCutOffKey is available
	if similarityScoreCutoff, hasSimilarityScore := data.GetOk(
		entityMatchingPipelineSimilarityScoreCutOffKey); hasSimilarityScore {
		cfg.SimilarityScoreCutoff = similarityScoreCutoff.(float32)
	}
	// Check if RerunInterval is available
	if rerunInterval, hasRerunInterval := data.GetOk(
		entityMatchingPipelineRerunInterval); hasRerunInterval {
		cfg.RerunInterval = rerunInterval.(string)
	}
	builder.WithEntityMatchingPipelineConfig(cfg)
}
