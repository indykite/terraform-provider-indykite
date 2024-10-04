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
	entityMatchingPipelineSimilarityScoreCutOffKey = "score"
)

func resourceEntityMatchingPipeline() *schema.Resource {
	readContext := configReadContextFunc(resourceEntityMatchingPipelineFlatten)

	return &schema.Resource{
		Description: "EntityMatchingPipeline is a configuration that allows run " +
			"a pipeline to create relationships from entity matching",

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
				Description: "List od source node types to be used in the entity matching pipeline.",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
			},
			entityMatchingPipelineTargetNodeFilterKey: {
				Type:        schema.TypeList,
				Description: "List of target node types to be used in the entity matching pipeline.",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
			},
			entityMatchingPipelineSimilarityScoreCutOffKey: {
				Type:         schema.TypeFloat,
				Description:  "Similarity score cutoff to be used in the entity matching pipeline.",
				Required:     true,
				ValidateFunc: validation.FloatBetween(0, 1),
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

	sourceTypes := make([]string, len(entitymatching.NodeFilter.GetSourceNodeTypes()))
	for i, source := range entitymatching.NodeFilter.GetSourceNodeTypes() {
		sourceTypes[i] = source
	}
	setData(&d, data, entityMatchingPipelineSourceNodeFilterKey, sourceTypes)
	targetTypes := make([]string, len(entitymatching.NodeFilter.GetTargetNodeTypes()))
	for i, target := range entitymatching.NodeFilter.GetTargetNodeTypes() {
		targetTypes[i] = target
	}
	setData(&d, data, entityMatchingPipelineTargetNodeFilterKey, targetTypes)
	score := entitymatching.GetSimilarityScoreCutoff()
	setData(&d, data, entityMatchingPipelineSimilarityScoreCutOffKey, score)

	return d
}

func resourceEntityMatchingPipelineBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	similarityScoreCutoff := data.Get(entityMatchingPipelineSimilarityScoreCutOffKey).(float64)
	sourceNoteTypes := rawArrayToTypedArray[string](data.Get(entityMatchingPipelineSourceNodeFilterKey))
	targetNodeTypes := rawArrayToTypedArray[string](data.Get(entityMatchingPipelineTargetNodeFilterKey))

	cfg := &configpb.EntityMatchingPipelineConfig{
		NodeFilter: &configpb.EntityMatchingPipelineConfig_NodeFilter{
			SourceNodeTypes: sourceNoteTypes,
			TargetNodeTypes: targetNodeTypes,
		},
		SimilarityScoreCutoff: float32(similarityScoreCutoff),
	}
	builder.WithEntityMatchingPipelineConfig(cfg)
}
