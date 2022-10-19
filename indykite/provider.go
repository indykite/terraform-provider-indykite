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
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/indykite/jarvis-sdk-go/config"
	api "github.com/indykite/jarvis-sdk-go/grpc"
	apicfg "github.com/indykite/jarvis-sdk-go/grpc/config"
)

type (
	tfConfig struct {
		terraformVersion string
	}

	metaContext struct {
		client *config.Client
		config *tfConfig
	}

	contextKey int
)

const (
	// ClientContext for unit testing to pass *config.Client along.
	ClientContext contextKey = 1
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	// The actual provider
	provider := &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"indykite_customer":           dataSourceCustomer(),
			"indykite_application_space":  dataSourceAppSpace(),
			"indykite_application_spaces": dataSourceAppSpaceList(),
			"indykite_application":        dataSourceApplication(),
			"indykite_applications":       dataSourceApplicationList(),
			"indykite_tenant":             dataSourceTenant(),
			"indykite_tenants":            dataSourceTenantList(),
			"indykite_application_agent":  dataSourceAppAgent(),
			"indykite_application_agents": dataSourceAppAgentList(),
			"indykite_oauth2_provider":    dataSourceOAuth2Provider(),
			"indykite_oauth2_application": dataSourceOAuth2Application(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"indykite_application_space":            resourceApplicationSpace(),
			"indykite_application":                  resourceApplication(),
			"indykite_tenant":                       resourceTenant(),
			"indykite_application_agent":            resourceApplicationAgent(),
			"indykite_application_agent_credential": resourceApplicationAgentCredential(),
			"indykite_auth_flow":                    resourceAuthFlow(),
			"indykite_authorization_policy":         resourceAuthorizationPolicy(),
			"indykite_email_notification":           resourceEmailNotification(),
			"indykite_ingest_mapping":               resourceIngestMapping(),
			"indykite_oauth2_client":                resourceOAuth2Client(),
			"indykite_oauth2_provider":              resourceOAuth2Provider(),
			"indykite_oauth2_application":           resourceOAuth2Application(),
		},
	}

	provider.ConfigureContextFunc =
		func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return providerConfigure(ctx, data, provider.TerraformVersion)
		}

	return provider
}

func providerConfigure(ctx context.Context, _ *schema.ResourceData, version string) (interface{}, diag.Diagnostics) {
	cfg := &tfConfig{terraformVersion: version}
	c, err := cfg.getClient(ctx)
	if err.HasError() {
		return nil, err
	}
	return &metaContext{client: c, config: cfg}, nil
}

func fromMeta(d *diag.Diagnostics, meta interface{}) *metaContext {
	client, ok := meta.(*metaContext)
	if !ok || client == nil {
		*d = append(*d, buildPluginError("Unable retrieve IndyKite client from meta"))
	}
	return client
}

func (x *metaContext) getClient() *config.Client {
	return x.client
}

// getClient configures and returns a fully initialized getClient.
func (c *tfConfig) getClient(ctx context.Context) (*config.Client, diag.Diagnostics) {
	if client, ok := ctx.Value(ClientContext).(*config.Client); ok {
		return client, nil
	}

	conn, err := config.NewClient(ctx,
		api.WithServiceAccount(), api.WithCredentialsLoader(apicfg.DefaultEnvironmentLoader))
	if err != nil {
		return nil, diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Unable to create IndyKite client",
			Detail:   err.Error(),
		}}
	}
	return conn, nil
}

func defaultDataTimeouts() *schema.ResourceTimeout {
	return defaultTimeouts("create", "update", "delete")
}

func defaultTimeouts(exclude ...string) *schema.ResourceTimeout {
	rs := &schema.ResourceTimeout{
		Default: schema.DefaultTimeout(4 * time.Minute),
	}
	if !contains(exclude, "create") {
		rs.Create = schema.DefaultTimeout(4 * time.Minute)
	}
	if !contains(exclude, "read") {
		rs.Read = schema.DefaultTimeout(4 * time.Minute)
	}
	if !contains(exclude, "update") {
		rs.Update = schema.DefaultTimeout(4 * time.Minute)
	}
	if !contains(exclude, "delete") {
		rs.Delete = schema.DefaultTimeout(4 * time.Minute)
	}
	return rs
}
