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
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	knowledgeQueryJSONQueryConfigKey = "query"
	knowledgeQueryStatusKey          = "status"
	knowledgeQueryPolicyID           = "policy_id"
)

func resourceKnowledgeQuery() *schema.Resource {
	return &schema.Resource{
		Description: "**Creating Policy:**  <br>" +
			"An authorization admin starts by creating a new subgraph or " +
			"selecting an existing one as the container for the policy they wish to create. " +
			"Next, the admin specifies a set of nodes and relationships within the subgraph and  " +
			"specifies the static filters and partial filters on the selected nodes and relationship.  " +
			"There must be exactly one node that is specified as the Subject node.  " +
			"However, two separate policies may contain two different Subject nodes.  " +
			"Note that not every node and relationship needs a filter or partial filter.  " +
			"The nodes and relationships, along with the filters and partial filters,  " +
			"form the necessary requirements for the queries that will be defined " +
			"in the context of this policy.  <br>" +
			"**Creating Query:**  <br>" +
			"Every query is created in the context of a policy. " +
			" While the policy describes the requirements, the query focuses on retrieving data.  " +
			"The policy admin starts by selecting a subgraph and a policy for the context of the query.  " +
			"The admin then specifies the read, upsert, and delete components for the query.  " +
			"When the admin is done specifying the query, the query combined with the policy " +
			"are translated to Cypher.  ",
		CreateContext: resKnowledgeQueryCreate,
		ReadContext:   resKnowledgeQueryRead,
		UpdateContext: resKnowledgeQueryUpdate,
		DeleteContext: resKnowledgeQueryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},

		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			locationKey:    locationSchema(),
			customerIDKey:  setComputed(customerIDSchema()),
			appSpaceIDKey:  setComputed(appSpaceIDSchema()),
			nameKey:        nameSchema(),
			displayNameKey: displayNameSchema(),
			descriptionKey: descriptionSchema(),
			createTimeKey:  createTimeSchema(),
			updateTimeKey:  updateTimeSchema(),

			knowledgeQueryJSONQueryConfigKey: {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
					validation.StringIsJSON,
				),
				Description: "Configuration of Knowledge Query in JSON format, the same one exported by The Hub.",
			},
			knowledgeQueryStatusKey: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(getMapStringKeys(KnowledgeQueryStatusTypes), false),
				Description: "Status of the Knowledge Query. Possible values are: " +
					strings.Join(getMapStringKeys(KnowledgeQueryStatusTypes), ", ") + ".",
			},
			knowledgeQueryPolicyID: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: ValidateGID,
				Description:      "ID of the Authorization Policy that is used to authorize the query.",
			},
		},
	}
}

func resKnowledgeQueryCreate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	// Map status from Terraform format to API format
	statusValue := data.Get(knowledgeQueryStatusKey).(string)
	apiStatus := KnowledgeQueryStatusToAPI[statusValue]

	req := CreateKnowledgeQueryRequest{
		ProjectID:   data.Get(locationKey).(string),
		Name:        data.Get(nameKey).(string),
		DisplayName: stringValue(optionalString(data, displayNameKey)),
		Description: stringValue(optionalString(data, descriptionKey)),
		Query:       data.Get(knowledgeQueryJSONQueryConfigKey).(string),
		Status:      apiStatus,
		PolicyID:    data.Get(knowledgeQueryPolicyID).(string),
	}

	var resp KnowledgeQueryResponse
	err := clientCtx.GetClient().Post(ctx, "/knowledge-queries", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)

	return resKnowledgeQueryRead(ctx, data, meta)
}

func resKnowledgeQueryRead(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp KnowledgeQueryResponse
	// Support both ID and name?location=parent_id formats
	path := buildReadPath("/knowledge-queries", data)
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
	setData(&d, data, knowledgeQueryJSONQueryConfigKey, resp.Query)

	// Map status from API format to Terraform format
	terraformStatus := KnowledgeQueryStatusFromAPI[resp.Status]
	if terraformStatus == "" {
		terraformStatus = resp.Status // Fallback to original value if not found
	}
	setData(&d, data, knowledgeQueryStatusKey, terraformStatus)
	setData(&d, data, knowledgeQueryPolicyID, resp.PolicyID)

	return d
}

func resKnowledgeQueryUpdate(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := UpdateKnowledgeQueryRequest{
		DisplayName: updateOptionalString(data, displayNameKey),
		Description: updateOptionalString(data, descriptionKey),
	}

	if data.HasChange(knowledgeQueryJSONQueryConfigKey) {
		query := data.Get(knowledgeQueryJSONQueryConfigKey).(string)
		req.Query = &query
	}

	if data.HasChange(knowledgeQueryStatusKey) {
		statusValue := data.Get(knowledgeQueryStatusKey).(string)
		apiStatus := KnowledgeQueryStatusToAPI[statusValue]
		req.Status = &apiStatus
	}

	if data.HasChange(knowledgeQueryPolicyID) {
		policyID := data.Get(knowledgeQueryPolicyID).(string)
		req.PolicyID = &policyID
	}

	var resp KnowledgeQueryResponse
	err := clientCtx.GetClient().Put(ctx, "/knowledge-queries/"+data.Id(), req, &resp)
	if HasFailed(&d, err) {
		return d
	}

	return resKnowledgeQueryRead(ctx, data, meta)
}

func resKnowledgeQueryDelete(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	err := clientCtx.GetClient().Delete(ctx, "/knowledge-queries/"+data.Id())
	HasFailed(&d, err)
	return d
}
