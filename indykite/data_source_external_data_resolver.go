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

func dataSourceExternalDataResolverList() *schema.Resource {
	// Note: headers and request_payload are intentionally omitted from the
	// list view because they commonly carry credentials (e.g. Authorization).
	entrySchema := listEntryCommonSchema(true)
	entrySchema[externalDataResolverURLKey] = computedStringSchema("URL of the external source")
	entrySchema[externalDataResolverMethodKey] = computedStringSchema("HTTP method used to call the external source")
	entrySchema[externalDataResolverRequestTypeKey] = computedStringSchema("Content type of the request")
	entrySchema[externalDataResolverResponseTypeKey] = computedStringSchema("Content type of the response")
	entrySchema[externalDataResolverResponseSelectorKey] = computedStringSchema("Selector applied to the response")

	list := &restListDataSource[ExternalDataResolverResponse]{
		path:            "/external-data-resolvers",
		scopeQueryParam: "project_id",
		scopeKey:        appSpaceIDKey,
		listAttrName:    "external_data_resolvers",
		fullFetch:       true,
		withFilter:      true,
		description:     "List External Data Resolvers in the given Application Space, filtered by exact name match.",
		entrySchema:     entrySchema,
		flatten: func(resolver *ExternalDataResolverResponse) map[string]any {
			return map[string]any{
				"id":                                    resolver.ID,
				customerIDKey:                           resolver.CustomerID,
				appSpaceIDKey:                           resolver.AppSpaceID,
				nameKey:                                 resolver.Name,
				displayNameKey:                          resolver.DisplayName,
				descriptionKey:                          resolver.Description,
				externalDataResolverURLKey:              resolver.URL,
				externalDataResolverMethodKey:           resolver.Method,
				externalDataResolverRequestTypeKey:      resolver.RequestType,
				externalDataResolverResponseTypeKey:     resolver.ResponseType,
				externalDataResolverResponseSelectorKey: resolver.ResponseSelector,
			}
		},
	}
	return list.dataSource()
}
