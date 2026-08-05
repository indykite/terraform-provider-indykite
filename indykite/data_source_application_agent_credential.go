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

func dataSourceAppAgentCredentialList() *schema.Resource {
	// Credentials have no name, so there is no exact-name filter;
	// all credentials in the Application Space are returned.
	entrySchema := map[string]*schema.Schema{
		"id":             computedStringSchema("Identifier of the credential"),
		customerIDKey:    computedStringSchema(customerIDDescription),
		appSpaceIDKey:    computedStringSchema(appSpaceIDDescription),
		applicationIDKey: computedStringSchema(applicationIDDescription),
		appAgentIDKey:    computedStringSchema(appAgentIDDescription),
		displayNameKey:   computedStringSchema("The display name for the credential."),
		kidKey:           computedStringSchema("Key ID of the credential"),
		createTimeKey:    createTimeSchema(),
		expireTimeKey:    computedStringSchema("Timestamp when the credential expires"),
	}

	list := &restListDataSource[ApplicationAgentCredentialResponse]{
		path:            "/application-agent-credentials",
		scopeQueryParam: "project_id",
		scopeKey:        appSpaceIDKey,
		listAttrName:    "app_agent_credentials",
		fullFetch:       false,
		withFilter:      false,
		description:     "List Application Agent Credentials in the given Application Space.",
		entrySchema:     entrySchema,
		flatten: func(credential *ApplicationAgentCredentialResponse) map[string]any {
			return map[string]any{
				"id":             credential.ID,
				customerIDKey:    credential.CustomerID,
				appSpaceIDKey:    credential.AppSpaceID,
				applicationIDKey: credential.ApplicationID,
				appAgentIDKey:    credential.ApplicationAgentID,
				displayNameKey:   credential.DisplayName,
				kidKey:           credential.Kid,
				createTimeKey:    formatTimeValue(credential.CreateTime),
				expireTimeKey:    formatTimeValue(credential.ExpireTime),
			}
		},
	}
	return list.dataSource()
}
