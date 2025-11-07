// Copyright (c) 2022 IndyKite
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
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCustomer() *schema.Resource {
	// Both name and customer_id are optional for backward compatibility
	// The data source always calls /organizations/current and validates against provided values
	customerIDSchemaOptional := customerIDSchema()
	customerIDSchemaOptional.Required = false
	customerIDSchemaOptional.Optional = true
	customerIDSchemaOptional.Computed = true

	nameSchemaOptional := nameSchema()
	nameSchemaOptional.Required = false
	nameSchemaOptional.Optional = true
	nameSchemaOptional.Computed = true

	return &schema.Resource{
		Description: "It is your entire workspace in the IndyKite platform, " +
			"and will represent your specific company or organization.",
		ReadContext: dataSourceCustomerRead,
		Schema: map[string]*schema.Schema{
			customerIDKey:  customerIDSchemaOptional,
			nameKey:        nameSchemaOptional,
			displayNameKey: setComputed(displayNameSchema()),
			descriptionKey: setComputed(descriptionSchema()),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),
		},
		Timeouts: defaultDataTimeouts(),
	}
}

func dataSourceCustomerRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}

	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp CustomerResponse

	// The only endpoint available is /organizations/current
	// We always call this endpoint regardless of whether customer_id or name is provided
	err := clientCtx.GetClient().Get(ctx, "/organizations/current", &resp)

	if HasFailed(&d, err) {
		return d
	}

	// If customer_id was provided, verify it matches the current organization
	if id, exists := data.GetOk(customerIDKey); exists {
		if resp.ID != id.(string) {
			return append(d, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"customer with ID '%s' not found (current organization ID is '%s')",
					id.(string), resp.ID),
			})
		}
	}

	// If name was provided, verify it matches the current organization
	if name, exists := data.GetOk(nameKey); exists {
		if resp.Name != name.(string) {
			return append(d, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"customer with name '%s' not found (current organization is '%s')",
					name.(string), resp.Name),
			})
		}
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.ID)
	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)

	return d
}
