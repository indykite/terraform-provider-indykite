// Copyright (c) 2024 IndyKite
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
	consentPurposeKey        = "purpose"
	consentApplicationIDKey  = "application_id"
	consentValidityPeriodKey = "validity_period"
	consentRevokeAfterUseKey = "revoke_after_use"
	consentDataPointsKey     = "data_points"
)

func resourceConsent() *schema.Resource {
	readContext := configReadContextFunc(resourceConsentFlatten)

	return &schema.Resource{
		CreateContext: configCreateContextFunc(resourceConsentBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceConsentBuild, readContext),
		DeleteContext: configDeleteContextFunc(),
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:             locationSchema(),
			customerIDKey:           setComputed(customerIDSchema()),
			appSpaceIDKey:           setComputed(appSpaceIDSchema()),
			nameKey:                 nameSchema(),
			displayNameKey:          displayNameSchema(),
			descriptionKey:          descriptionSchema(),
			createTimeKey:           createTimeSchema(),
			updateTimeKey:           updateTimeSchema(),
			consentApplicationIDKey: applicationIDSchema(),

			consentPurposeKey: {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Purpose of the consent",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			consentValidityPeriodKey: {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Specifies the duration in second that the consent remains valid, ranging from 1 day to 2 years",
				ValidateFunc: validation.IntBetween(86400, 63072000),
			},
			consentRevokeAfterUseKey: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set to true, the consent will be revoked after it is used once",
			},
			consentDataPointsKey: {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringIsNotEmpty,
						validation.StringIsJSON,
					)},
				Description:      "Data points is a list of properties related to the Digital twin that the consent is for",
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
		},
	}
}

func resourceConsentFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) diag.Diagnostics {
	var d diag.Diagnostics
	consentCfg := resp.GetConfigNode().GetConsentConfig()
	setData(&d, data, consentPurposeKey, consentCfg.GetPurpose())
	setData(&d, data, consentApplicationIDKey, consentCfg.GetApplicationId())
	setData(&d, data, consentValidityPeriodKey, consentCfg.GetValidityPeriod())
	setData(&d, data, consentRevokeAfterUseKey, consentCfg.GetRevokeAfterUse())
	setData(&d, data, consentDataPointsKey, consentCfg.GetDataPoints())
	return d
}

func resourceConsentBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	cfg := &configpb.ConsentConfiguration{
		Purpose:        data.Get(consentPurposeKey).(string),
		ApplicationId:  data.Get(consentApplicationIDKey).(string),
		ValidityPeriod: uint64(data.Get(consentValidityPeriodKey).(int)),
		RevokeAfterUse: data.Get(consentRevokeAfterUseKey).(bool),
		DataPoints:     rawArrayToTypedArray[string](data.Get(consentDataPointsKey)),
	}
	builder.WithConsentConfig(cfg)
}
