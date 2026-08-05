// Copyright (c) 2026 IndyKite
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceEntityMatchingPipelineList() *schema.Resource {
	entrySchema := listEntryCommonSchema(true)
	entrySchema[entityMatchingPipelineSimilarityScoreCutOffKey] = &schema.Schema{
		Type:        schema.TypeFloat,
		Computed:    true,
		Description: "Similarity score cutoff of the pipeline",
	}
	entrySchema[entityMatchingPipelineRerunInterval] = computedStringSchema("Rerun interval of the pipeline")

	list := &restListDataSource[EntityMatchingPipelineResponse]{
		path:            "/entity-matching-pipelines",
		scopeQueryParam: "project_id",
		scopeKey:        appSpaceIDKey,
		listAttrName:    "entity_matching_pipelines",
		fullFetch:       true,
		withFilter:      true,
		description: "List Entity Matching Pipelines in the given Application Space, " +
			"filtered by exact name match.",
		entrySchema: entrySchema,
		flatten: func(pipeline *EntityMatchingPipelineResponse) map[string]any {
			return map[string]any{
				"id":           pipeline.ID,
				customerIDKey:  pipeline.CustomerID,
				appSpaceIDKey:  pipeline.AppSpaceID,
				nameKey:        pipeline.Name,
				displayNameKey: pipeline.DisplayName,
				descriptionKey: pipeline.Description,
				entityMatchingPipelineSimilarityScoreCutOffKey: float64(pipeline.SimilarityScoreCutoff),
				entityMatchingPipelineRerunInterval:            pipeline.RerunInterval,
			}
		},
	}
	return list.dataSource()
}
