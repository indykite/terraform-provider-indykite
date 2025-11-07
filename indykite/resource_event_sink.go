// Copyright (c) 2025 IndyKite
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
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	providersKey        = "providers"
	routesKey           = "routes"
	providerNameKey     = "provider_name"
	providerIDKey       = "provider_id"
	stopProcessingKey   = "stop_processing"
	keysValuesKey       = "keys_values_filter"
	keyValuePairsKey    = "key_value_pairs"
	kafkaKey            = "kafka"
	azureEventGridKey   = "azure_event_grid"
	azureServiceBusKey  = "azure_service_bus"
	brokersKey          = "brokers"
	topicKey            = "topic"
	keyKey              = "key"
	valueKey            = "value"
	evTypeKey           = "event_type"
	disableTLSKey       = "disable_tls"
	tlsSkipVerifyKey    = "tls_skip_verify"
	usernameKey         = "username"
	passwordKey         = "password"
	topicEndpointKey    = "topic_endpoint"
	accessKey           = "access_key"
	connectionStringKey = "connection_string"
	queueKey            = "queue_or_topic_name"
	providerDisplayKey  = "provider_display_name"
	routeDisplayKey     = "route_display_name"
	routeIDKey          = "route_id"
	lastErrorKey        = "last_error"
	supportedFilters    = `
## Supported filters

| **Method** | **Event Type** | **Key** | **Value (example)** |
| --- | --- | --- | --- |
|  | **Ingest Events** |  |  |
| **BatchUpsertNodes** | indykite.audit.capture.upsert.node | captureLabel | Car |
|  |  | captureLabel | Green |
| **BatchUpsertRelationships** | indykite.audit.capture.upsert.relationship | captureLabel | RENT |
| **BatchDeleteNodes** | indykite.audit.capture.delete.node | captureLabel | Car |
|  |  | captureLabel | Green |
| **BatchDeleteRelationships** | indykite.audit.capture.delete.relationship | captureLabel | RENT |
| **BatchDeleteNodeProperties** | indykite.audit.capture.delete.node.property |  |  |
| **BatchDeleteRelationshipProperties** | indykite.audit.capture.delete.relationship.property |  |  |
| **BatchDeleteNodeTags** | indykite.audit.capture.delete.node.tag | captureLabel | Car |
|  |  | captureLabel | Green |
|  | **Configuration Events** |  |  |
| Config | indykite.audit.config.create |  |  |
|  | indykite.audit.config.read |  |  |
|  | indykite.audit.config.update |  |  |
|  | indykite.audit.config.delete |  |  |
|  | indykite.audit.config.permission.assign |  |  |
|  | indykite.audit.config.permission.revoke |  |  |
|  | **Token Events** |  |  |
| TokenIntrospect | indykite.audit.credentials.token.introspected |  |  |
|  | **Authorization Events** |  |  |
| Authorization | indykite.audit.authorization.isauthorized |  |  |
|  | indykite.audit.authorization.whatauthorized |  |  |
|  | indykite.audit.authorization.whoauthorized |  |  |
|  | **Ciq Events** |  |  |
| Ciq | indykite.audit.ciq.execute |  |  |
|  |  |  |  | `
)

func resourceEventSink() *schema.Resource {
	providerOneOf := []string{kafkaKey, azureEventGridKey, azureServiceBusKey}

	return &schema.Resource{
		Description: `
		Event Sink configuration is used to configure outbound events.

		There can be only one configuration per AppSpace (Project).

		Outbound events are designed to notify external systems about important changes within
		the IndyKite Knowledge Graph (IKG).

		These external systems may require real-time synchronization or need to react to
		changes occurring in the platform.

		` + supportedFilters,

		CreateContext: resEventSinkCreate,
		ReadContext:   resEventSinkRead,
		UpdateContext: resEventSinkUpdate,
		DeleteContext: resEventSinkDelete,
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
			providersKey: {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Resource{Schema: providerSchema()},
			},
			routesKey: {
				Type:     schema.TypeList,
				Elem:     &schema.Resource{Schema: routeSchema()},
				Required: true,
			},
		},
		CustomizeDiff: validateProviderOneOf(providerOneOf),
	}
}

func providerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		providerNameKey: {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(2, 63),
				validation.StringMatch(regexp.MustCompile(`^[a-z](?:[-a-z0-9]{0,61}[a-z0-9])$`),
					"must start with a lowercase letter, followed by 0-62 characters "+
						"(lowercase letters, digits, and hyphens in the middle)."),
			),
		},
		kafkaKey: {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "KafkaSinkConfig",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					brokersKey: {
						Type:        schema.TypeList,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Required:    true,
						Description: "Brokers specify Kafka destinations to connect to.",
					},
					topicKey: {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.All(
							validation.StringLenBetween(1, 249),
							validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9._-]+$`),
								"must contain only letters, numbers, underscores or hyphens"),
						),
					},
					disableTLSKey: {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Disable TLS for communication. Highly NOT RECOMMENDED.",
					},
					tlsSkipVerifyKey: {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Skip TLS certificate verification. NOT RECOMMENDED.",
					},
					usernameKey: {
						Type:     schema.TypeString,
						Required: true,
					},
					passwordKey: {
						Type:      schema.TypeString,
						Required:  true,
						Sensitive: true,
					},
					providerDisplayKey: {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.All(
							validation.StringLenBetween(2, 254),
						),
					},
					lastErrorKey: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Last error message from the Kafka sink",
					},
				},
			},
		},
		azureEventGridKey: {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "AzureEventGridSinkConfig",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					topicEndpointKey: {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.All(
							validation.StringLenBetween(1, 1024),
						),
					},
					accessKey: {
						Type:      schema.TypeString,
						Required:  true,
						Sensitive: true,
						ValidateFunc: validation.All(
							validation.StringLenBetween(1, 1024),
						),
					},
					providerDisplayKey: {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.All(
							validation.StringLenBetween(2, 254),
						),
					},
					lastErrorKey: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Last error message from the Azure Event Grid sink",
					},
				},
			},
		},
		azureServiceBusKey: {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "AzureServiceBusSinkConfig",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					connectionStringKey: {
						Type:      schema.TypeString,
						Required:  true,
						Sensitive: true,
						ValidateFunc: validation.All(
							validation.StringLenBetween(1, 1024),
						),
					},
					queueKey: {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.All(
							validation.StringLenBetween(1, 1024),
						),
					},
					providerDisplayKey: {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.All(
							validation.StringLenBetween(2, 254),
						),
					},
					lastErrorKey: {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Last error message from the Azure Service Bus sink",
					},
				},
			},
		},
	}
}

func routeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		providerIDKey: {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(2, 63),
				validation.StringMatch(regexp.MustCompile(`^[a-z](?:[-a-z0-9]{0,61}[a-z0-9])$`),
					"must contain only lowercase letters, numbers, or hyphens"),
			),
		},
		stopProcessingKey: {
			Type:     schema.TypeBool,
			Optional: true,
		},
		keysValuesKey: {
			Type:     schema.TypeList,
			Elem:     &schema.Resource{Schema: keysValuesSchema()},
			Optional: true,
			MaxItems: 1,
		},
		routeDisplayKey: {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(2, 254),
			),
		},
		routeIDKey: {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(2, 63),
			),
		},
	}
}

func keysValuesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		keyValuePairsKey: {
			Type:        schema.TypeList,
			Description: "List of key/value pairs for the ingest event types. ",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					keyKey: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Key for the ingest eventType",
					},
					valueKey: {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Value for the ingest eventType",
					},
				},
			},
		},
		evTypeKey: {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.All(
				validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_*\\.]+$`),
					"must contain only letters, numbers, underscores, asterisks and dots"),
			),
		},
	}
}

func resEventSinkCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	providers := data.Get(providersKey).([]any)
	routes := data.Get(routesKey).([]any)

	req := CreateEventSinkRequest{
		ProjectID:   data.Get(locationKey).(string),
		Name:        data.Get(nameKey).(string),
		DisplayName: stringValue(optionalString(data, displayNameKey)),
		Description: stringValue(optionalString(data, descriptionKey)),
		Providers:   buildProvidersMap(providers),
		Routes:      buildRoutesList(routes),
	}

	var resp EventSinkResponse
	err := clientCtx.GetClient().Post(ctx, "/event-sinks", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

	return resEventSinkRead(ctx, data, meta)
}

func resEventSinkRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp EventSinkResponse
	// Support both ID and name?location=parent_id formats
	path := buildReadPath("/event-sinks", data)
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
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)

	// Flatten providers and routes from Config
	flattenEventSinkConfig(&d, data, resp.Config)

	return d
}

func resEventSinkUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := UpdateEventSinkRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	if data.HasChange(providersKey) || data.HasChange(routesKey) {
		req.Config = buildEventSinkConfig(data)
	}

	var resp EventSinkResponse
	err := clientCtx.GetClient().Put(ctx, "/event-sinks/"+data.Id(), req, &resp)
	if HasFailed(&d, err) {
		return d
	}

	return resEventSinkRead(ctx, data, meta)
}

func resEventSinkDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	err := clientCtx.GetClient().Delete(ctx, "/event-sinks/"+data.Id())
	HasFailed(&d, err)
	return d
}

// buildEventSinkConfig converts Terraform schema data to Config map.
func buildEventSinkConfig(data *schema.ResourceData) map[string]any {
	providers := data.Get(providersKey).([]any)
	routes := data.Get(routesKey).([]any)

	return map[string]any{
		"providers": buildProvidersMap(providers),
		"routes":    buildRoutesList(routes),
	}
}

// buildProvidersMap builds the providers map from Terraform schema data.
func buildProvidersMap(providers []any) map[string]any {
	providersMap := make(map[string]any, len(providers))
	for _, provider := range providers {
		item, ok := provider.(map[string]any)
		if !ok {
			continue
		}
		key := item[providerNameKey].(string)
		providerConfig := buildProviderConfig(item)
		if providerConfig != nil {
			providersMap[key] = providerConfig
		}
	}
	return providersMap
}

// buildProviderConfig builds a single provider configuration.
func buildProviderConfig(item map[string]any) map[string]any {
	if kafkaList, ok := item[kafkaKey].([]any); ok && len(kafkaList) > 0 {
		return buildKafkaProviderMap(kafkaList[0].(map[string]any))
	}
	if gridList, ok := item[azureEventGridKey].([]any); ok && len(gridList) > 0 {
		return buildAzureEventGridMap(gridList[0].(map[string]any))
	}
	if busList, ok := item[azureServiceBusKey].([]any); ok && len(busList) > 0 {
		return buildAzureServiceBusMap(busList[0].(map[string]any))
	}
	return nil
}

// buildKafkaProviderMap builds Kafka provider configuration map.
func buildKafkaProviderMap(kafkaData map[string]any) map[string]any {
	return map[string]any{
		"kafka": map[string]any{
			"brokers":         kafkaData[brokersKey],
			"topic":           kafkaData[topicKey],
			"username":        kafkaData[usernameKey],
			"password":        kafkaData[passwordKey],
			"disable_tls":     kafkaData[disableTLSKey],
			"tls_skip_verify": kafkaData[tlsSkipVerifyKey],
			"display_name":    kafkaData[providerDisplayKey],
		},
	}
}

// buildAzureEventGridMap builds Azure Event Grid configuration map.
func buildAzureEventGridMap(gridData map[string]any) map[string]any {
	return map[string]any{
		"azure_event_grid": map[string]any{
			"topic_endpoint": gridData[topicEndpointKey],
			"access_key":     gridData[accessKey],
			"display_name":   gridData[providerDisplayKey],
		},
	}
}

// buildAzureServiceBusMap builds Azure Service Bus configuration map.
func buildAzureServiceBusMap(busData map[string]any) map[string]any {
	return map[string]any{
		"azure_service_bus": map[string]any{
			"connection_string":   busData[connectionStringKey],
			"queue_or_topic_name": busData[queueKey],
			"display_name":        busData[providerDisplayKey],
		},
	}
}

// buildRoutesList builds the routes list from Terraform schema data.
func buildRoutesList(routes []any) []any {
	routesList := make([]any, len(routes))
	for i, route := range routes {
		item, ok := route.(map[string]any)
		if !ok {
			continue
		}
		routesList[i] = buildRouteMap(item)
	}
	return routesList
}

// buildRouteMap builds a single route map.
func buildRouteMap(item map[string]any) map[string]any {
	routeMap := map[string]any{
		"provider_id":     item[providerIDKey],
		"stop_processing": item[stopProcessingKey],
		"display_name":    item[routeDisplayKey],
		"id":              item[routeIDKey],
	}

	if kvList, ok := item[keysValuesKey].([]any); ok && len(kvList) > 0 {
		routeMap["event_type_key_values_filter"] = buildKeysValuesMap(kvList[0].(map[string]any))
	}

	return routeMap
}

// buildKeysValuesMap builds the keysValues map for a route.
func buildKeysValuesMap(kvData map[string]any) map[string]any {
	var pairs []any
	if pairsList, ok := kvData[keyValuePairsKey].([]any); ok {
		pairs = make([]any, 0, len(pairsList))
		for _, pair := range pairsList {
			if pairData, ok := pair.(map[string]any); ok {
				pairs = append(pairs, map[string]any{
					"key":   pairData[keyKey],
					"value": pairData[valueKey],
				})
			}
		}
	}
	return map[string]any{
		"key_value_pairs": pairs,
		"event_type":      kvData[evTypeKey],
	}
}

// buildKafkaConfig builds Kafka configuration preserving sensitive password from state.
func buildKafkaConfig(kafkaData map[string]any, data *schema.ResourceData) map[string]any {
	// Preserve sensitive password from state
	var oldPassword any
	if val, ok := data.Get(providersKey).([]any); ok && len(val) > 0 {
		for _, y := range val {
			if kafkaList, ok := y.(map[string]any)[kafkaKey].([]any); ok && len(kafkaList) > 0 {
				kafka := kafkaList[0].(map[string]any)
				oldPassword = kafka[passwordKey]
			}
		}
	}

	// Helper to get value from either snake_case or camelCase
	getValue := func(snakeCase, camelCase string) any {
		if val, ok := kafkaData[snakeCase]; ok {
			return val
		}
		return kafkaData[camelCase]
	}

	return map[string]any{
		brokersKey:         getValue("brokers", "brokers"),
		topicKey:           getValue("topic", "topic"),
		usernameKey:        getValue("username", "username"),
		passwordKey:        oldPassword,
		disableTLSKey:      getValue("disable_tls", "disableTls"),
		tlsSkipVerifyKey:   getValue("tls_skip_verify", "tlsSkipVerify"),
		providerDisplayKey: getValue("display_name", "displayName"),
	}
}

// flattenEventSinkConfig converts Config map to Terraform schema data.
func buildAzureEventGridConfig(gridData map[string]any, data *schema.ResourceData) map[string]any {
	// Preserve sensitive access key from state
	var oldAccessKey any
	if val, ok := data.Get(providersKey).([]any); ok && len(val) > 0 {
		for _, y := range val {
			if gridList, ok := y.(map[string]any)[azureEventGridKey].([]any); ok && len(gridList) > 0 {
				grid := gridList[0].(map[string]any)
				oldAccessKey = grid[accessKey]
			}
		}
	}

	// Helper to get value from either snake_case or camelCase
	getValue := func(snakeCase, camelCase string) any {
		if val, ok := gridData[snakeCase]; ok {
			return val
		}
		return gridData[camelCase]
	}

	return map[string]any{
		topicEndpointKey:   getValue("topic_endpoint", "topicEndpoint"),
		accessKey:          oldAccessKey,
		providerDisplayKey: getValue("display_name", "displayName"),
	}
}

func buildAzureServiceBusConfig(busData map[string]any, data *schema.ResourceData) map[string]any {
	// Preserve sensitive connection string from state
	var oldConnection any
	if val, ok := data.Get(providersKey).([]any); ok && len(val) > 0 {
		for _, y := range val {
			if busList, ok := y.(map[string]any)[azureServiceBusKey].([]any); ok && len(busList) > 0 {
				bus := busList[0].(map[string]any)
				oldConnection = bus[connectionStringKey]
			}
		}
	}

	// Helper to get value from either snake_case or camelCase
	getValue := func(snakeCase, camelCase string) any {
		if val, ok := busData[snakeCase]; ok {
			return val
		}
		return busData[camelCase]
	}

	return map[string]any{
		connectionStringKey: oldConnection,
		queueKey:            getValue("queue_or_topic_name", "queueOrTopicName"),
		providerDisplayKey:  getValue("display_name", "displayName"),
	}
}

func flattenEventSinkConfig(d *diag.Diagnostics, data *schema.ResourceData, config map[string]any) {
	providersMap, _ := config["providers"].(map[string]any)
	var results []map[string]any

	for key, p := range providersMap {
		providerData, _ := p.(map[string]any)

		// Try both snake_case and camelCase for provider types
		if kafkaData, ok := providerData["kafka"].(map[string]any); ok {
			kafkaConfig := buildKafkaConfig(kafkaData, data)
			results = append(results, map[string]any{
				providerNameKey: key,
				kafkaKey:        []any{kafkaConfig},
			})
		} else if gridData, ok := providerData["azure_event_grid"].(map[string]any); !ok {
			if gridData, ok = providerData["azureEventGrid"].(map[string]any); ok {
				gridConfig := buildAzureEventGridConfig(gridData, data)
				results = append(results, map[string]any{
					providerNameKey:   key,
					azureEventGridKey: []any{gridConfig},
				})
			}
		} else {
			gridConfig := buildAzureEventGridConfig(gridData, data)
			results = append(results, map[string]any{
				providerNameKey:   key,
				azureEventGridKey: []any{gridConfig},
			})
		}
		if busData, ok := providerData["azure_service_bus"].(map[string]any); !ok {
			if busData, ok = providerData["azureServiceBus"].(map[string]any); ok {
				busConfig := buildAzureServiceBusConfig(busData, data)
				results = append(results, map[string]any{
					providerNameKey:    key,
					azureServiceBusKey: []any{busConfig},
				})
			}
		} else {
			busConfig := buildAzureServiceBusConfig(busData, data)
			results = append(results, map[string]any{
				providerNameKey:    key,
				azureServiceBusKey: []any{busConfig},
			})
		}
	}
	setData(d, data, providersKey, results)

	routesList, _ := config["routes"].([]any)
	routes := make([]any, len(routesList))
	for i, r := range routesList {
		routeData, _ := r.(map[string]any)

		// Helper to get value from either snake_case or camelCase
		getValue := func(snakeCase, camelCase string) any {
			if val, ok := routeData[snakeCase]; ok {
				return val
			}
			return routeData[camelCase]
		}

		routeMap := map[string]any{
			providerIDKey:     getValue("provider_id", "providerId"),
			stopProcessingKey: getValue("stop_processing", "stopProcessing"),
			routeDisplayKey:   getValue("display_name", "displayName"),
			routeIDKey:        routeData["id"],
		}

		// Try both snake_case and camelCase for keysValues
		var kvData map[string]any
		var ok bool
		if kvData, ok = routeData["event_type_key_values_filter"].(map[string]any); !ok {
			kvData, ok = routeData["keysValues"].(map[string]any)
		}

		if ok {
			// Helper to get value from keysValues map
			getKVValue := func(snakeCase, camelCase string) any {
				if val, ok := kvData[snakeCase]; ok {
					return val
				}
				return kvData[camelCase]
			}

			pairsList, _ := getKVValue("key_value_pairs", "keyValuePairs").([]any)
			keyValuePairs := make([]any, len(pairsList))
			for j, pair := range pairsList {
				pairData, _ := pair.(map[string]any)
				keyValuePairs[j] = map[string]any{
					keyKey:   pairData["key"],
					valueKey: pairData["value"],
				}
			}
			routeMap[keysValuesKey] = []map[string]any{
				{
					keyValuePairsKey: keyValuePairs,
					evTypeKey:        getKVValue("event_type", "eventType"),
				},
			}
		}

		routes[i] = routeMap
	}
	setData(d, data, routesKey, routes)
}

func validateProviderOneOf(providerTypes []string) schema.CustomizeDiffFunc {
	return func(_ context.Context, d *schema.ResourceDiff, _ any) error {
		providers := d.Get(providersKey).([]any)
		for i, p := range providers {
			if p == nil {
				continue
			}
			providerMap := p.(map[string]any)
			count := 0
			for _, value := range providerTypes {
				if _, ok := providerMap[value]; ok && providerMap[value] != nil {
					switch v := providerMap[value].(type) {
					case []any:
						if len(v) > 0 {
							count++
						}
					case []string:
						if len(v) > 0 {
							count++
						}
					}
				}
			}
			if count != 1 {
				return fmt.Errorf("exactly one of providers must be specified in providers[%d]", i)
			}
		}
		return nil
	}
}
