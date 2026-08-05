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

func dataSourceKnowledgeQueryList() *schema.Resource {
	entrySchema := listEntryCommonSchema(true)
	entrySchema[knowledgeQueryJSONQueryConfigKey] = computedStringSchema("Query document as JSON string")
	entrySchema[knowledgeQueryStatusKey] = computedStringSchema("Status of the knowledge query")
	entrySchema[knowledgeQueryPolicyID] = computedStringSchema(
		"Identifier of the Authorization Policy the query is bound to")

	list := &restListDataSource[KnowledgeQueryResponse]{
		path:            "/knowledge-queries",
		scopeQueryParam: "project_id",
		scopeKey:        appSpaceIDKey,
		listAttrName:    "knowledge_queries",
		fullFetch:       true,
		withFilter:      true,
		description:     "List Knowledge Queries in the given Application Space, filtered by exact name match.",
		entrySchema:     entrySchema,
		flatten: func(query *KnowledgeQueryResponse) map[string]any {
			return map[string]any{
				"id":                             query.ID,
				customerIDKey:                    query.CustomerID,
				appSpaceIDKey:                    query.AppSpaceID,
				nameKey:                          query.Name,
				displayNameKey:                   query.DisplayName,
				descriptionKey:                   query.Description,
				knowledgeQueryJSONQueryConfigKey: query.Query,
				knowledgeQueryStatusKey:          query.Status,
				knowledgeQueryPolicyID:           query.PolicyID,
			}
		},
	}
	return list.dataSource()
}
