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
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	providersKey        = "providers"
	routesKey           = "routes"
	providerNameKey     = "provider_name"
	providerIDKey       = "provider_id"
	stopProcessingKey   = "stop_processing"
	keysValuesKey       = "keys_values_filter"
	keyValuePairsKey    = "key_value_pairs"
	providerKey         = "provider"
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
	supportedFilters    = `
## Supported filters

| **Method** | **Event Type** | **Key** | **Value (example)** |
| --- | --- | --- | --- |
|  | **Ingest Events** |  |  |
| **IngestRecord, StreamRecords, Ingest (internal)** | indykite.audit.capture.upsert.node | captureLabel | Car |
|  |  | captureLabel | Green |
|  | indykite.audit.capture.upsert.relationship | captureLabel | RENT |
|  | indykite.audit.capture.delete.node | captureLabel | Car |
|  |  | captureLabel | Green |
|  | indykite.audit.capture.delete.relationship | captureLabel | RENT |
|  | indykite.audit.capture.delete.node.property |  |  |
|  | indykite.audit.capture.delete.relationship.property |  |  |
| **BatchUpsertNodes** | indykite.audit.capture.batch.upsert.node | captureLabel | Car |
|  |  | captureLabel | Green |
| **BatchUpsertRelationships** | indykite.audit.capture.batch.upsert.relationship | captureLabel | RENT |
| **BatchDeleteNodes** | indykite.audit.capture.batch.delete.node | captureLabel | Car |
|  |  | captureLabel | Green |
| **BatchDeleteRelationships** | indykite.audit.capture.batch.delete.relationship | captureLabel | RENT |
| **BatchDeleteNodeProperties** | indykite.audit.capture.batch.delete.node.property |  |  |
| **BatchDeleteRelationshipProperties** | indykite.audit.capture.delete.relationship.property |  |  |
| **BatchDeleteNodeTags** | indykite.audit.capture.batch.delete.node.tag | captureLabel | Car |
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
	readContext := configReadContextFunc(resourceEventSinkFlatten)
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

		CreateContext: configCreateContextFunc(resourceEventSinkBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceEventSinkBuild, readContext),
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

func resourceEventSinkFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) diag.Diagnostics {
	var d diag.Diagnostics
	eventSink := resp.GetConfigNode().GetEventSinkConfig()
	var results []map[string]any
	for key, p := range eventSink.GetProviders() {
		switch p.GetProvider().(type) {
		case *configpb.EventSinkConfig_Provider_Kafka:
			var oldPassword any
			if val, ok := data.Get(providersKey).([]any); ok && len(val) > 0 {
				for _, y := range val {
					if kafkaList, ok := y.(map[string]any)[kafkaKey].([]any); ok && len(kafkaList) > 0 {
						kafka := kafkaList[0].(map[string]any)
						oldPassword = kafka[passwordKey]
					}
				}
			}
			kafkaConfig := map[string]any{
				brokersKey:  p.GetKafka().GetBrokers(),
				topicKey:    p.GetKafka().GetTopic(),
				usernameKey: p.GetKafka().GetUsername(),
				// Password is not retrieved from response, but omitting here would result in removing.
				// First read old value and set it here too.
				passwordKey:        oldPassword,
				disableTLSKey:      p.GetKafka().GetDisableTls(),
				tlsSkipVerifyKey:   p.GetKafka().GetTlsSkipVerify(),
				providerDisplayKey: p.GetKafka().GetDisplayName().GetValue(),
			}
			result := map[string]any{
				providerNameKey: key,
				kafkaKey:        []any{kafkaConfig},
			}
			results = append(results, result)
		case *configpb.EventSinkConfig_Provider_AzureEventGrid:
			var oldAccessKey any
			if val, ok := data.Get(providersKey).([]any); ok && len(val) > 0 {
				for _, y := range val {
					if gridList, ok := y.(map[string]any)[azureEventGridKey].([]any); ok && len(gridList) > 0 {
						grid := gridList[0].(map[string]any)
						oldAccessKey = grid[accessKey]
					}
				}
			}
			gridConfig := map[string]any{
				topicEndpointKey:   p.GetAzureEventGrid().GetTopicEndpoint(),
				accessKey:          oldAccessKey,
				providerDisplayKey: p.GetAzureEventGrid().GetDisplayName().GetValue(),
			}
			result := map[string]any{
				providerNameKey:   key,
				azureEventGridKey: []any{gridConfig},
			}
			results = append(results, result)
		case *configpb.EventSinkConfig_Provider_AzureServiceBus:
			var oldConnection any
			if val, ok := data.Get(providersKey).([]any); ok && len(val) > 0 {
				for _, y := range val {
					if busList, ok := y.(map[string]any)[azureServiceBusKey].([]any); ok && len(busList) > 0 {
						bus := busList[0].(map[string]any)
						oldConnection = bus[connectionStringKey]
					}
				}
			}
			busConfig := map[string]any{
				connectionStringKey: oldConnection,
				queueKey:            p.GetAzureServiceBus().GetQueueOrTopicName(),
				providerDisplayKey:  p.GetAzureServiceBus().GetDisplayName().GetValue(),
			}
			result := map[string]any{
				providerNameKey:    key,
				azureServiceBusKey: []any{busConfig},
			}
			results = append(results, result)
		}
	}
	setData(&d, data, providersKey, results)

	routes := make([]any, len(eventSink.GetRoutes()))
	for i, route := range eventSink.GetRoutes() {
		switch filter := route.GetFilter().(type) {
		case *configpb.EventSinkConfig_Route_KeysValues:
			keyValuePairs := make([]any, len(route.GetKeysValues().GetKeyValuePairs()))
			for i, pair := range route.GetKeysValues().GetKeyValuePairs() {
				keyValuePairs[i] = map[string]any{
					keyKey:   pair.GetKey(),
					valueKey: pair.GetValue(),
				}
			}
			routes[i] = map[string]any{
				providerIDKey:     route.GetProviderId(),
				stopProcessingKey: route.GetStopProcessing(),
				keysValuesKey: []map[string]any{
					{
						keyValuePairsKey: keyValuePairs,
						evTypeKey:        filter.KeysValues.GetEventType(),
					},
				},
				routeDisplayKey: route.GetDisplayName().GetValue(),
				routeIDKey:      route.GetId().GetValue(),
			}
		default:
			return append(d, buildPluginError(fmt.Sprintf("unsupported EventSink Filter: %T", route)))
		}
	}
	setData(&d, data, routesKey, routes)
	return d
}

func resourceEventSinkBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	providers := data.Get(providersKey).([]any)
	routes := data.Get(routesKey).([]any)
	cfg := &configpb.EventSinkConfig{
		Providers: make(map[string]*configpb.EventSinkConfig_Provider, len(providers)),
		Routes:    make([]*configpb.EventSinkConfig_Route, len(routes)),
	}
	for _, provider := range providers {
		item, ok := provider.(map[string]any)
		if !ok {
			continue
		}
		key := item[providerNameKey].(string)
		if kafkaList, ok := item[kafkaKey].([]any); ok && len(kafkaList) > 0 {
			kafkaData := kafkaList[0].(map[string]any)
			cfg.Providers[key] = &configpb.EventSinkConfig_Provider{
				Provider: &configpb.EventSinkConfig_Provider_Kafka{
					Kafka: &configpb.KafkaSinkConfig{
						Brokers:  rawArrayToTypedArray[string](kafkaData[brokersKey]),
						Topic:    kafkaData[topicKey].(string),
						Username: kafkaData[usernameKey].(string),
						Password: kafkaData[passwordKey].(string),
						DisableTls: func() bool {
							if val, ok := kafkaData[disableTLSKey].(bool); ok {
								return val
							}
							return false
						}(),
						TlsSkipVerify: func() bool {
							if val, ok := kafkaData[tlsSkipVerifyKey].(bool); ok {
								return val
							}
							return false
						}(),
						DisplayName: func() *wrapperspb.StringValue {
							v, ok := kafkaData[providerDisplayKey].(string)
							if !ok || v == "" {
								return nil
							}
							return wrapperspb.String(v)
						}(),
					},
				},
			}
			continue
		}
		if gridList, ok := item[azureEventGridKey].([]any); ok && len(gridList) > 0 {
			gridData := gridList[0].(map[string]any)
			cfg.Providers[key] = &configpb.EventSinkConfig_Provider{
				Provider: &configpb.EventSinkConfig_Provider_AzureEventGrid{
					AzureEventGrid: &configpb.AzureEventGridSinkConfig{
						TopicEndpoint: gridData[topicEndpointKey].(string),
						AccessKey:     gridData[accessKey].(string),
						DisplayName: func() *wrapperspb.StringValue {
							v, ok := gridData[providerDisplayKey].(string)
							if !ok || v == "" {
								return nil
							}
							return wrapperspb.String(v)
						}(),
					},
				},
			}
			continue
		}
		if busList, ok := item[azureServiceBusKey].([]any); ok && len(busList) > 0 {
			busData := busList[0].(map[string]any)
			cfg.Providers[key] = &configpb.EventSinkConfig_Provider{
				Provider: &configpb.EventSinkConfig_Provider_AzureServiceBus{
					AzureServiceBus: &configpb.AzureServiceBusSinkConfig{
						ConnectionString: busData[connectionStringKey].(string),
						QueueOrTopicName: busData[queueKey].(string),
						DisplayName: func() *wrapperspb.StringValue {
							v, ok := busData[providerDisplayKey].(string)
							if !ok || v == "" {
								return nil
							}
							return wrapperspb.String(v)
						}(),
					},
				},
			}
		}
	}
	cfg.Routes = getRoutes(data)
	builder.WithEventSinkConfig(cfg)
}

func getRoutes(data *schema.ResourceData) []*configpb.EventSinkConfig_Route {
	routesSet, ok := data.Get(routesKey).([]any)
	if !ok {
		return nil
	}
	var routes = make([]*configpb.EventSinkConfig_Route, len(routesSet))
	for i, o := range routesSet {
		item, exists := o.(map[string]any)
		if !exists {
			continue
		}
		val, ok := item[keysValuesKey]
		if !ok {
			continue
		}
		list, ok := val.([]any)
		if !ok {
			continue
		}

		if len(list) > 0 {
			routes[i] = &configpb.EventSinkConfig_Route{
				ProviderId:     item[providerIDKey].(string),
				StopProcessing: item[stopProcessingKey].(bool),
				Filter: &configpb.EventSinkConfig_Route_KeysValues{
					KeysValues: &configpb.EventSinkConfig_Route_EventTypeKeysValues{
						KeyValuePairs: getKeyValuePairs(item),
						EventType:     checkEventType(item),
					},
				},
				DisplayName: func() *wrapperspb.StringValue {
					v, ok := item[routeDisplayKey].(string)
					if !ok || v == "" {
						return nil
					}
					return wrapperspb.String(v)
				}(),
				Id: func() *wrapperspb.StringValue {
					v, ok := item[routeIDKey].(string)
					if !ok || v == "" {
						return nil
					}
					return wrapperspb.String(v)
				}(),
			}
		}
	}
	return routes
}

func validateProviderOneOf(providerTypes []string) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, _ any) error {
		_ = ctx

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

func checkEventType(item map[string]any) string {
	var eventType string
	if innerMap, ok := item[keysValuesKey].([]any)[0].(map[string]any); ok {
		if inn, ok := innerMap[evTypeKey]; ok {
			if str, ok := inn.(string); ok {
				eventType = str
			}
		}
	}
	return eventType
}

func getKeyValuePairs(item map[string]any) []*configpb.EventSinkConfig_Route_KeyValuePair {
	innerMap, ok := item[keysValuesKey].([]any)[0].(map[string]any)
	if !ok {
		return nil
	}
	if pairsMap, ok := innerMap[keyValuePairsKey]; ok {
		var pairs = make([]*configpb.EventSinkConfig_Route_KeyValuePair, len(pairsMap.([]any)))
		pairsSet, ok := pairsMap.([]any)
		if !ok {
			return nil
		}
		for i, o := range pairsSet {
			if pairMap, ok := o.(map[string]any); ok {
				pairs[i] = &configpb.EventSinkConfig_Route_KeyValuePair{
					Key:   pairMap[keyKey].(string),
					Value: pairMap[valueKey].(string),
				}
			}
		}
		return pairs
	}
	return nil
}
