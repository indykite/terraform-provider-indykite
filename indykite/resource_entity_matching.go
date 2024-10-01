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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/entitymatching"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	"github.com/indykite/indykite-sdk-go/helpers"
)

const (
	entityMatchingPipelineSourceNodeFilterTypeKey = "source_node_type_filter"
	entityMatchingPipelineTargetNodeFilterTypeKey = "target_node_type_filter"
	entityMatchingPipelineSimilarityScoreTypeKey  = "similarity_score_cutoff"
	entityMatchingResultKey                       = "entity_matching_result"
)

func resourceRunEntityMatchingPipeline() *schema.Resource {
	return &schema.Resource{
		Description: ``,

		CreateContext: createEntityMatchingAndRunFunc,
		ReadContext:   ignoreConfigReadContextFunc,
		UpdateContext: ignoreConfigUpdateContextFunc,
		DeleteContext: ignoreConfigDeleteContextFunc,
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

			entityMatchingPipelineSourceNodeFilterTypeKey: {
				Type:        schema.TypeString,
				Description: "Source node type to be used in the entity matching pipeline.",
				Required:    true,
			},
			entityMatchingPipelineTargetNodeFilterTypeKey: {
				Type:        schema.TypeString,
				Description: "Target node type to be used in the entity matching pipeline.",
				Required:    true,
			},
			entityMatchingPipelineSimilarityScoreTypeKey: {
				Type:         schema.TypeFloat,
				Description:  "Similarity score cutoff to be used in the entity matching pipeline.",
				Required:     true,
				ValidateFunc: validation.FloatBetween(0, 1),
			},
			entityMatchingResultKey: {
				Type:        schema.TypeString,
				Description: "Entity matching result.",
				Computed:    true,
			},
		},
	}
}

func createEntityMatchingAndRunFunc(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var (
		d   diag.Diagnostics
		cem *entitymatching.Client
	)
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}

	location := data.Get(locationKey).(string)
	name := data.Get(nameKey).(string)
	similarityScoreCutoff := data.Get(entityMatchingPipelineSimilarityScoreTypeKey).(float64)
	sourceNoteTypes := data.Get(entityMatchingPipelineSourceNodeFilterTypeKey).(string)
	targetNodeTypes := data.Get(entityMatchingPipelineTargetNodeFilterTypeKey).(string)

	cfg := &configpb.EntityMatchingPipelineConfig{
		NodeFilter: &configpb.EntityMatchingPipelineConfig_NodeFilter{
			SourceNodeTypes: []string{sourceNoteTypes},
			TargetNodeTypes: []string{targetNodeTypes},
		},
		SimilarityScoreCutoff: float32(similarityScoreCutoff),
	}

	var helpersClient helpers.Client
	cem, d = getEntityMatchingClient(ctx)
	if d.HasError() {
		return d
	}
	helpersClient.ClientEntitymatching = cem
	helpersClient.ClientConfig = clientCtx.GetClient()

	res, err := helpersClient.CreateAndRunEntityMatching(location, name, cfg, float32(similarityScoreCutoff))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(location)
	setData(&d, data, entityMatchingResultKey, res.ConfigNodeID)
	return d
}

// ignoreConfigReadContextFunc is a helper function to ignore update context function.
// This is a temporary solution to avoid the update context function for the entity matching pipeline.
// DO NOT USE THIS FUNCTION ANYWHERE ELSE!
func ignoreConfigReadContextFunc(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{}
}

// ignoreConfigUpdateContextFunc is a helper function to ignore update context function.
// This is a temporary solution to avoid the update context function for the entity matching pipeline.
// DO NOT USE THIS FUNCTION ANYWHERE ELSE!
func ignoreConfigUpdateContextFunc(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{}
}

// ignoreConfigDeleteContextFunc is a helper function to ignore delete context function.
// This is a temporary solution to avoid the delete context function for the entity matching pipeline.
// DO NOT USE THIS FUNCTION ANYWHERE ELSE!
func ignoreConfigDeleteContextFunc(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.Diagnostics{}
}
