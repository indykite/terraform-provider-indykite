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
	"math"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	trustScoreProfileNodeClassification = "node_classification"
	trustScoreProfileDimensionsKey      = "dimension"
	trustScoreProfileSchedule           = "schedule"
	trustScoreProfileName               = "name"
	trustScoreProfileWeight             = "weight"
)

func resourceTrustScoreProfile() *schema.Resource {
	return &schema.Resource{
		Description: "The Trust Score Profile helps assess how trustworthy data is. " +
			"It allows applications, authorization policies, and AI systems to define and check  " +
			"whether data meets specific reliability requirements. " +
			"By validating key factors — such as how recent, complete, and accurate the data is —  " +
			"the Trust Score ensures that only high-quality and reliable data is used in decision-making. " +
			"This reduces risk and improves the overall quality of downstream processes. ",

		CreateContext: resTrustScoreProfileCreate,
		ReadContext:   resTrustScoreProfileRead,
		UpdateContext: resTrustScoreProfileUpdate,
		DeleteContext: resTrustScoreProfileDelete,
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
							Type:     schema.TypeString,
							Required: true,
							Description: "Name of the trust score dimensions. Possible values are: `" +
								strings.Join(getMapStringKeys(TrustScoreDimensionNames), "`, `") + "`.  " +
								"`Origin`: Identifies where the data comes from, " +
								"ensuring its source is transparent and trustworthy.  " +
								"`Validity`: Checks whether the data is in the correct format " +
								"and follows expected rules.  " +
								"`Completeness`: Confirms that no critical information is missing from the data.  " +
								"`Freshness`: Measures how up-to-date the data is to ensure it's still relevant.  " +
								"`Verification`: Ensures the data has been reviewed and confirmed " +
								"as accurate by a trusted source.",
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
				Type: schema.TypeString,
				Description: "Schedule sets the time between re-calculations. Possible values are: `" +
					strings.Join(getMapStringKeys(TrustScoreProfileScheduleFrequencies), "`, `") + "`.",
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice(getMapStringKeys(
						TrustScoreProfileScheduleFrequencies), false),
				},
			},
		},
	}
}

func resTrustScoreProfileCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	// Map schedule from Terraform format to API format
	scheduleValue := data.Get(trustScoreProfileSchedule).(string)
	apiSchedule := TrustScoreProfileScheduleToAPI[scheduleValue]

	req := CreateTrustScoreProfileRequest{
		ProjectID:          data.Get(locationKey).(string),
		Name:               data.Get(nameKey).(string),
		DisplayName:        stringValue(optionalString(data, displayNameKey)),
		Description:        stringValue(optionalString(data, descriptionKey)),
		NodeClassification: data.Get(trustScoreProfileNodeClassification).(string),
		Dimensions:         buildDimensions(data),
		Schedule:           apiSchedule,
	}

	var resp TrustScoreProfileResponse
	err := clientCtx.GetClient().Post(ctx, "/trust-score-profiles", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

	return resTrustScoreProfileRead(ctx, data, meta)
}

func resTrustScoreProfileRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp TrustScoreProfileResponse
	// Support both ID and name?location=parent_id formats
	path := buildReadPath("/trust-score-profiles", data)
	err := clientCtx.GetClient().Get(ctx, path, &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, appSpaceIDKey, resp.AppSpaceID)

	// Set location based on which is present
	if resp.AppSpaceID != "" {
		setData(&d, data, locationKey, resp.AppSpaceID)
	} else if resp.CustomerID != "" {
		setData(&d, data, locationKey, resp.CustomerID)
	}

	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, trustScoreProfileNodeClassification, resp.NodeClassification)

	// Convert dimensions
	var ratio float64 = 10000 // 4 decimal places to round number when converting float32 to float64
	dimensions := make([]any, len(resp.Dimensions))
	for i, dim := range resp.Dimensions {
		// Map dimension name from API format to Terraform format
		terraformName := TrustScoreDimensionFromAPI[dim.Name]
		if terraformName == "" {
			terraformName = dim.Name // Fallback to original value if not found
		}
		dimensions[i] = map[string]any{
			trustScoreProfileName:   terraformName,
			trustScoreProfileWeight: math.Round(float64(dim.Weight)*ratio) / ratio,
		}
	}
	setData(&d, data, trustScoreProfileDimensionsKey, dimensions)

	// Map schedule from API format to Terraform format
	terraformSchedule := TrustScoreProfileScheduleFromAPI[resp.Schedule]
	if terraformSchedule == "" {
		terraformSchedule = resp.Schedule // Fallback to original value if not found
	}
	setData(&d, data, trustScoreProfileSchedule, terraformSchedule)

	return d
}

func resTrustScoreProfileUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := UpdateTrustScoreProfileRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	if data.HasChange(trustScoreProfileDimensionsKey) {
		req.Dimensions = buildDimensions(data)
	}

	if data.HasChange(trustScoreProfileSchedule) {
		scheduleValue := data.Get(trustScoreProfileSchedule).(string)
		apiSchedule := TrustScoreProfileScheduleToAPI[scheduleValue]
		req.Schedule = &apiSchedule
	}

	var resp TrustScoreProfileResponse
	err := clientCtx.GetClient().Put(ctx, "/trust-score-profiles/"+data.Id(), req, &resp)
	if HasFailed(&d, err) {
		return d
	}

	return resTrustScoreProfileRead(ctx, data, meta)
}

func resTrustScoreProfileDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	err := clientCtx.GetClient().Delete(ctx, "/trust-score-profiles/"+data.Id())
	HasFailed(&d, err)
	return d
}

// buildDimensions converts Terraform schema dimensions to REST API format.
func buildDimensions(data *schema.ResourceData) []*TrustScoreDimension {
	dimensionsSet := data.Get(trustScoreProfileDimensionsKey).([]any)
	dimensions := make([]*TrustScoreDimension, len(dimensionsSet))
	for i, o := range dimensionsSet {
		item, ok := o.(map[string]any)
		if !ok {
			continue
		}
		// Map dimension name from Terraform format to API format
		terraformName := item[trustScoreProfileName].(string)
		apiName := TrustScoreDimensionToAPI[terraformName]
		if apiName == "" {
			apiName = terraformName // Fallback to original value if not found
		}
		dimensions[i] = &TrustScoreDimension{
			Name:   apiName,
			Weight: float32(item[trustScoreProfileWeight].(float64)),
		}
	}
	return dimensions
}
