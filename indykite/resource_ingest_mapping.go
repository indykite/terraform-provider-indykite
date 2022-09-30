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
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/jarvis-sdk-go/config"
	configpb "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	ingestMappingJSONConfigKey = "json_config"
)

func resourceIngestMapping() *schema.Resource {
	readContext := configReadContextFunc(resourceIngestMappingFlatten)

	return &schema.Resource{
		CreateContext: configCreateContextFunc(resourceIngestMappingBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceIngestMappingBuild, readContext),
		DeleteContext: configDeleteContextFunc(),
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:    locationSchema(),
			customerIDKey:  setComputed(customerIDSchema()),
			appSpaceIDKey:  setComputed(appSpaceIDSchema()),
			tenantIDKey:    setComputed(tenantIDSchema()),
			nameKey:        nameSchema(),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),

			ingestMappingJSONConfigKey: {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				ValidateFunc:     validation.All(validation.StringIsJSON, ingestMappingValidateJSON),
				Description:      "Configuration of Ingest mapping in JSON format, the same one exported by Console.",
			},
		},
	}
}

func resourceIngestMappingFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) (d diag.Diagnostics) {
	clientConf := resp.GetConfigNode().GetIngestMappingConfig()
	if clientConf == nil {
		return append(d, buildPluginError("config in the response is not valid IngestMappingConfig"))
	}

	jsonVal, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(clientConf)
	if err != nil {
		return append(d, buildPluginError("failed to marshall message into JSON: "+err.Error()))
	}

	setData(&d, data, ingestMappingJSONConfigKey, string(jsonVal))

	return d
}

func ingestMappingConfigUnamrshalJSON(jsonVal string) (*configpb.IngestMappingConfig, error) {
	cfg := &configpb.IngestMappingConfig{}
	err := protojson.Unmarshal([]byte(jsonVal), cfg)
	return cfg, err
}

func resourceIngestMappingBuild(
	d *diag.Diagnostics,
	data *schema.ResourceData,
	_ *metaContext,
	builder *config.NodeRequest,
) {
	cfg, err := ingestMappingConfigUnamrshalJSON(data.Get(ingestMappingJSONConfigKey).(string))
	if err != nil {
		*d = append(*d, buildPluginErrorWithAttrName(
			"Failed to Unmarshal IngestMapping config JSON into Proto message",
			ingestMappingJSONConfigKey,
		))
		return
	}
	builder.WithIngestMappingConfig(cfg)
}

func ingestMappingValidateJSON(val interface{}, key string) (warnings []string, errors []error) {
	v, ok := val.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", key))
		return warnings, errors
	}

	cfg, err := ingestMappingConfigUnamrshalJSON(v)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q cannot be unmarshalled into Proto message: %s", key, err))
		return warnings, errors
	}

	err = betterValidationErrorWithPath(cfg.Validate())
	if err != nil {
		errors = append(errors, fmt.Errorf("%q has %s", key, err.Error()))
	}

	return warnings, errors
}
