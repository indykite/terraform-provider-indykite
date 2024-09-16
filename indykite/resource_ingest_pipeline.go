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
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/indykite/indykite-sdk-go/config"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

//nolint:gosec // there are no secrets
const (
	ingestPipelineSourcesTypeKey       = "sources"
	ingestPipelineOperationsTypeKey    = "operations"
	ingestPipelineAppAgentTokenTypeKey = "app_agent_token"
)

var (
	ingestPipelineAppAgentTokenRegex = regexp.MustCompile(`^[A-Za-z0-9-_]+?\.[A-Za-z0-9-_]+?\.[A-Za-z0-9-_]+?$`)
	ingestPipelineAllowedOperations  = []string{
		"OPERATION_UPSERT_NODE",
		"OPERATION_UPSERT_RELATIONSHIP",
		"OPERATION_DELETE_NODE",
		"OPERATION_DELETE_RELATIONSHIP",
		"OPERATION_DELETE_NODE_PROPERTY",
		"OPERATION_DELETE_RELATIONSHIP_PROPERTY",
	}
	//nolint:lll // long description
	ingestPipelineOperationsDescription = "List of operations which will be allowed to be used in the ingest pipeline. Valid values are: \n  - \"" +
		strings.Join(ingestPipelineAllowedOperations, "\" \n  - \"") + "\"\n"
)

func resourceIngestPipeline() *schema.Resource {
	readContext := configReadContextFunc(resourceIngestPipelineFlatten)

	return &schema.Resource{
		Description: `Ingest pipeline configuration adds support for 3rd party data sources, which can be used to ingest data to the IndyKite system. The configuration also allows you to define the allowed operations in the pipeline.`,

		CreateContext: configCreateContextFunc(resourceIngestPipelineBuild, readContext),
		ReadContext:   readContext,
		UpdateContext: configUpdateContextFunc(resourceIngestPipelineBuild, readContext),
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

			ingestPipelineSourcesTypeKey: {
				Type:        schema.TypeList,
				Description: "List of sources to be used in the ingest pipeline.",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
				MaxItems: 10,
			},
			ingestPipelineOperationsTypeKey: {
				Type:        schema.TypeList,
				Description: ingestPipelineOperationsDescription,
				Required:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ingestPipelineAllowedOperations, false),
				},
				MinItems: 1,
				MaxItems: 6,
			},
			ingestPipelineAppAgentTokenTypeKey: {
				Type:        schema.TypeString,
				Description: "Application agent token is used to identify the application space in IndyKite APIs.",
				Sensitive:   true,
				Required:    true,
				ValidateFunc: validation.StringMatch(
					ingestPipelineAppAgentTokenRegex, "must be valid application agent token.",
				),
			},
		},
	}
}

func resourceIngestPipelineFlatten(
	data *schema.ResourceData,
	resp *configpb.ReadConfigNodeResponse,
) diag.Diagnostics {
	var d diag.Diagnostics
	ipCfg := resp.GetConfigNode().GetIngestPipelineConfig()

	operationsMapping := make([]string, len(ipCfg.GetOperations()))
	for i, intOp := range ipCfg.GetOperations() {
		strOp, ok := IngestPipelineOperationTypesReverse[intOp]
		if !ok {
			d = append(d, buildPluginError(fmt.Sprintf("unsupported Ingest Pipeline Operation: %d", intOp)))
			return d
		}
		operationsMapping[i] = strOp
	}
	setData(&d, data, ingestPipelineOperationsTypeKey, operationsMapping)

	sourcesMapping := make([]string, len(ipCfg.GetSources()))
	for i, source := range ipCfg.GetSources() {
		sourcesMapping[i] = source
	}
	setData(&d, data, ingestPipelineSourcesTypeKey, sourcesMapping)

	return d
}

func resourceIngestPipelineBuild(
	d *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	sources := data.Get(ingestPipelineSourcesTypeKey).([]any)
	operations := data.Get(ingestPipelineOperationsTypeKey).([]any)
	cfg := &configpb.IngestPipelineConfig{
		Sources:       rawArrayToTypedArray[string](sources),
		AppAgentToken: data.Get(ingestPipelineAppAgentTokenTypeKey).(string),
	}

	var cfgOperations = make([]configpb.IngestPipelineOperation, len(operations))
	for i, o := range operations {
		cfgOperation, ok := IngestPipelineOperationTypes[o.(string)]
		if !ok {
			*d = append(*d, buildPluginError(fmt.Sprintf("unsupported Ingest Pipeline Operation: %s", o)))
		}
		cfgOperations[i] = cfgOperation
	}
	cfg.Operations = cfgOperations

	builder.WithIngestPipelineConfig(cfg)
}
