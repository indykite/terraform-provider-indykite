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

func dataSourceTokenIntrospectList() *schema.Resource {
	entrySchema := listEntryCommonSchema(true)
	entrySchema[tokenIntrospectIKGNodeTypeKey] = computedStringSchema("IKG node type the introspected token maps to")
	entrySchema[tokenIntrospectPerformUpsertKey] = &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the token subject is upserted into the IKG",
	}

	list := &restListDataSource[TokenIntrospectResponse]{
		path:            "/token-introspects",
		scopeQueryParam: "project_id",
		scopeKey:        appSpaceIDKey,
		listAttrName:    "token_introspects",
		fullFetch:       true,
		withFilter:      true,
		description: "List Token Introspect configurations in the given Application Space, " +
			"filtered by exact name match.",
		entrySchema: entrySchema,
		flatten: func(ti *TokenIntrospectResponse) map[string]any {
			return map[string]any{
				"id":                            ti.ID,
				customerIDKey:                   ti.CustomerID,
				appSpaceIDKey:                   ti.AppSpaceID,
				nameKey:                         ti.Name,
				displayNameKey:                  ti.DisplayName,
				descriptionKey:                  ti.Description,
				tokenIntrospectIKGNodeTypeKey:   ti.IKGNodeType,
				tokenIntrospectPerformUpsertKey: ti.PerformUpsert,
			}
		},
	}
	return list.dataSource()
}
