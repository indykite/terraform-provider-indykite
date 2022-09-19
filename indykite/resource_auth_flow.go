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
	"github.com/indykite/jarvis-sdk-go/config"
	configpb "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
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
	if resp == nil {
		return diag.Errorf("empty Auth Flow config")
	}
	authFlowNodeConfig := resp.ConfigNode.GetAuthFlowConfig()
	if authFlowNodeConfig == nil {
		return diag.Errorf("config in the response is not valid AuthFlowNodeConfig")
	}

	switch authFlowNodeConfig.SourceFormat {
	case configpb.AuthFlowConfig_FORMAT_BARE_JSON:
		Set(&d, data, authFlowJSONKey, string(authFlowNodeConfig.Source))
	case configpb.AuthFlowConfig_FORMAT_BARE_YAML:
		Set(&d, data, authFlowYamlKey, string(authFlowNodeConfig.Source))
	case configpb.AuthFlowConfig_FORMAT_RICH_JSON:
		Set(&d, data, authFlowHasUIJson, true)
	default:
		d = append(d, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "unsupported AuthFlow node config",
			Detail:   fmt.Sprintf("%T is not supported yet", authFlowNodeConfig.SourceFormat),
		})
	}

	return d
}

func resourceAuthFlowBuild(
	d *diag.Diagnostics,
	data *schema.ResourceData,
	meta *MetaContext,
	builder *config.NodeRequest,
) {
	cfg := new(configpb.AuthFlowConfig)

	if data.HasChange(authFlowHasUIJson) {
		*d = append(*d, diag.Diagnostic{
			Severity: diag.Error,
			Summary: fmt.Sprintf(
				"property %s is readonly and cannot be changed in the config", authFlowHasUIJson),
			AttributePath: cty.IndexStringPath(authFlowHasUIJson),
		})
		return
	}

	if data.Get(authFlowHasUIJson).(bool) {
		*d = append(*d, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Auth flow is managed by the Console UI and cannot be changed with Terraform",
			AttributePath: cty.IndexStringPath(authFlowHasUIJson),
		})
		return
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
