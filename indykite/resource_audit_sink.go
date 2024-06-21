// Copyright (c) 2023 IndyKite
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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

const (
	auditSinkKafkaProviderKey      = "kafka"
	auditSinkKafkaBrokersKey       = "brokers"
	auditSinkKafkaTopicKey         = "topic"
	auditSinkKafkaUsernameKey      = "username"
	auditSinkKafkaPasswordKey      = "password"
	auditSinkKafkaDisableTLSKey    = "disable_tls"
	auditSinkKafkaSkipTLSVerifyKey = "tls_skip_verify"
)

func resourceAuditSink() *schema.Resource {
	readContext := configReadContextFunc(resourceAuditSinkFlatten)

	return &schema.Resource{
		Description: `Audit Sink configuration can be used to forward audit logs to your own consumer.`,

		CreateContext: configCreateContextFunc(resourceAuditSinkBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceAuditSinkBuild, readContext),
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

			auditSinkKafkaProviderKey: {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						auditSinkKafkaBrokersKey: {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						auditSinkKafkaTopicKey:    {Type: schema.TypeString, Required: true},
						auditSinkKafkaUsernameKey: {Type: schema.TypeString, Required: true},
						auditSinkKafkaPasswordKey: {Type: schema.TypeString, Required: true, Sensitive: true},
						auditSinkKafkaDisableTLSKey: {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Disable TLS for communication. Highly NOT RECOMMENDED.",
						},
						auditSinkKafkaSkipTLSVerifyKey: {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Skip TLS certificate verification. NOT RECOMMENDED.",
						},
					},
				},
			},
		},
	}
}

func resourceAuditSinkFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) diag.Diagnostics {
	var d diag.Diagnostics
	switch p := resp.GetConfigNode().GetAuditSinkConfig().GetProvider().(type) {
	case *configpb.AuditSinkConfig_Kafka:
		var oldPassword any
		if val, ok := data.Get(auditSinkKafkaProviderKey).([]any); ok && len(val) > 0 {
			dataMap, _ := val[0].(map[string]any)
			oldPassword = dataMap[auditSinkKafkaPasswordKey]
		}
		setData(&d, data, auditSinkKafkaProviderKey, []map[string]any{{
			auditSinkKafkaBrokersKey:  p.Kafka.Brokers,
			auditSinkKafkaTopicKey:    p.Kafka.Topic,
			auditSinkKafkaUsernameKey: p.Kafka.Username,
			// Password is not retrieved from response, but omitting here would result in removing.
			// First read old value and set it here too.
			auditSinkKafkaPasswordKey:      oldPassword,
			auditSinkKafkaDisableTLSKey:    p.Kafka.DisableTls,
			auditSinkKafkaSkipTLSVerifyKey: p.Kafka.TlsSkipVerify,
		}})
	default:
		return append(d, buildPluginError(fmt.Sprintf("unsupported AuditSink Provider: %T", p)))
	}

	return d
}

func resourceAuditSinkBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	cfg := &configpb.AuditSinkConfig{}
	if val, ok := data.GetOk(auditSinkKafkaProviderKey); ok {
		mapVal := val.([]any)[0].(map[string]any)
		_ = mapVal
		cfg.Provider = &configpb.AuditSinkConfig_Kafka{
			Kafka: &configpb.KafkaSinkConfig{
				Brokers:       rawArrayToTypedArray[string](mapVal[auditSinkKafkaBrokersKey]),
				Topic:         mapVal[auditSinkKafkaTopicKey].(string),
				Username:      mapVal[auditSinkKafkaUsernameKey].(string),
				Password:      mapVal[auditSinkKafkaPasswordKey].(string),
				DisableTls:    mapVal[auditSinkKafkaDisableTLSKey].(bool),
				TlsSkipVerify: mapVal[auditSinkKafkaSkipTLSVerifyKey].(bool),
			},
		}
	}
	builder.WithAuditSinkConfig(cfg)
}
