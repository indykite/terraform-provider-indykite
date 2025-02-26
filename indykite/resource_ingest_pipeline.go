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
	"regexp"

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
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Deprecated:  "This field is deprecated and is not used anymore.",
				Description: "List of operations is no longer used. All previously saved values are ignored.",
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

	sourcesMapping := make([]string, len(ipCfg.GetSources()))
	for i, source := range ipCfg.GetSources() {
		sourcesMapping[i] = source
	}
	setData(&d, data, ingestPipelineSourcesTypeKey, sourcesMapping)

	return d
}

func resourceIngestPipelineBuild(
	_ *diag.Diagnostics,
	data *schema.ResourceData,
	_ *ClientContext,
	builder *config.NodeRequest,
) {
	cfg := &configpb.IngestPipelineConfig{
		Sources:       rawArrayToTypedArray[string](data.Get(ingestPipelineSourcesTypeKey).([]any)),
		AppAgentToken: data.Get(ingestPipelineAppAgentTokenTypeKey).(string),
	}

	builder.WithIngestPipelineConfig(cfg)
}
