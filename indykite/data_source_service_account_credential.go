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

func dataSourceServiceAccountCredentialList() *schema.Resource {
	// Credentials have no name, so there is no exact-name filter;
	// all credentials in the Customer are returned.
	entrySchema := map[string]*schema.Schema{
		"id":                computedStringSchema("Identifier of the credential"),
		customerIDKey:       computedStringSchema(customerIDDescription),
		serviceAccountIDKey: computedStringSchema("Identifier of the Service Account"),
		displayNameKey:      computedStringSchema("The display name for the credential."),
		kidKey:              computedStringSchema("Key ID of the credential"),
		createTimeKey:       createTimeSchema(),
		expireTimeKey:       computedStringSchema("Timestamp when the credential expires"),
	}

	list := &restListDataSource[ServiceAccountCredentialResponse]{
		path:            "/service-account-credentials",
		scopeQueryParam: "organization_id",
		scopeKey:        customerIDKey,
		listAttrName:    "service_account_credentials",
		fullFetch:       false,
		withFilter:      false,
		description:     "List Service Account Credentials in the given Customer.",
		entrySchema:     entrySchema,
		flatten: func(credential *ServiceAccountCredentialResponse) map[string]any {
			return map[string]any{
				"id":                credential.ID,
				customerIDKey:       credential.OrganizationID,
				serviceAccountIDKey: credential.ServiceAccountID,
				displayNameKey:      credential.DisplayName,
				kidKey:              credential.Kid,
				createTimeKey:       formatTimeValue(credential.CreateTime),
				expireTimeKey:       credential.ExpireTime,
			}
		},
	}
	return list.dataSource()
}
