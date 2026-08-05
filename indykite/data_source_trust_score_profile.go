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

func dataSourceTrustScoreProfileList() *schema.Resource {
	entrySchema := listEntryCommonSchema(true)
	entrySchema[trustScoreProfileNodeClassification] = computedStringSchema(
		"Node classification the profile applies to")
	entrySchema[trustScoreProfileSchedule] = computedStringSchema("Update frequency of the trust score")

	list := &restListDataSource[TrustScoreProfileResponse]{
		path:            "/trust-score-profiles",
		scopeQueryParam: "project_id",
		scopeKey:        appSpaceIDKey,
		listAttrName:    "trust_score_profiles",
		fullFetch:       true,
		withFilter:      true,
		description:     "List Trust Score Profiles in the given Application Space, filtered by exact name match.",
		entrySchema:     entrySchema,
		flatten: func(profile *TrustScoreProfileResponse) map[string]any {
			return map[string]any{
				"id":                                profile.ID,
				customerIDKey:                       profile.CustomerID,
				appSpaceIDKey:                       profile.AppSpaceID,
				nameKey:                             profile.Name,
				displayNameKey:                      profile.DisplayName,
				descriptionKey:                      profile.Description,
				trustScoreProfileNodeClassification: profile.NodeClassification,
				trustScoreProfileSchedule:           profile.Schedule,
			}
		},
	}
	return list.dataSource()
}
