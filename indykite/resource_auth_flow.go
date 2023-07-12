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

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

const (
	authFlowJSONKey   = "json"
	authFlowYamlKey   = "yaml"
	authFlowHasUIJson = "has_ui"
)

func resourceAuthFlow() *schema.Resource {
	readContext := configReadContextFunc(resourceAuthFlowFlatten)

	oneOfConfigType := []string{authFlowJSONKey, authFlowYamlKey}
	return &schema.Resource{
		CreateContext: configCreateContextFunc(resourceAuthFlowBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceAuthFlowBuild, readContext),
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
			// config common schema ends here
			authFlowJSONKey: setExactlyOneOf(&schema.Schema{
				Type:             schema.TypeString,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				ValidateFunc:     validation.StringIsJSON,
				Description:      "The Auth Flow configuration in JSON format",
			}, authFlowJSONKey, oneOfConfigType),
			authFlowYamlKey: setExactlyOneOf(&schema.Schema{
				Type:             schema.TypeString,
				DiffSuppressFunc: SuppressYamlDiff,
				ValidateDiagFunc: ValidateYaml,
				Description:      "The Auth Flow configuration in YAML format, with possible references under '.references' key",
			}, authFlowYamlKey, oneOfConfigType),
			authFlowHasUIJson: {Type: schema.TypeBool, Computed: true},
		},
	}
}

func resourceAuthFlowFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) (d diag.Diagnostics) {
	authFlowNodeConfig := resp.GetConfigNode().GetAuthFlowConfig()
	if authFlowNodeConfig == nil {
		return diag.Diagnostics{buildPluginError("config in the response is not valid AuthFlowNodeConfig")}
	}

	hasUIFlow := false
	jsonData, yamlData := "", ""
	switch authFlowNodeConfig.SourceFormat {
	case configpb.AuthFlowConfig_FORMAT_BARE_JSON:
		jsonData = string(authFlowNodeConfig.Source)
	case configpb.AuthFlowConfig_FORMAT_BARE_YAML:
		yamlData = string(authFlowNodeConfig.Source)
	case configpb.AuthFlowConfig_FORMAT_RICH_JSON:
		hasUIFlow = true
	case configpb.AuthFlowConfig_FORMAT_INVALID:
		d = append(d, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "invalid auth flow was saved before, create auth flow again",
		})
	default:
		d = append(d,
			buildPluginError(fmt.Sprintf("Auth Flow type %T is not supported yet", authFlowNodeConfig.SourceFormat)),
		)
	}
	setData(&d, data, authFlowHasUIJson, hasUIFlow)
	setData(&d, data, authFlowJSONKey, jsonData)
	setData(&d, data, authFlowYamlKey, yamlData)

	return d
}

func resourceAuthFlowBuild(
	d *diag.Diagnostics,
	data *schema.ResourceData,
	meta *ClientContext,
	builder *config.NodeRequest,
) {
	cfg := new(configpb.AuthFlowConfig)

	// This function is executed only during APPLY, seems it is impossible to trigger Warning during PLAN
	if data.Get(authFlowHasUIJson).(bool) {
		*d = append(*d, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       "Auth flow was managed by the Console UI. This change will discard Console Diagram.",
			AttributePath: cty.IndexStringPath(authFlowHasUIJson),
		})
	}

	if val, ok := data.GetOk(authFlowJSONKey); ok {
		cfg.SourceFormat = configpb.AuthFlowConfig_FORMAT_BARE_JSON
		cfg.Source = []byte(val.(string))
	}
	if val, ok := data.GetOk(authFlowYamlKey); ok {
		cfg.SourceFormat = configpb.AuthFlowConfig_FORMAT_BARE_YAML
		cfg.Source = []byte(val.(string))
	}
	builder.WithAuthFlowConfig(cfg)
}
