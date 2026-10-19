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
	"maps"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const providerTypeKey = "type"

func dataSourceEventSinkList() *schema.Resource {
	// Provider configurations are not listed because they carry credentials (passwords, access keys);
	// only a summary of each provider is exposed.
	entrySchema := listEntryCommonSchema(true)
	entrySchema[providersKey] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Providers of the event sink, sorted by name, without their credentials.",
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			providerNameKey: computedStringSchema("Name of the provider, referenced by the routes."),
			providerTypeKey: computedStringSchema(
				"Type of the provider: `kafka`, `azure_event_grid`, `azure_service_bus` or `pubsub`."),
			providerDisplayKey: computedStringSchema("Display name of the provider."),
			lastErrorKey:       computedStringSchema("Last delivery error of the provider, empty when none."),
		}},
	}

	list := &restListDataSource[EventSinkResponse]{
		path:            "/event-sinks",
		scopeQueryParam: "project_id",
		scopeKey:        appSpaceIDKey,
		listAttrName:    "event_sinks",
		fullFetch:       false,
		withFilter:      true,
		description:     "List Event Sinks in the given Application Space, filtered by exact name match.",
		entrySchema:     entrySchema,
		flatten: func(sink *EventSinkResponse) map[string]any {
			providers, _ := eventSinkConfigFromResponse(sink)["providers"].(map[string]any)
			return map[string]any{
				"id":           sink.ID,
				customerIDKey:  sink.CustomerID,
				appSpaceIDKey:  sink.AppSpaceID,
				nameKey:        sink.Name,
				displayNameKey: sink.DisplayName,
				descriptionKey: sink.Description,
				providersKey:   summarizeEventSinkProviders(providers),
			}
		},
	}
	return list.dataSource()
}

// summarizeEventSinkProviders returns the name, type, display name and last error of each provider,
// sorted by name. Credentials are never copied.
func summarizeEventSinkProviders(providers map[string]any) []map[string]any {
	providerTypes := map[string]string{
		"kafka": kafkaKey, "azure_event_grid": azureEventGridKey, "azureEventGrid": azureEventGridKey,
		"azure_service_bus": azureServiceBusKey, "azureServiceBus": azureServiceBusKey,
		"pubsub": pubsubKey, "pubSub": pubsubKey,
	}
	summaries := make([]map[string]any, 0, len(providers))
	for _, name := range slices.Sorted(maps.Keys(providers)) {
		provider, _ := providers[name].(map[string]any)
		for apiType, providerType := range providerTypes {
			cfg, ok := provider[apiType].(map[string]any)
			if !ok {
				continue
			}
			summaries = append(summaries, map[string]any{
				providerNameKey:    name,
				providerTypeKey:    providerType,
				providerDisplayKey: getEither(cfg, "display_name", "displayName"),
				lastErrorKey:       getEither(cfg, "last_error", "lastError"),
			})
			break
		}
	}
	return summaries
}
