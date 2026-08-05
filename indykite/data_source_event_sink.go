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

func dataSourceEventSinkList() *schema.Resource {
	// Note: provider/route configuration is intentionally omitted from the
	// list view because providers carry credentials (passwords, access keys).
	list := &restListDataSource[EventSinkResponse]{
		path:            "/event-sinks",
		scopeQueryParam: "project_id",
		scopeKey:        appSpaceIDKey,
		listAttrName:    "event_sinks",
		fullFetch:       false,
		withFilter:      true,
		description:     "List Event Sinks in the given Application Space, filtered by exact name match.",
		entrySchema:     listEntryCommonSchema(true),
		flatten: func(sink *EventSinkResponse) map[string]any {
			return map[string]any{
				"id":           sink.ID,
				customerIDKey:  sink.CustomerID,
				appSpaceIDKey:  sink.AppSpaceID,
				nameKey:        sink.Name,
				displayNameKey: sink.DisplayName,
				descriptionKey: sink.Description,
			}
		},
	}
	return list.dataSource()
}
