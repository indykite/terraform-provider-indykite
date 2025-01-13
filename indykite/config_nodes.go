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
	"github.com/indykite/indykite-sdk-go/config"
	sdkerror "github.com/indykite/indykite-sdk-go/errors"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	"google.golang.org/grpc/codes"
)

type preBuildConfig func(
	d *diag.Diagnostics,
	data *schema.ResourceData,
	meta *ClientContext,
	builder *config.NodeRequest)
type postFlattenConfig func(data *schema.ResourceData, resp *configpb.ReadConfigNodeResponse) diag.Diagnostics

func configCreateContextFunc(configBuilder preBuildConfig, read schema.ReadContextFunc) schema.CreateContextFunc {
	return func(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
		var d diag.Diagnostics
		clientCtx := getClientContext(&d, meta)

		name := data.Get(nameKey).(string)
		builder, err := newCreateConfigBuilder(&d, name)
		if err != nil || clientCtx == nil {
			return d
		}

		if v, ok := data.GetOk(displayNameKey); ok {
			builder.WithDisplayName(v.(string))
		}
		if v, ok := data.GetOk(descriptionKey); ok {
			builder.WithDescription(v.(string))
		}

		if data.HasChanges(customerIDKey, appSpaceIDKey) {
			// This is error shouldn't happen as those fields should be marked as read-only in definition
			return append(d, buildPluginError(fmt.Sprintf(
				"properties %s and %s are readonly, use %s instead and report to us",
				customerIDKey, appSpaceIDKey, locationKey,
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
			configBuilder(&d, data, clientCtx, builder)
		}
		if d.HasError() {
			return d
		}

		resp, err := invokeCreateConfigNode(ctx, data, clientCtx, builder)
		if HasFailed(&d, err) {
			return d
		}
		data.SetId(resp.GetId())

		// Join Warning (errors are checked above) with optional errors/warnings from read callback
		return append(d, read(ctx, data, meta)...)
	}
}

func configReadContextFunc(flatten postFlattenConfig) schema.ReadContextFunc {
	return func(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
		var d diag.Diagnostics
		clientCtx := getClientContext(&d, meta)
		builder, err := config.NewRead(data.Id())
		if err != nil || clientCtx == nil {
			return d
		}

		resp, err := invokeReadConfigNode(ctx, data, clientCtx, builder)
		if readHasFailed(&d, err, data) {
			return d
		}
		if resp.GetConfigNode() == nil {
			return append(d, buildPluginError("config response is empty"))
		}
		data.SetId(resp.GetConfigNode().GetId())
		setData(&d, data, customerIDKey, resp.GetConfigNode().GetCustomerId())
		setData(&d, data, appSpaceIDKey, resp.GetConfigNode().GetAppSpaceId())

		switch {
		case resp.GetConfigNode().GetAppSpaceId() != "":
			setData(&d, data, locationKey, resp.GetConfigNode().GetAppSpaceId())
		case resp.GetConfigNode().GetCustomerId() != "":
			setData(&d, data, locationKey, resp.GetConfigNode().GetCustomerId())
		}

		setData(&d, data, nameKey, resp.GetConfigNode().GetName())
		setData(&d, data, displayNameKey, resp.GetConfigNode().GetDisplayName())
		setData(&d, data, descriptionKey, resp.GetConfigNode().GetDescription())
		setData(&d, data, createTimeKey, resp.GetConfigNode().GetCreateTime())
		setData(&d, data, updateTimeKey, resp.GetConfigNode().GetUpdateTime())

		// Post-Process
		if d.HasError() {
			return d
		}
		// Join Warning (errors are checked above) with optional errors/warnings from read callback
		return append(d, flatten(data, resp)...)
	}
}

func configUpdateContextFunc(configBuilder preBuildConfig, read schema.ReadContextFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
		var d diag.Diagnostics
		clientCtx := getClientContext(&d, meta)

		builder, err := config.NewUpdate(data.Id())
		if err != nil || clientCtx == nil {
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
			configBuilder(&d, data, clientCtx, builder)
		}
		if d.HasError() {
			return d
		}

		resp, err := invokeUpdateConfigNode(ctx, data, clientCtx, builder)
		if HasFailed(&d, err) {
			return d
		}
		data.SetId(resp.GetId())
		// Join Warning (errors are checked above) with optional errors/warnings from read callback
		return append(d, read(ctx, data, meta)...)
	}
}

func configDeleteContextFunc() schema.DeleteContextFunc {
	return func(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
		var d diag.Diagnostics
		clientCtx := getClientContext(&d, meta)

		builder, err := config.NewDelete(data.Id())
		if err != nil || clientCtx == nil {
			return d
		}

		if d.HasError() {
			return d
		}
		_, err = invokeDeleteConfigNode(ctx, data, clientCtx, builder)
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
			HasFailed(&d, err)
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

func invokeCreateConfigNode(
	parent context.Context,
	data *schema.ResourceData,
	clientCtx *ClientContext,
	b *config.NodeRequest,
) (*configpb.CreateConfigNodeResponse, error) {
	ctx, cancel := context.WithTimeout(parent, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	return clientCtx.GetClient().CreateConfigNode(ctx, b)
}

func invokeReadConfigNode(
	parent context.Context,
	data *schema.ResourceData,
	clientCtx *ClientContext,
	b *config.NodeRequest,
) (*configpb.ReadConfigNodeResponse, error) {
	ctx, cancel := context.WithTimeout(parent, data.Timeout(schema.TimeoutRead))
	defer cancel()

	return clientCtx.GetClient().ReadConfigNode(ctx, b)
}

func invokeUpdateConfigNode(
	parent context.Context,
	data *schema.ResourceData,
	clientCtx *ClientContext,
	b *config.NodeRequest,
) (*configpb.UpdateConfigNodeResponse, error) {
	ctx, cancel := context.WithTimeout(parent, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	return clientCtx.GetClient().UpdateConfigNode(ctx, b)
}

func invokeDeleteConfigNode(
	parent context.Context,
	data *schema.ResourceData,
	clientCtx *ClientContext,
	b *config.NodeRequest,
) (*configpb.DeleteConfigNodeResponse, error) {
	ctx, cancel := context.WithTimeout(parent, data.Timeout(schema.TimeoutDelete))
	defer cancel()

	return clientCtx.GetClient().DeleteConfigNode(ctx, b)
}
