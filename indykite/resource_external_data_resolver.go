// Copyright (c) 2024 IndyKite
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
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	externalDataResolverURLKey              = "url"
	externalDataResolverMethodKey           = "method"
	externalDataResolverHeadersKey          = "headers"
	externalDataResolverRequestTypeKey      = "request_type"
	externalDataResolverRequestPayloadKey   = "request_payload"
	externalDataResolverResponseTypeKey     = "response_type"
	externalDataResolverResponseSelectorKey = "response_selector"
)

var (
	externalDataResolverHeaderRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
)

func resourceExternalDataResolver() *schema.Resource {
	return &schema.Resource{
		Description: "ExternalDataResolver is a configuration that allows to fetch data from external sources",

		CreateContext: resExternalDataResolverCreate,
		ReadContext:   resExternalDataResolverRead,
		UpdateContext: resExternalDataResolverUpdate,
		DeleteContext: resExternalDataResolverDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:   locationSchema(),
			customerIDKey: setComputed(customerIDSchema()),
			appSpaceIDKey: setComputed(appSpaceIDSchema()),

			nameKey:        nameSchema(),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),

			externalDataResolverURLKey: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
					validation.IsURLWithHTTPorHTTPS,
				),
				Description: "Full URL to endpoint that will be called",
			},
			externalDataResolverMethodKey: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
					validation.StringInSlice([]string{"GET", "POST", "PUT", "PATCH"}, true),
				),
				Description: "HTTP method to be used for the request. Valid values are: GET, POST, PUT, PATCH.",
			},
			externalDataResolverHeadersKey: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(externalDataResolverHeaderRegex, "invalid key name"),
							Description:  "The name of the header",
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
							Description: "List of values for the header",
						},
					},
				},
				Description: "Headers to be sent with the request, including authorization if needed",
			},
			externalDataResolverRequestTypeKey: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(getMapStringKeys(ExternalDataResolverConfigContentType), true),
				DiffSuppressFunc: func(_, old, newStr string, _ *schema.ResourceData) bool {
					return strings.EqualFold(old, newStr)
				},
				Description: "Request type specify format of request body payload and how to set Content-Type header. Currently only `json` is supported",
			},
			externalDataResolverRequestPayloadKey: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Request payload to be sent to the endpoint. It should be in proper format based on request type",
			},
			externalDataResolverResponseTypeKey: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(getMapStringKeys(ExternalDataResolverConfigContentType), true),
				DiffSuppressFunc: func(_, old, newStr string, _ *schema.ResourceData) bool {
					return strings.EqualFold(old, newStr)
				},
				Description: "Response Type specify expected Content-Type header of response. If mismatch with real response, it will fail. Currently only `json` is supported",
			},
			externalDataResolverResponseSelectorKey: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Selector to extract data from response. Should be in requested format based on Response Type.",
			},
		},
	}
}

func resExternalDataResolverCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	req := CreateExternalDataResolverRequest{
		ProjectID:        data.Get(locationKey).(string),
		Name:             data.Get(nameKey).(string),
		DisplayName:      stringValue(optionalString(data, displayNameKey)),
		Description:      stringValue(optionalString(data, descriptionKey)),
		URL:              data.Get(externalDataResolverURLKey).(string),
		Method:           data.Get(externalDataResolverMethodKey).(string),
		Headers:          buildHeaders(data),
		RequestType:      strings.ToUpper(data.Get(externalDataResolverRequestTypeKey).(string)),
		RequestPayload:   data.Get(externalDataResolverRequestPayloadKey).(string),
		ResponseType:     strings.ToUpper(data.Get(externalDataResolverResponseTypeKey).(string)),
		ResponseSelector: data.Get(externalDataResolverResponseSelectorKey).(string),
	}

	var resp ExternalDataResolverResponse
	err := clientCtx.GetClient().Post(ctx, "/external-data-resolvers", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

	return resExternalDataResolverRead(ctx, data, meta)
}

func resExternalDataResolverRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp ExternalDataResolverResponse
	// Support both ID and name?location=parent_id formats
	path := buildReadPath("/external-data-resolvers", data)
	err := clientCtx.GetClient().Get(ctx, path, &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceID)

	// Set location based on which is present
	if resp.AppSpaceID != "" {
		setData(&d, data, locationKey, resp.AppSpaceID)
	} else if resp.CustomerID != "" {
		setData(&d, data, locationKey, resp.CustomerID)
	}

	setData(&d, data, nameKey, resp.Name)
	// Only set optional fields if they have non-empty values
	if resp.DisplayName != "" {
		setData(&d, data, displayNameKey, resp.DisplayName)
	}
	if resp.Description != "" {
		setData(&d, data, descriptionKey, resp.Description)
	}
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)

	setData(&d, data, externalDataResolverURLKey, resp.URL)
	setData(&d, data, externalDataResolverMethodKey, resp.Method)

	// Convert headers map to list for Terraform schema
	headersList := make([]any, 0, len(resp.Headers))
	for name, value := range resp.Headers {
		// Value could be a string or an array
		var values []string
		switch v := value.(type) {
		case string:
			values = []string{v}
		case []any:
			values = make([]string, len(v))
			for i, val := range v {
				values[i] = val.(string)
			}
		case []string:
			values = v
		}
		headerMap := map[string]any{
			"name":   name,
			"values": values,
		}
		headersList = append(headersList, headerMap)
	}
	setData(&d, data, externalDataResolverHeadersKey, headersList)

	setData(&d, data, externalDataResolverRequestTypeKey, strings.ToLower(resp.RequestType))
	// Only set request_payload if it has a non-empty value
	if resp.RequestPayload != "" {
		setData(&d, data, externalDataResolverRequestPayloadKey, resp.RequestPayload)
	}
	setData(&d, data, externalDataResolverResponseTypeKey, strings.ToLower(resp.ResponseType))
	setData(&d, data, externalDataResolverResponseSelectorKey, resp.ResponseSelector)

	return d
}

func resExternalDataResolverUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := UpdateExternalDataResolverRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	if data.HasChange(externalDataResolverURLKey) {
		url := data.Get(externalDataResolverURLKey).(string)
		req.URL = &url
	}

	if data.HasChange(externalDataResolverMethodKey) {
		method := data.Get(externalDataResolverMethodKey).(string)
		req.Method = &method
	}

	if data.HasChange(externalDataResolverHeadersKey) {
		req.Headers = buildHeaders(data)
	}

	if data.HasChange(externalDataResolverRequestTypeKey) {
		requestType := strings.ToUpper(data.Get(externalDataResolverRequestTypeKey).(string))
		req.RequestType = &requestType
	}

	if data.HasChange(externalDataResolverRequestPayloadKey) {
		requestPayload := data.Get(externalDataResolverRequestPayloadKey).(string)
		req.RequestPayload = &requestPayload
	}

	if data.HasChange(externalDataResolverResponseTypeKey) {
		responseType := strings.ToUpper(data.Get(externalDataResolverResponseTypeKey).(string))
		req.ResponseType = &responseType
	}

	if data.HasChange(externalDataResolverResponseSelectorKey) {
		responseSelector := data.Get(externalDataResolverResponseSelectorKey).(string)
		req.ResponseSelector = &responseSelector
	}

	var resp ExternalDataResolverResponse
	err := clientCtx.GetClient().Put(ctx, "/external-data-resolvers/"+data.Id(), req, &resp)
	if HasFailed(&d, err) {
		return d
	}

	return resExternalDataResolverRead(ctx, data, meta)
}

func resExternalDataResolverDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	err := clientCtx.GetClient().Delete(ctx, "/external-data-resolvers/"+data.Id())
	HasFailed(&d, err)
	return d
}

// buildHeaders converts Terraform schema headers to REST API headers format.
func buildHeaders(data *schema.ResourceData) map[string]any {
	headersSet := data.Get(externalDataResolverHeadersKey).(*schema.Set)
	headers := make(map[string]any, headersSet.Len())
	for _, h := range headersSet.List() {
		headerData := h.(map[string]any)
		name := headerData["name"].(string)
		values := rawArrayToTypedArray[string](headerData["values"])
		// Store as array of values
		headers[name] = values
	}
	return headers
}
