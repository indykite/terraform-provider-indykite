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

func dataSourceAuditSigningList() *schema.Resource {
	entrySchema := listEntryCommonSchema(true)
	entrySchema[auditSigningProviderKey] = computedStringSchema(
		"Provider managing the signing key: PLATFORM_MANAGED, CUSTOMER_GCP_KMS, CUSTOMER_AWS_KMS " +
			"or CUSTOMER_AZURE_KEY_VAULT")
	entrySchema[auditSigningKeyResourceKey] = computedStringSchema(
		"Resource identifier of the customer managed signing key")
	entrySchema[auditSigningKidKey] = computedStringSchema("Key ID (kid) published with signed audit records")
	entrySchema[auditSigningAuthParamsKey] = &schema.Schema{
		Type:     schema.TypeMap,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Description: "Authentication parameter names configured for the customer managed key. " +
			"Values are always masked by the API and returned empty.",
	}

	list := &restListDataSource[AuditSigningResponse]{
		path:            "/audit-signings",
		scopeQueryParam: "project_id",
		scopeKey:        appSpaceIDKey,
		listAttrName:    "audit_signings",
		fullFetch:       true,
		withFilter:      true,
		description: "List Audit Signing configurations in the given Application Space, " +
			"filtered by exact name match.",
		entrySchema: entrySchema,
		flatten: func(cfg *AuditSigningResponse) map[string]any {
			authParams := make(map[string]any, len(cfg.AuthParams))
			for k, v := range cfg.AuthParams {
				authParams[k] = v
			}
			return map[string]any{
				"id":                       cfg.ID,
				customerIDKey:              cfg.CustomerID,
				appSpaceIDKey:              cfg.AppSpaceID,
				nameKey:                    cfg.Name,
				displayNameKey:             cfg.DisplayName,
				descriptionKey:             cfg.Description,
				auditSigningProviderKey:    cfg.Provider,
				auditSigningKeyResourceKey: stringValue(cfg.KeyResource),
				auditSigningKidKey:         stringValue(cfg.Kid),
				auditSigningAuthParamsKey:  authParams,
			}
		},
	}
	return list.dataSource()
}
