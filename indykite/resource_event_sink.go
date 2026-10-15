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
	"slices"

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
	pubsubKey           = "pubsub"
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
	credentialsJSONKey  = "credentials_json"
	topicNameKey        = "topic_name"
	providerDisplayKey  = "provider_display_name"
	routeDisplayKey     = "route_display_name"
	routeIDKey          = "route_id"
	includeCdcEventsKey = "include_cdc_events"
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
| **BatchDeleteNodePropertyMetadata** | indykite.audit.capture.delete.node.property.metadata |  |  |
| All Capture Events | indykite.audit.capture.* |  |  |
|  | **Configuration Events** |  |  |
| Config | indykite.audit.config.create |  |  |
|  | indykite.audit.config.read |  |  |
|  | indykite.audit.config.update |  |  |
|  | indykite.audit.config.delete |  |  |
|  | indykite.audit.config.list |  |  |
|  | indykite.audit.config.permission.assign |  |  |
|  | indykite.audit.config.permission.revoke |  |  |
| All Config Events | indykite.audit.config.* |  |  |
|  | **Token Events** |  |  |
| TokenIntrospect | indykite.audit.credentials.token.introspected |  |  |
|  | **Authorization Events** |  |  |
| Authorization | indykite.audit.authorization.evaluation |  |  |
|  | indykite.audit.authorization.evaluations |  |  |
|  | indykite.audit.authorization.searchsubject |  |  |
|  | indykite.audit.authorization.searchresource |  |  |
|  | indykite.audit.authorization.searchaction |  |  |
| All Authorization Events | indykite.audit.authorization.* |  |  |
|  | **Ciq Events** |  |  |
| Ciq | indykite.audit.ciq.execute |  |  |
|  | **CDC Events** (requires include_cdc_events = true) |  |  |
| CDC | indykite.audit.cdc.node.create | label0 | Person |
|  | indykite.audit.cdc.node.update | label0 | Person |
|  | indykite.audit.cdc.node.delete | label0 | Person |
|  | indykite.audit.cdc.relationship.create |  |  |
|  | indykite.audit.cdc.relationship.update |  |  |
|  | indykite.audit.cdc.relationship.delete |  |  |
| All CDC Events | indykite.audit.cdc.* |  |  |
|  |  |  |  |

CDC node events carry the node labels as keys label0, label1, ... in alphabetical order, one key per label.`
)

func resourceEventSink() *schema.Resource {
	providerOneOf := []string{kafkaKey, azureEventGridKey, azureServiceBusKey, pubsubKey}

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
			StateContext: projectStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:   locationSchema(),
			projectIDKey:  projectIDSchema(),
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
			includeCdcEventsKey: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "When true, CDC (Change Data Capture) events will be emitted to this event sink. " +
					"When false or unset, CDC events will not be emitted. " +
					"Defaults to false for backward compatibility.",
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
						Type: schema.TypeList,
						Elem: &schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateMinCharLength(8),
						},
						Required:    true,
						MinItems:    1,
						Description: "Brokers specify Kafka destinations to connect to, each at least 8 characters.",
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
		pubsubKey: {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "PubSubSinkConfig (Google Cloud Pub/Sub)",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					projectIDKey: {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.All(
							validation.StringLenBetween(6, 30),
						),
					},
					topicNameKey: {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.All(
							validation.StringLenBetween(3, 255),
						),
					},
					credentialsJSONKey: {
						Type:      schema.TypeString,
						Required:  true,
						Sensitive: true,
						ValidateFunc: validation.All(
							validation.StringLenBetween(1, 10240),
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
						Description: "Last error message from the Pub/Sub sink",
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
			Type:        schema.TypeList,
			Elem:        &schema.Resource{Schema: keysValuesSchema()},
			Description: "Filter selecting the events routed to the provider. The API requires it on every route.",
			Required:    true,
			MaxItems:    1,
		},
		routeDisplayKey: {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(2, 254),
			),
		},
		routeIDKey: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Kept in Terraform state only; it is not sent to the API, which has no route identifier.",
			Deprecated: "route_id is not stored by the API and has no effect. " +
				"It will be removed in a future major version.",
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
						Type:         schema.TypeString,
						Required:     true,
						Description:  "Key for the ingest eventType",
						ValidateFunc: validation.StringIsNotEmpty,
					},
					valueKey: {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "Value for the ingest eventType",
						ValidateFunc: validation.StringIsNotEmpty,
					},
				},
			},
		},
		evTypeKey: {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.All(
				validation.StringMatch(
					regexp.MustCompile(`^[a-zA-Z0-9_*\.]+$`), //nolint:gocritic // keep the escaped dot for readability
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
		ProjectID:        projectIDFromData(data),
		Name:             data.Get(nameKey).(string),
		DisplayName:      stringValue(optionalString(data, displayNameKey)),
		Description:      stringValue(optionalString(data, descriptionKey)),
		Providers:        buildProvidersMap(providers),
		Routes:           buildRoutesList(routes),
		IncludeCDCEvents: data.Get(includeCdcEventsKey).(bool),
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
	setProjectIDData(&d, data, resp.AppSpaceID, resp.CustomerID)

	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)

	flattenEventSinkConfig(&d, data, eventSinkConfigFromResponse(&resp))

	return d
}

// eventSinkConfigFromResponse returns providers, routes and include_cdc_events of the response.
// The API returns them at the top level; the legacy nested "config" envelope is used as a fallback.
func eventSinkConfigFromResponse(resp *EventSinkResponse) map[string]any {
	if resp.Providers == nil && resp.Routes == nil && resp.IncludeCDCEvents == nil {
		return resp.Config
	}
	config := map[string]any{
		"providers": resp.Providers,
		"routes":    resp.Routes,
	}
	if resp.IncludeCDCEvents != nil {
		config["include_cdc_events"] = *resp.IncludeCDCEvents
	}
	return config
}

func resEventSinkUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	if hasNameChanged(&d, data) {
		return d
	}

	providers := data.Get(providersKey).([]any)
	routes := data.Get(routesKey).([]any)

	req := UpdateEventSinkRequest{
		DisplayName:      updateOptionalString(data, displayNameKey),
		Description:      updateOptionalString(data, descriptionKey),
		Providers:        buildProvidersMap(providers),
		Routes:           buildRoutesList(routes),
		IncludeCDCEvents: data.Get(includeCdcEventsKey).(bool),
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
	if pubsubList, ok := item[pubsubKey].([]any); ok && len(pubsubList) > 0 {
		return buildPubSubProviderMap(pubsubList[0].(map[string]any))
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

// buildPubSubProviderMap builds Google Cloud Pub/Sub configuration map.
func buildPubSubProviderMap(pubsubData map[string]any) map[string]any {
	return map[string]any{
		"pubsub": map[string]any{
			"project_id":       pubsubData[projectIDKey],
			"topic_name":       pubsubData[topicNameKey],
			"credentials_json": pubsubData[credentialsJSONKey],
			"display_name":     pubsubData[providerDisplayKey],
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
	}

	if kvList, ok := item[keysValuesKey].([]any); ok && len(kvList) > 0 {
		routeMap["event_type_key_values_filter"] = buildKeysValuesMap(kvList[0].(map[string]any))
	}

	return routeMap
}

// buildKeysValuesMap builds the event_type_key_values_filter map for a route.
// The API names the key/value pairs context_key_value and silently drops unknown fields.
func buildKeysValuesMap(kvData map[string]any) map[string]any {
	filter := map[string]any{
		"event_type": kvData[evTypeKey],
	}
	pairsList, _ := kvData[keyValuePairsKey].([]any)
	pairs := make([]any, 0, len(pairsList))
	for _, pair := range pairsList {
		if pairData, ok := pair.(map[string]any); ok {
			pairs = append(pairs, map[string]any{
				"key":   pairData[keyKey],
				"value": pairData[valueKey],
			})
		}
	}
	if len(pairs) > 0 {
		filter["context_key_value"] = pairs
	}
	return filter
}

// buildKafkaConfig builds Kafka configuration preserving sensitive password from the prior state block.
func buildKafkaConfig(kafkaData, prior map[string]any) map[string]any {
	oldPassword := prior[passwordKey]

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
		lastErrorKey:       getValue("last_error", "lastError"),
	}
}

// buildAzureEventGridConfig builds Azure Event Grid configuration preserving sensitive access key
// from the prior state block.
func buildAzureEventGridConfig(gridData, prior map[string]any) map[string]any {
	oldAccessKey := prior[accessKey]

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
		lastErrorKey:       getValue("last_error", "lastError"),
	}
}

// buildAzureServiceBusConfig builds Azure Service Bus configuration preserving sensitive connection
// string from the prior state block.
func buildAzureServiceBusConfig(busData, prior map[string]any) map[string]any {
	oldConnection := prior[connectionStringKey]

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
		lastErrorKey:        getValue("last_error", "lastError"),
	}
}

// buildPubSubConfig builds Google Cloud Pub/Sub configuration preserving sensitive credentials
// from the prior state block.
func buildPubSubConfig(pubsubData, prior map[string]any) map[string]any {
	oldCredentials := prior[credentialsJSONKey]

	// Helper to get value from either snake_case or camelCase
	getValue := func(snakeCase, camelCase string) any {
		if val, ok := pubsubData[snakeCase]; ok {
			return val
		}
		return pubsubData[camelCase]
	}

	return map[string]any{
		projectIDKey:       getValue("project_id", "projectId"),
		topicNameKey:       getValue("topic_name", "topicName"),
		credentialsJSONKey: oldCredentials,
		providerDisplayKey: getValue("display_name", "displayName"),
		lastErrorKey:       getValue("last_error", "lastError"),
	}
}

func flattenEventSinkConfig(d *diag.Diagnostics, data *schema.ResourceData, config map[string]any) {
	providersMap, _ := config["providers"].(map[string]any)
	setData(d, data, providersKey, flattenEventSinkProviders(data, providersMap))

	routesList, _ := config["routes"].([]any)
	setData(d, data, routesKey, flattenEventSinkRoutes(data, routesList))

	if v, ok := config["include_cdc_events"]; ok {
		setData(d, data, includeCdcEventsKey, v)
	} else if v, ok := config["includeCdcEvents"]; ok {
		setData(d, data, includeCdcEventsKey, v)
	}
}

// pickMap returns the first value, as a map, found under any of the given keys.
// Used to tolerate both snake_case and camelCase keys in API responses.
func pickMap(m map[string]any, keys ...string) (map[string]any, bool) {
	for _, k := range keys {
		if v, ok := m[k].(map[string]any); ok {
			return v, true
		}
	}
	return nil, false
}

// getEither returns the value of the first of the given keys present in m.
// Used to tolerate both current and legacy key names in API responses.
func getEither(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if val, ok := m[k]; ok {
			return val
		}
	}
	return nil
}

// orderedProviderNames returns the provider names of the API response in the order of the current
// state, followed by the remaining names sorted. The API returns providers as an unordered map,
// while providers is an ordered list in the schema.
func orderedProviderNames(data *schema.ResourceData, providersMap map[string]any) []string {
	names := make([]string, 0, len(providersMap))
	seen := make(map[string]bool, len(providersMap))
	stateProviders, _ := data.Get(providersKey).([]any)
	for _, p := range stateProviders {
		item, _ := p.(map[string]any)
		name, _ := item[providerNameKey].(string)
		if _, ok := providersMap[name]; ok && !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	rest := make([]string, 0, len(providersMap)-len(names))
	for name := range providersMap {
		if !seen[name] {
			rest = append(rest, name)
		}
	}
	slices.Sort(rest)
	return append(names, rest...)
}

// priorProviderBlock returns the typeKey block (e.g. kafka) of the provider named name from the
// current state. The API returns secrets empty, so they are preserved from this block.
func priorProviderBlock(data *schema.ResourceData, name, typeKey string) map[string]any {
	stateProviders, _ := data.Get(providersKey).([]any)
	for _, p := range stateProviders {
		item, _ := p.(map[string]any)
		if item[providerNameKey] != name {
			continue
		}
		if blocks, ok := item[typeKey].([]any); ok && len(blocks) > 0 {
			block, _ := blocks[0].(map[string]any)
			return block
		}
	}
	return nil
}

// flattenEventSinkProviders converts the API providers map into Terraform schema blocks.
func flattenEventSinkProviders(data *schema.ResourceData, providersMap map[string]any) []map[string]any {
	var results []map[string]any
	for _, key := range orderedProviderNames(data, providersMap) {
		providerData, _ := providersMap[key].(map[string]any)

		if kafkaData, ok := pickMap(providerData, "kafka"); ok {
			results = append(results, map[string]any{
				providerNameKey: key,
				kafkaKey: []any{buildKafkaConfig(kafkaData,
					priorProviderBlock(data, key, kafkaKey))},
			})
		}
		if gridData, ok := pickMap(providerData, "azure_event_grid", "azureEventGrid"); ok {
			results = append(results, map[string]any{
				providerNameKey: key,
				azureEventGridKey: []any{buildAzureEventGridConfig(gridData,
					priorProviderBlock(data, key, azureEventGridKey))},
			})
		}
		if busData, ok := pickMap(providerData, "azure_service_bus", "azureServiceBus"); ok {
			results = append(results, map[string]any{
				providerNameKey: key,
				azureServiceBusKey: []any{buildAzureServiceBusConfig(busData,
					priorProviderBlock(data, key, azureServiceBusKey))},
			})
		}
		if pubsubData, ok := pickMap(providerData, "pubsub", "pubSub"); ok {
			results = append(results, map[string]any{
				providerNameKey: key,
				pubsubKey: []any{buildPubSubConfig(pubsubData,
					priorProviderBlock(data, key, pubsubKey))},
			})
		}
	}
	return results
}

// flattenEventSinkRoutes converts the API routes list into Terraform schema blocks.
// The API keeps the route order and has no route identifier, so route_id is preserved
// from the route at the same position in the current state.
func flattenEventSinkRoutes(data *schema.ResourceData, routesList []any) []any {
	stateRoutes, _ := data.Get(routesKey).([]any)
	routes := make([]any, len(routesList))
	for i, r := range routesList {
		routeData, _ := r.(map[string]any)
		routeMap := flattenEventSinkRoute(routeData)
		if i < len(stateRoutes) {
			stateRoute, _ := stateRoutes[i].(map[string]any)
			routeMap[routeIDKey] = stateRoute[routeIDKey]
		}
		routes[i] = routeMap
	}
	return routes
}

// flattenEventSinkRoute converts a single API route into a Terraform schema block.
func flattenEventSinkRoute(routeData map[string]any) map[string]any {
	routeMap := map[string]any{
		providerIDKey:     getEither(routeData, "provider_id", "providerId"),
		stopProcessingKey: getEither(routeData, "stop_processing", "stopProcessing"),
		routeDisplayKey:   getEither(routeData, "display_name", "displayName"),
	}

	kvData, ok := pickMap(routeData, "event_type_key_values_filter", "keysValues")
	if !ok {
		return routeMap
	}

	pairsList, _ := getEither(kvData, "context_key_value", "key_value_pairs", "keyValuePairs").([]any)
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
			evTypeKey:        getEither(kvData, "event_type", "eventType"),
		},
	}
	return routeMap
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
