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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

const (
	trustScoreProfileNodeClassification = "node_classification"
	trustScoreProfileDimensionsKey      = "dimensions"
	trustScoreProfileSchedule           = "schedule"
	trustScoreProfileName               = "name"
	trustScoreProfileWeight             = "weight"
)

func resourceTrustScoreProfile() *schema.Resource {
	readContext := configReadContextFunc(resourceTrustScoreProfileFlatten)

	return &schema.Resource{
		Description: "The TrustScoreProfile enables data consumers to evaluate the reliability of data, given the intent.  " +
			"By leveraging it, they can ensure that the data is suitable for effective use in downstream processes,  " +
			"increasing data quality and reducing data risk. ",

		CreateContext: configCreateContextFunc(resourceTrustScoreProfileBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceTrustScoreProfileBuild, readContext),
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

			trustScoreProfileNodeClassification: {
				Type:        schema.TypeString,
				Description: "NodeClassification is a node label in PascalCase, cannot be modified once set.",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
				),
			},
			trustScoreProfileDimensionsKey: {
				Type:        schema.TypeList,
				Description: "List of dimensions that will be used to calculate the trust score.",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						trustScoreProfileName: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the trust score dimension, must be one of the predefined names.",
						},
						trustScoreProfileWeight: {
							Type:         schema.TypeFloat,
							Required:     true,
							Description:  "Weight represents how relevant the dimension is in the trust score calculation.",
							ValidateFunc: validation.FloatBetween(0, 1),
						},
					}},
				MinItems: 1,
			},
			trustScoreProfileSchedule: {
				Type:        schema.TypeString,
				Description: "Schedule sets the time between re-calculations, must be one of the predefined intervals.",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice(getMapStringKeys(
						TrustScoreProfileScheduleFrequencies), false),
				},
			},
		},
	}
}

func resourceTrustScoreProfileFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) diag.Diagnostics {
	var d diag.Diagnostics

	trustScoreProfile := resp.GetConfigNode().GetTrustScoreProfileConfig()

	nodeClassification := trustScoreProfile.GetNodeClassification()
	setData(&d, data, trustScoreProfileNodeClassification, nodeClassification)

	dimensions := make([]any, len(trustScoreProfile.GetDimensions()))
	for i, dim := range trustScoreProfile.GetDimensions() {
		dimensions[i] = map[string]any{
			trustScoreProfileName:   ReverseProtoEnumMap(TrustScoreDimensionNames)[dim.GetName()],
			trustScoreProfileWeight: float32(dim.GetWeight()),
		}
	}
	setData(&d, data, trustScoreProfileDimensionsKey, dimensions)

	schedule := resp.GetConfigNode().GetTrustScoreProfileConfig().GetSchedule()
	if schedule == configpb.TrustScoreProfileConfig_UPDATE_FREQUENCY_INVALID {
		return append(d, buildPluginError("schedule in the response is not valid"))
	}
	scheduleKey, exist := ReverseProtoEnumMap(TrustScoreProfileScheduleFrequencies)[schedule]
	if !exist {
		d = append(d, buildPluginError("unsupported Frequency Type: "+schedule.String()))
	}
	setData(&d, data, trustScoreProfileSchedule, scheduleKey)

	return d
}

func resourceTrustScoreProfileBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	cfg := &configpb.TrustScoreProfileConfig{
		NodeClassification: data.Get(trustScoreProfileNodeClassification).(string),
		Dimensions:         getDimensions(data),
		Schedule:           TrustScoreProfileScheduleFrequencies[data.Get(trustScoreProfileSchedule).(string)],
	}

	builder.WithTrustScoreProfileConfig(cfg)
}

func getDimensions(data *schema.ResourceData) []*configpb.TrustScoreDimension {
	dimensionsSet := data.Get(trustScoreProfileDimensionsKey).([]any)
	var dimensions = make([]*configpb.TrustScoreDimension, len(dimensionsSet))
	for i, o := range dimensionsSet {
		item, ok := o.(map[string]any)
		if !ok {
			continue
		}
		dimensions[i] = &configpb.TrustScoreDimension{
			Name:   TrustScoreDimensionNames[item["name"].(string)],
			Weight: float32(item["weight"].(float64)),
		}
	}
	return dimensions
}
