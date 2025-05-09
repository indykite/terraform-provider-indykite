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
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

const (
	knowledgeQueryJSONQueryConfigKey = "query"
	knowledgeQueryStatusKey          = "status"
	knowledgeQueryPolicyID           = "policy_id"
)

func resourceKnowledgeQuery() *schema.Resource {
	readContext := configReadContextFunc(resourceKnowledgeQueryFlatten)

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
		CreateContext: configCreateContextFunc(resourceKnowledgeQueryBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceKnowledgeQueryBuild, readContext),
		DeleteContext: configDeleteContextFunc(),
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

func resourceKnowledgeQueryFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) diag.Diagnostics {
	var d diag.Diagnostics
	query := resp.GetConfigNode().GetKnowledgeQueryConfig().GetQuery()
	if query == "" {
		return append(d, buildPluginError("config in the response is not valid KnowledgeQueryConfig"))
	}
	setData(&d, data, knowledgeQueryJSONQueryConfigKey, query)

	status := resp.GetConfigNode().GetKnowledgeQueryConfig().GetStatus()
	if status == configpb.KnowledgeQueryConfig_STATUS_INVALID {
		return append(d, buildPluginError("status in the response is not valid"))
	}
	statusKey, exist := ReverseProtoEnumMap(KnowledgeQueryStatusTypes)[status]
	if !exist {
		d = append(d, buildPluginError("unsupported Policy Status Type: "+status.String()))
	}
	setData(&d, data, knowledgeQueryStatusKey, statusKey)

	setData(&d, data, knowledgeQueryPolicyID, resp.GetConfigNode().GetKnowledgeQueryConfig().GetPolicyId())

	return d
}

func knowledgeQueryConfigBuilder(data *schema.ResourceData) *configpb.KnowledgeQueryConfig {
	cfg := &configpb.KnowledgeQueryConfig{
		Query:    data.Get(knowledgeQueryJSONQueryConfigKey).(string),
		Status:   KnowledgeQueryStatusTypes[data.Get(knowledgeQueryStatusKey).(string)],
		PolicyId: data.Get(knowledgeQueryPolicyID).(string),
	}
	return cfg
}

func resourceKnowledgeQueryBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	cfg := knowledgeQueryConfigBuilder(data)
	builder.WithKnowledgeQueryConfig(cfg)
}
