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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

func dataSourceCustomer() *schema.Resource {
	oneOfIdentifiers := []string{nameKey, customerIDKey}
	return &schema.Resource{
		ReadContext: dataSourceCustomerRead,
		Schema: map[string]*schema.Schema{
			customerIDKey:  setExactlyOneOf(customerIDSchema(), customerIDKey, oneOfIdentifiers),
			nameKey:        setExactlyOneOf(nameSchema(), nameKey, oneOfIdentifiers),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
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

	req := &configpb.ReadCustomerRequest{}
	if name, exists := data.GetOk(nameKey); exists {
		req.Identifier = &configpb.ReadCustomerRequest_Name{
			Name: name.(string),
		}
	} else if id, exists := data.GetOk(customerIDKey); exists {
		req.Identifier = &configpb.ReadCustomerRequest_Id{
			Id: id.(string),
		}
	}

	if err := betterValidationErrorWithPath(req.Validate()); err != nil {
		return append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  err.Error(),
		})
	}

	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadCustomer(ctx, req)
	if HasFailed(&d, err) {
		return d
	}

	if resp.GetCustomer() == nil {
		return append(d, buildPluginError("empty response from server"))
	}
	data.SetId(resp.GetCustomer().GetId())
	setData(&d, data, nameKey, resp.GetCustomer().GetName())
	setData(&d, data, displayNameKey, resp.GetCustomer().GetDisplayName())
	setData(&d, data, descriptionKey, resp.GetCustomer().GetDescription())
	setData(&d, data, createTimeKey, resp.GetCustomer().GetCreateTime())
	setData(&d, data, updateTimeKey, resp.GetCustomer().GetUpdateTime())

	return d
}
