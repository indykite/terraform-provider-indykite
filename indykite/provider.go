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
	"container/ring"
	"context"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/indykite/indykite-sdk-go/config"
	api "github.com/indykite/indykite-sdk-go/grpc"
	apicfg "github.com/indykite/indykite-sdk-go/grpc/config"
)

type (
	tfConfig struct {
		terraformVersion string
	}

	// ClientContext defines structure returned by ConfigureContextFunc,
	// which is passed into resources as meta arguemnt.
	ClientContext struct {
		configClient *config.Client
		config       *tfConfig
		bookmarks    *clientBookmarks
	}

	clientBookmarks struct {
		queueRing *ring.Ring
		count     int
		sync.Mutex
	}

	contextKey int
)

const (
	clientContextKey contextKey = 1

	// parallelism sets how many concurrent resource Terraform can process in the same time.
	// Default value is 10. If there is a way to get real value, use that instead of hardcoded value.
	// DO NOT SET it too high, as this amount of bookmarks will be send in requests.
	parallelism = 10
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	// Bookmarks must be set globally on Provider level.
	// Because ConfigureContextFunc is called multiple times per resource.
	bookmarks := &clientBookmarks{
		queueRing: ring.New(parallelism),
		count:     0,
	}

	// The actual provider
	provider := &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"indykite_customer":           dataSourceCustomer(),
			"indykite_application_space":  dataSourceAppSpace(),
			"indykite_application_spaces": dataSourceAppSpaceList(),
			"indykite_application":        dataSourceApplication(),
			"indykite_applications":       dataSourceApplicationList(),
			"indykite_application_agent":  dataSourceAppAgent(),
			"indykite_application_agents": dataSourceAppAgentList(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"indykite_application_space":            resourceApplicationSpace(),
			"indykite_application":                  resourceApplication(),
			"indykite_audit_sink":                   resourceAuditSink(),
			"indykite_application_agent":            resourceApplicationAgent(),
			"indykite_application_agent_credential": resourceApplicationAgentCredential(),
			"indykite_authorization_policy":         resourceAuthorizationPolicy(),
			"indykite_token_introspect":             resourceTokenIntrospect(),
			"indykite_ingest_pipeline":              resourceIngestPipeline(),
			"indykite_external_data_resolver":       resourceExternalDataResolver(),
			"indykite_consent":                      resourceConsent(),
			"indykite_entity_matching_pipeline":     resourceEntityMatchingPipeline(),
		},
	}

	provider.ConfigureContextFunc =
		func(ctx context.Context, _ *schema.ResourceData) (any, diag.Diagnostics) {
			return providerConfigure(ctx, bookmarks, provider.TerraformVersion)
		}

	return provider
}

func providerConfigure(
	ctx context.Context,
	bookmarks *clientBookmarks,
	version string,
) (any, diag.Diagnostics) {
	cfg := &tfConfig{terraformVersion: version}
	c, diags := cfg.getConfigClient(ctx) // Rename 'err' to 'diags' for clarity
	if diags.HasError() {
		return nil, diags
	}
	return &ClientContext{
		configClient: c,
		config:       cfg,
		bookmarks:    bookmarks,
	}, diags // Return diagnostics even if they contain only warnings
}

// getClientContext converts meta into ClientContext structure.
func getClientContext(d *diag.Diagnostics, meta any) *ClientContext {
	clientCtx, ok := meta.(*ClientContext)
	if !ok || clientCtx == nil {
		*d = append(*d, buildPluginError("Unable retrieve IndyKite client from meta"))
	}
	return clientCtx
}

// GetClient returns Config client, which exposes the whole config API.
func (x *ClientContext) GetClient() *config.Client {
	return x.configClient
}

// AddBookmarks adds new bookmarks to round queue.
// Calling repeatedly will add more and more bookmarks while remove old ones.
// Size of queue should reflect Terraform parallelism and API restrictions.
func (x *ClientContext) AddBookmarks(bookmarks ...string) {
	if len(bookmarks) == 0 {
		return
	}
	x.bookmarks.Lock()
	defer x.bookmarks.Unlock()
	for _, b := range bookmarks {
		if b == "" {
			continue
		}

		x.bookmarks.queueRing.Value = b
		x.bookmarks.queueRing = x.bookmarks.queueRing.Next()
		if x.bookmarks.count < parallelism {
			x.bookmarks.count++
		}
	}
}

// GetBookmarks returns all stored bookmarks in the round queue.
// Size of queue, and thus amount of bookmarks returned, should reflect Terraform parallelism and API restrictions.
func (x *ClientContext) GetBookmarks() []string {
	x.bookmarks.Lock()
	defer x.bookmarks.Unlock()

	b := make([]string, 0, x.bookmarks.count)
	x.bookmarks.queueRing.Do(func(a any) {
		if v, _ := a.(string); v != "" {
			b = append(b, v)
		}
	})

	return b
}

// WithClient stores the config client into the context.
func WithClient(ctx context.Context, c *config.Client) context.Context {
	return context.WithValue(ctx, clientContextKey, c)
}

// getConfigClient configures and returns a fully initialized getConfigClient.
func (*tfConfig) getConfigClient(ctx context.Context) (*config.Client, diag.Diagnostics) {
	if client, ok := ctx.Value(clientContextKey).(*config.Client); ok {
		return client, nil
	}

	// This can be called multiple times, because it is called from ConfigureContextFunc,
	// which is called for each resource.
	conn, err := config.NewClient(ctx,
		api.WithServiceAccount(), api.WithCredentialsLoader(apicfg.DefaultEnvironmentLoaderConfig))
	if err != nil {
		conn, err = config.NewClient(ctx,
			api.WithServiceAccount(), api.WithCredentialsLoader(apicfg.DefaultEnvironmentLoader))
		if err != nil {
			return nil, diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Unable to create IndyKite Config client",
				Detail:   err.Error(),
			}}
		}
		return conn, diag.Diagnostics{{
			Severity: diag.Warning,
			Summary: "Using deprecated environment variable for IndyKite Config client." +
				" Please use INDYKITE_SERVICE_ACCOUNT_CREDENTIALS instead.",
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
