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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

const (
	authzJSONConfigKey = "json"
	authzTagsKey       = "tags"
	authzStatusKey     = "status"
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

			authzJSONConfigKey: {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
					validation.StringIsJSON,
				),
				Description: "Configuration of Authorization Policy in JSON format, the same one exported by Console.",
			},
			authzStatusKey: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(getMapStringKeys(AuthorizationPolicyStatusTypes), false),
				Description:  "Status of the Authorization Policy. active, inactive",
			},
			authzTagsKey: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Tags of the Authorization Policy.",
			},
		},
	}
}

func resourceAuthorizationPolicyFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) (d diag.Diagnostics) {
	policy := resp.GetConfigNode().GetAuthorizationPolicyConfig().GetPolicy()
	if policy == "" {
		return append(d, buildPluginError("config in the response is not valid AuthorizationPolicyConfig"))
	}
	setData(&d, data, authzJSONConfigKey, policy)

	status := resp.GetConfigNode().GetAuthorizationPolicyConfig().GetStatus()
	if status == configpb.AuthorizationPolicyConfig_STATUS_INVALID {
		return append(d, buildPluginError("status in the response is not valid"))
	}

	statusKey, exist := ReverseProtoEnumMap(AuthorizationPolicyStatusTypes)[status]
	if !exist {
		d = append(d, buildPluginError("unsupported Policy Status Type: "+status.String()))
	}
	setData(&d, data, authzStatusKey, statusKey)

	tags := resp.GetConfigNode().GetAuthorizationPolicyConfig().GetTags()
	setData(&d, data, authzTagsKey, tags)

	return d
}

func authorizationPolicyConfigBuilder(data *schema.ResourceData) *configpb.AuthorizationPolicyConfig {
	cfg := &configpb.AuthorizationPolicyConfig{
		Policy: data.Get(authzJSONConfigKey).(string),
		Status: AuthorizationPolicyStatusTypes[data.Get(authzStatusKey).(string)],
		Tags:   rawArrayToStringArray(data.Get(authzTagsKey).([]interface{})),
	}
	return cfg
}

func resourceAuthorizationPolicyBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *metaContext,
	builder *config.NodeRequest,
) {
	cfg := authorizationPolicyConfigBuilder(data)
	builder.WithAuthorizationPolicyConfig(cfg)
}
