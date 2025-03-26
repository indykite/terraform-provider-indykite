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
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

const (
	providersKey       = "providers"
	routesKey          = "routes"
	providerNameKey    = "provider_name"
	providerIDKey      = "provider_id"
	stopProcessingKey  = "stop_processing"
	eventTypeKey       = "event_type_filter"
	contextKeyValueKey = "context_key_value_filter"
	providerKey        = "provider"
	kafkaKey           = "kafka"
	brokersKey         = "brokers"
	topicKey           = "topic"
	keyKey             = "key"
	valueKey           = "value"
	disableTLSKey      = "disable_tls"
	tlsSkipVerifyKey   = "tls_skip_verify"
	usernameKey        = "username"
	passwordKey        = "password"
)

func resourceEventSink() *schema.Resource {
	readContext := configReadContextFunc(resourceEventSinkFlatten)

	return &schema.Resource{
		Description: `Event Sink configuration is used to configure outbound events.`,

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
			Description: "KafkaSinkConfig",
			Required:    true,
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
		eventTypeKey: {
			Type:     schema.TypeString,
			Optional: true,
		},
		contextKeyValueKey: {
			Type:     schema.TypeList,
			Elem:     &schema.Resource{Schema: keyValueSchema()},
			Optional: true,
			MaxItems: 1,
		},
	}
}

func keyValueSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		keyKey: {
			Type:     schema.TypeString,
			Required: true,
		},
		valueKey: {
			Type:     schema.TypeString,
			Required: true,
		},
	}
}

func resourceEventSinkFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) diag.Diagnostics {
	var d diag.Diagnostics
	eventSink := resp.GetConfigNode().GetEventSinkConfig()
	var results []map[string]any //nolint:prealloc // no prealloc
	for key, p := range eventSink.GetProviders() {
		var oldPassword any
		if val, ok := data.Get(providersKey).([]any); ok && len(val) > 0 {
			kafkaMap, _ := val[0].(map[string]any)[kafkaKey].([]any)[0].(map[string]any)
			oldPassword = kafkaMap[passwordKey]
		}
		kafkaConfig := map[string]any{
			brokersKey:  p.GetKafka().GetBrokers(),
			topicKey:    p.GetKafka().GetTopic(),
			usernameKey: p.GetKafka().GetUsername(),
			// Password is not retrieved from response, but omitting here would result in removing.
			// First read old value and set it here too.
			passwordKey: oldPassword,
		}

		if p.GetKafka().GetDisableTls() {
			kafkaConfig[disableTLSKey] = p.GetKafka().GetDisableTls()
		}
		if p.GetKafka().GetTlsSkipVerify() {
			kafkaConfig[tlsSkipVerifyKey] = p.GetKafka().GetTlsSkipVerify()
		}
		result := map[string]any{
			providerNameKey: key,
			kafkaKey:        []any{kafkaConfig},
		}
		results = append(results, result)
	}
	setData(&d, data, providersKey, results)

	routes := make([]any, len(eventSink.GetRoutes()))
	for i, route := range eventSink.GetRoutes() {
		switch filter := route.GetFilter().(type) {
		case *configpb.EventSinkConfig_Route_EventType:
			routes[i] = map[string]any{
				providerIDKey:     route.GetProviderId(),
				stopProcessingKey: route.GetStopProcessing(),
				eventTypeKey:      filter.EventType,
			}

		case *configpb.EventSinkConfig_Route_ContextKeyValue:
			routes[i] = map[string]any{
				providerIDKey:     route.GetProviderId(),
				stopProcessingKey: route.GetStopProcessing(),
				contextKeyValueKey: []map[string]any{
					{
						keyKey:   filter.ContextKeyValue.GetKey(),
						valueKey: filter.ContextKeyValue.GetValue(),
					},
				},
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
		item, ok := o.(map[string]any)
		if !ok {
			continue
		}
		// Handle eventTypeKey case
		if _, ok := item[eventTypeKey]; ok {
			routes[i] = &configpb.EventSinkConfig_Route{
				ProviderId:     item[providerIDKey].(string),
				StopProcessing: item[stopProcessingKey].(bool),
				Filter: &configpb.EventSinkConfig_Route_EventType{
					EventType: item[eventTypeKey].(string),
				},
			}
		}

		// Handle contextKeyValueKey case
		if val, ok := item[contextKeyValueKey]; ok {
			if list, ok := val.([]any); ok {
				if len(list) > 0 {
					routes[i] = &configpb.EventSinkConfig_Route{
						ProviderId:     item[providerIDKey].(string),
						StopProcessing: item[stopProcessingKey].(bool),
						Filter: &configpb.EventSinkConfig_Route_ContextKeyValue{
							ContextKeyValue: &configpb.EventSinkConfig_Route_KeyValue{
								Key:   item[contextKeyValueKey].([]any)[0].(map[string]any)[keyKey].(string),
								Value: item[contextKeyValueKey].([]any)[0].(map[string]any)[valueKey].(string),
							},
						},
					}
				}
			}
		}
	}
	return routes
}
