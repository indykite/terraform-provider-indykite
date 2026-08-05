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
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// restListDataSource describes a data source listing one Config API collection.
// All list endpoints share the same shape: GET {path}?{scopeQueryParam}={scope GID}
// returning {"data": [...]}; entries are optionally filtered client-side by exact name match.
type restListDataSource[T any] struct {
	entrySchema map[string]*schema.Schema
	// flatten converts one API response entry into a listAttrName element;
	// its nameKey value is matched against the filter.
	flatten func(item *T) map[string]any
	// path is the REST collection path, e.g. "/authorization-policies".
	path string
	// scopeQueryParam is the query parameter carrying the parent scope: "project_id" or "organization_id".
	scopeQueryParam string
	// scopeKey is the Terraform attribute holding the parent scope GID: appSpaceIDKey or customerIDKey.
	scopeKey string
	// listAttrName is the computed attribute holding the resulting entries,
	// also used as the middle segment of the synthesized Terraform ID.
	listAttrName string
	description  string
	// fullFetch requests full configuration instead of metadata only;
	// required when entrySchema exposes fields stored in the configuration.
	fullFetch bool
	// withFilter enables the required exact-name filter;
	// disabled for credentials, which have no name to match on.
	withFilter bool
}

func (l *restListDataSource[T]) dataSource() *schema.Resource {
	scopeSchema := appSpaceIDSchema()
	if l.scopeKey == customerIDKey {
		scopeSchema = customerIDSchema()
	}
	dataSchema := map[string]*schema.Schema{
		l.scopeKey: scopeSchema,
		l.listAttrName: {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Resource{Schema: l.entrySchema},
		},
	}
	if l.withFilter {
		dataSchema[filterKey] = exactNameFilterSchema()
	}
	return &schema.Resource{
		Description: l.description,
		ReadContext: l.readContext,
		Schema:      dataSchema,
		Timeouts:    defaultDataTimeouts(),
	}
}

func (l *restListDataSource[T]) readContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	var match []string
	if l.withFilter {
		rawFilter := data.Get(filterKey).([]any)
		match = make([]string, len(rawFilter))
		for i, v := range rawFilter {
			match[i] = v.(string)
		}
	}

	clientCtx := getClientContext(&d, meta)
	if d.HasError() {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	scopeID := data.Get(l.scopeKey).(string)
	path := l.path + "?" + l.scopeQueryParam + "=" + scopeID
	if l.fullFetch {
		path += "&full_fetch=true"
	}
	var resp ListResponse[T]
	err := clientCtx.GetClient().Get(ctx, path, &resp)
	if HasFailed(&d, err) {
		return d
	}

	entries := make([]map[string]any, 0, len(resp.Data))
	for i := range resp.Data {
		entry := l.flatten(&resp.Data[i])
		if l.withFilter {
			// Apply exact name match filter (MinItems: 1 ensures filter is always present)
			name, _ := entry[nameKey].(string)
			matchFound := false
			for _, filter := range match {
				if name == filter {
					matchFound = true
					break
				}
			}
			if !matchFound {
				continue
			}
		}
		entries = append(entries, entry)
	}
	setData(&d, data, l.listAttrName, entries)

	id := scopeID + "/" + l.listAttrName
	if l.withFilter {
		id += "/" + strings.Join(match, ",")
	}
	data.SetId(id)
	return d
}

func computedStringSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: description,
	}
}

// listEntryCommonSchema returns the computed attributes shared by all named
// config list entries; project-scoped collections also expose app_space_id.
func listEntryCommonSchema(projectScoped bool) map[string]*schema.Schema {
	entry := map[string]*schema.Schema{
		"id":           computedStringSchema("Identifier of the resource"),
		customerIDKey:  computedStringSchema(customerIDDescription),
		nameKey:        computedStringSchema(nameDescription),
		displayNameKey: computedStringSchema("The display name for the instance."),
		descriptionKey: computedStringSchema("Description of the resource."),
	}
	if projectScoped {
		entry[appSpaceIDKey] = computedStringSchema(appSpaceIDDescription)
	}
	return entry
}

// formatTimeValue renders server timestamps for list entries the same way
// setData does for top-level attributes.
func formatTimeValue(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}
