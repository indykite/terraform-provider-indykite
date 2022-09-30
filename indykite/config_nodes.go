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
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/indykite/jarvis-sdk-go/config"
	sdkerror "github.com/indykite/jarvis-sdk-go/errors"
	configpb "github.com/indykite/jarvis-sdk-go/gen/indykite/config/v1beta1"
	"google.golang.org/grpc/codes"
)

type preBuildConfig func(
	d *diag.Diagnostics,
	data *schema.ResourceData,
	meta *metaContext,
	builder *config.NodeRequest)
type postFlattenConfig func(data *schema.ResourceData, resp *configpb.ReadConfigNodeResponse) diag.Diagnostics

func configCreateContextFunc(configBuilder preBuildConfig, read schema.ReadContextFunc) schema.CreateContextFunc {
	return func(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
		client := fromMeta(&d, meta)

		name := data.Get(nameKey).(string)
		builder, err := newCreateConfigBuilder(&d, name)
		if err != nil || client == nil {
			return d
		}

		if v, ok := data.GetOk(displayNameKey); ok {
			builder.WithDisplayName(v.(string))
		}
		if v, ok := data.GetOk(descriptionKey); ok {
			builder.WithDescription(v.(string))
		}

		if data.HasChanges(customerIDKey, appSpaceIDKey, tenantIDKey) {
			// This is error shouldn't happen as those fields should be marked as read-only in definition
			return append(d, buildPluginError(fmt.Sprintf(
				"properties %s, %s and %s are readonly, use %s instead and report to us",
				customerIDKey, appSpaceIDKey, tenantIDKey, locationKey,
			)))
		}

		loc, ok := data.GetOk(locationKey)
		if !ok {
			// This is error shouldn't happen as those fields should be marked as required in definition
			return append(d, buildPluginErrorWithAttrName("location is required for creation", locationKey))
		}
		builder.ForLocation(loc.(string))

		// Pre-Process
		if configBuilder != nil {
			configBuilder(&d, data, client, builder)
		}
		if d.HasError() {
			return d
		}

		resp, err := invokeCreateConfigNode(ctx, data, client.getClient(), builder)
		if hasFailed(&d, err) {
			return d
		}
		data.SetId(resp.Id)

		return read(ctx, data, meta)
	}
}

func configReadContextFunc(flatten postFlattenConfig) schema.ReadContextFunc {
	return func(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
		client := fromMeta(&d, meta)

		builder, err := config.NewRead(data.Id())
		if err != nil || client == nil {
			return d
		}

		resp, err := invokeReadConfigNode(ctx, data, client.getClient(), builder)
		if hasFailed(&d, err) {
			return d
		}
		if resp.ConfigNode == nil {
			return append(d, buildPluginError("config response is empty"))
		}
		data.SetId(resp.ConfigNode.Id)
		setData(&d, data, customerIDKey, resp.ConfigNode.CustomerId)
		setData(&d, data, appSpaceIDKey, resp.ConfigNode.AppSpaceId)
		setData(&d, data, tenantIDKey, resp.ConfigNode.TenantId)

		switch {
		case resp.ConfigNode.TenantId != "":
			setData(&d, data, locationKey, resp.ConfigNode.TenantId)
		case resp.ConfigNode.AppSpaceId != "":
			setData(&d, data, locationKey, resp.ConfigNode.AppSpaceId)
		case resp.ConfigNode.CustomerId != "":
			setData(&d, data, locationKey, resp.ConfigNode.CustomerId)
		}

		setData(&d, data, nameKey, resp.ConfigNode.Name)
		setData(&d, data, displayNameKey, resp.ConfigNode.DisplayName)
		setData(&d, data, descriptionKey, resp.ConfigNode.Description)
		setData(&d, data, createTimeKey, resp.ConfigNode.CreateTime)
		setData(&d, data, updateTimeKey, resp.ConfigNode.UpdateTime)

		// Post-Process
		return flatten(data, resp)
	}
}

func configUpdateContextFunc(configBuilder preBuildConfig, read schema.ReadContextFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
		client := fromMeta(&d, meta)

		builder, err := config.NewUpdate(data.Id())
		if err != nil || client == nil {
			return d
		}

		if data.HasChange(displayNameKey) {
			builder.WithDisplayName(data.Get(displayNameKey).(string))
		}
		if data.HasChange(descriptionKey) {
			builder.WithDescription(data.Get(descriptionKey).(string))
		}

		// Pre-Process
		if configBuilder != nil {
			configBuilder(&d, data, client, builder)
		}
		if d.HasError() {
			return d
		}

		resp, err := invokeUpdateConfigNode(ctx, data, client.getClient(), builder)
		if hasFailed(&d, err) {
			return d
		}
		data.SetId(resp.Id)
		return read(ctx, data, meta)
	}
}

func configDeleteContextFunc() schema.DeleteContextFunc {
	return func(ctx context.Context, data *schema.ResourceData, meta interface{}) (d diag.Diagnostics) {
		client := fromMeta(&d, meta)

		builder, err := config.NewDelete(data.Id())
		if err != nil || client == nil {
			return d
		}

		if d.HasError() {
			return d
		}
		_, err = invokeDeleteConfigNode(ctx, data, client.getClient(), builder)
		if err != nil {
			var er *sdkerror.StatusError
			if errors.As(err, &er) {
				if er.Code() == codes.NotFound {
					log.Print("[WARN] Removing failed, because it's gone")
					// The resource doesn't exist anymore
					data.SetId("")
					return nil
				}
			}
			hasFailed(&d, err)
		}
		return d
	}
}

func newCreateConfigBuilder(d *diag.Diagnostics, name string) (*config.NodeRequest, error) {
	b, err := config.NewCreate(name)
	if err != nil {
		*d = append(*d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  err.Error(),
			Detail:   "Value can have lowercase letters, digits, or hyphens. It must start with a lowercase letter and end with a letter or number.",
		})
		return nil, err
	}
	return b, nil
}

func invokeCreateConfigNode(parent context.Context, data *schema.ResourceData,
	client *config.Client, b *config.NodeRequest) (*configpb.CreateConfigNodeResponse, error) {
	ctx, cancel := context.WithTimeout(parent, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	resp, err := client.CreateConfigNode(ctx, b)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func invokeReadConfigNode(parent context.Context, data *schema.ResourceData,
	client *config.Client, b *config.NodeRequest) (*configpb.ReadConfigNodeResponse, error) {
	ctx, cancel := context.WithTimeout(parent, data.Timeout(schema.TimeoutRead))
	defer cancel()

	resp, err := client.ReadConfigNode(ctx, b)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func invokeUpdateConfigNode(parent context.Context, data *schema.ResourceData,
	client *config.Client, b *config.NodeRequest) (*configpb.UpdateConfigNodeResponse, error) {
	ctx, cancel := context.WithTimeout(parent, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	resp, err := client.UpdateConfigNode(ctx, b)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func invokeDeleteConfigNode(parent context.Context, data *schema.ResourceData,
	client *config.Client, b *config.NodeRequest) (*configpb.DeleteConfigNodeResponse, error) {
	ctx, cancel := context.WithTimeout(parent, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	resp, err := client.DeleteConfigNode(ctx, b)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
