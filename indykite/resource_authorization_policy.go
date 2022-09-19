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
	authorizationPolicyJSONConfigKey = "json_config"
)

func resourceAuthorizationPolicy() *schema.Resource {
	readContext := configReadContextFunc(resourceAuthorizationPolicyFlatten)

	return &schema.Resource{
		CreateContext: configCreateContextFunc(resourceAuthorizationPolicyBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceAuthorizationPolicyBuild, readContext),
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

			authorizationPolicyJSONConfigKey: {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				ValidateFunc: validation.All(
					validation.StringIsJSON,
					authorizationPolicyValidateJSON,
				),
				Description: "Configuration of Authorization Policy in JSON format, the same one exported by Console.",
			},
		},
	}
}

func resourceAuthorizationPolicyFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) (d diag.Diagnostics) {
	clientConf := resp.GetConfigNode().GetAuthorizationPolicyConfig()
	if clientConf == nil {
		return append(d, buildPluginError("config in the response is not valid AuthorizationPolicyConfig"))
	}

	jsonVal, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(clientConf)
	if err != nil {
		return append(d, buildPluginError("failed to marshall message into JSON: "+err.Error()))
	}

	Set(&d, data, authorizationPolicyJSONConfigKey, string(jsonVal))

	return d
}

func authorizationPolicyConfigUnamrshalJSON(jsonVal string) (*configpb.AuthorizationPolicyConfig, error) {
	cfg := &configpb.AuthorizationPolicyConfig{}
	err := protojson.Unmarshal([]byte(jsonVal), cfg)
	return cfg, err
}

func resourceAuthorizationPolicyBuild(
	d *diag.Diagnostics,
	data *schema.ResourceData,
	_ *MetaContext,
	builder *config.NodeRequest,
) {
	cfg, err := authorizationPolicyConfigUnamrshalJSON(data.Get(authorizationPolicyJSONConfigKey).(string))
	if err != nil {
		*d = append(*d, buildPluginErrorWithPath(
			"Failed to Unmarshal AuthorizationPolicy config JSON into Proto message",
			authorizationPolicyJSONConfigKey,
		))
		return
	}
	builder.WithAuthorizationPolicyConfig(cfg)
}

func authorizationPolicyValidateJSON(val interface{}, key string) (warnings []string, errors []error) {
	v, ok := val.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", key))
		return warnings, errors
	}

	cfg, err := authorizationPolicyConfigUnamrshalJSON(v)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q cannot be unmarshalled into Proto message: %s", key, err))
		return warnings, errors
	}

	err = BetterValidationErrorWithPath(cfg.Validate())
	if err != nil {
		errors = append(errors, fmt.Errorf("%q has %s", key, err.Error()))
	}

	return warnings, errors
}
