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

func dataSourceAuthorizationPolicyList() *schema.Resource {
	entrySchema := listEntryCommonSchema(true)
	entrySchema[authzJSONConfigKey] = computedStringSchema("Policy document as JSON string")
	entrySchema[authzStatusKey] = computedStringSchema("Status of the policy")
	entrySchema[authzTagsKey] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Description: "Tags of the policy",
	}

	list := &restListDataSource[AuthorizationPolicyResponse]{
		path:            "/authorization-policies",
		scopeQueryParam: "project_id",
		scopeKey:        appSpaceIDKey,
		listAttrName:    "authorization_policies",
		fullFetch:       true,
		withFilter:      true,
		description:     "List Authorization Policies in the given Application Space, filtered by exact name match.",
		entrySchema:     entrySchema,
		flatten: func(policy *AuthorizationPolicyResponse) map[string]any {
			return map[string]any{
				"id":               policy.ID,
				customerIDKey:      policy.CustomerID,
				appSpaceIDKey:      policy.AppSpaceID,
				nameKey:            policy.Name,
				displayNameKey:     policy.DisplayName,
				descriptionKey:     policy.Description,
				authzJSONConfigKey: policy.Policy,
				authzStatusKey:     policy.Status,
				authzTagsKey:       policy.Tags,
			}
		},
	}
	return list.dataSource()
}
