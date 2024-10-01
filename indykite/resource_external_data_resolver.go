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
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
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
	readContext := configReadContextFunc(resourceExternalDataResolverFlatten)

	return &schema.Resource{
		Description: "ExternalDataResolver is a configuration that allows to fetch data from external sources",

		CreateContext: configCreateContextFunc(resourceExternalDataResolverBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceExternalDataResolverBuild, readContext),
		DeleteContext: configDeleteContextFunc(),
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
				Description: "HTTP method to be used for the request",
			},
			externalDataResolverHeadersKey: {
				Type:     schema.TypeSet,
				Required: true,
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
				Description: "Request type specify format of request body payload and how to set Content-Type header",
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
				Description: "Response Type specify expected Content-Type header of response. If mismatch with real response, it will fail",
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

func resourceExternalDataResolverFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) diag.Diagnostics {
	var d diag.Diagnostics
	resolver := resp.GetConfigNode().GetExternalDataResolverConfig()
	url := resolver.GetUrl()
	setData(&d, data, externalDataResolverURLKey, url)

	method := resolver.GetMethod()
	setData(&d, data, externalDataResolverMethodKey, method)

	headerNames := make([]string, 0, len(resolver.GetHeaders()))
	for name := range resolver.GetHeaders() {
		headerNames = append(headerNames, name)
	}
	sort.Strings(headerNames)

	headersList := make([]any, 0, len(headerNames))
	for _, name := range headerNames {
		header := resolver.GetHeaders()[name]
		headerMap := map[string]any{
			"name":   name,
			"values": header.GetValues(),
		}
		headersList = append(headersList, headerMap)
	}
	setData(&d, data, externalDataResolverHeadersKey, headersList)

	requestType := resolver.GetRequestType()
	setData(&d, data, externalDataResolverRequestTypeKey, externalDataResolverContentTypeToString[requestType])

	requestPayload := resolver.GetRequestPayload()
	setData(&d, data, externalDataResolverRequestPayloadKey, string(requestPayload))

	responseType := resolver.GetResponseType()
	setData(&d, data, externalDataResolverResponseTypeKey, externalDataResolverContentTypeToString[responseType])

	responseSelector := resolver.GetResponseSelector()
	setData(&d, data, externalDataResolverResponseSelectorKey, responseSelector)

	return d
}

func resourceExternalDataResolverBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	requestPayloadStr := data.Get(externalDataResolverRequestPayloadKey).(string)
	cfg := &configpb.ExternalDataResolverConfig{
		Url:     data.Get(externalDataResolverURLKey).(string),
		Method:  data.Get(externalDataResolverMethodKey).(string),
		Headers: getHeaders(data),
		RequestType: ExternalDataResolverConfigContentType[strings.ToLower(data.Get(
			externalDataResolverRequestTypeKey).(string))],
		RequestPayload: []byte(requestPayloadStr),
		ResponseType: ExternalDataResolverConfigContentType[strings.ToLower(data.Get(
			externalDataResolverResponseTypeKey).(string))],
		ResponseSelector: data.Get(externalDataResolverResponseSelectorKey).(string),
	}
	builder.WithExternalDataResolverConfig(cfg)
}

func getHeaders(data *schema.ResourceData) map[string]*configpb.ExternalDataResolverConfig_Header {
	headersSet := data.Get(externalDataResolverHeadersKey).(*schema.Set)
	headers := make(map[string]*configpb.ExternalDataResolverConfig_Header, headersSet.Len())
	for _, h := range headersSet.List() {
		headerData := h.(map[string]any)
		name := headerData["name"].(string)
		values := rawArrayToTypedArray[string](headerData["values"])
		headers[name] = &configpb.ExternalDataResolverConfig_Header{Values: values}
	}
	return headers
}
