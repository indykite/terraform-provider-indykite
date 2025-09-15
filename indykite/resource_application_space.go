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
	config "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
)

func resourceApplicationSpace() *schema.Resource {
	return &schema.Resource{
		Description:   "It is workspace or environment for your applications.  ",
		CreateContext: resAppSpaceCreateContext,
		ReadContext:   resAppSpaceReadContext,
		UpdateContext: resAppSpaceUpdateContext,
		DeleteContext: resAppSpaceDeleteContext,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: defaultTimeouts(),
		Schema: map[string]*schema.Schema{
			customerIDKey:         customerIDSchema(),
			nameKey:               nameSchema(),
			displayNameKey:        displayNameSchema(),
			descriptionKey:        descriptionSchema(),
			createTimeKey:         createTimeSchema(),
			updateTimeKey:         updateTimeSchema(),
			deletionProtectionKey: deletionProtectionSchema(),
			regionKey:             regionSchema(),
			ikgSizeKey:            ikgSizeSchema(),
			replicaRegionKey:      replicaRegionSchema(),
			dbConnectionKey:       dbConnectionSchema(),
		},
	}
}

func getDBConnection(data *schema.ResourceData) *config.DBConnection {
	dbConnRaw := data.Get(dbConnectionKey)
	if dbConnRaw == nil {
		return nil
	}

	dbConnList, ok := dbConnRaw.([]any)
	if !ok || len(dbConnList) == 0 || dbConnList[0] == nil {
		return nil
	}

	dbConnData, ok := dbConnList[0].(map[string]any)
	if !ok {
		return nil
	}

	// Extract values with safe type assertions
	url, _ := dbConnData[dbURLKey].(string)
	username, _ := dbConnData[dbUsernameKey].(string)
	password, _ := dbConnData[dbPasswordKey].(string)
	name, _ := dbConnData[dbNameKey].(string)

	// Only return a DBConnection if at least URL is provided
	if url == "" {
		return nil
	}

	return &config.DBConnection{
		Url:      url,
		Username: username,
		Password: password,
		Name:     name,
	}
}

func setDBConnectionData(d *diag.Diagnostics, data *schema.ResourceData, dbConn *config.DBConnection) {
	if dbConn == nil {
		setData(d, data, dbConnectionKey, []map[string]any{})
		return
	}

	dbConnData := []map[string]any{
		{
			dbURLKey:      dbConn.GetUrl(),
			dbUsernameKey: dbConn.GetUsername(),
			dbPasswordKey: dbConn.GetPassword(),
			dbNameKey:     dbConn.GetName(),
		},
	}
	setData(d, data, dbConnectionKey, dbConnData)
}

func resAppSpaceCreateContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	const maxWait = 20 * time.Minute
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, maxWait)
	defer cancel()

	name := data.Get(nameKey).(string)
	resp, err := clientCtx.GetClient().CreateApplicationSpace(ctx, &config.CreateApplicationSpaceRequest{
		CustomerId:    data.Get(customerIDKey).(string),
		Name:          name,
		DisplayName:   optionalString(data, displayNameKey),
		Description:   optionalString(data, descriptionKey),
		Region:        data.Get(regionKey).(string),
		IkgSize:       data.Get(ikgSizeKey).(string),
		ReplicaRegion: data.Get(replicaRegionKey).(string),
		DbConnection:  getDBConnection(data),
	})
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.GetId())
	return resAppSpaceReadAfterCreateContext(ctx, data, meta)
}

func getStatus(ctx context.Context, clientCtx *ClientContext, data *schema.ResourceData) (string, diag.Diagnostics) {
	var d diag.Diagnostics
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadApplicationSpace(ctx, &config.ReadApplicationSpaceRequest{
		Identifier: &config.ReadApplicationSpaceRequest_Id{
			Id: data.Id(),
		},
	})

	if err != nil {
		return "", diag.Diagnostics{buildPluginError("read application space failed")}
	}
	if resp == nil {
		return "", diag.Diagnostics{buildPluginError("read application space: empty response")}
	}
	if resp.GetAppSpace() == nil {
		return "", diag.Diagnostics{buildPluginError("read application space: missing AppSpace in response")}
	}

	data.SetId(resp.GetAppSpace().GetId())
	setData(&d, data, customerIDKey, resp.GetAppSpace().GetCustomerId())
	setData(&d, data, nameKey, resp.GetAppSpace().GetName())
	setData(&d, data, displayNameKey, resp.GetAppSpace().GetDisplayName())
	setData(&d, data, descriptionKey, resp.GetAppSpace().GetDescription())
	setData(&d, data, createTimeKey, resp.GetAppSpace().GetCreateTime())
	setData(&d, data, updateTimeKey, resp.GetAppSpace().GetUpdateTime())
	setData(&d, data, regionKey, resp.GetAppSpace().GetRegion())
	setData(&d, data, ikgSizeKey, resp.GetAppSpace().GetIkgSize())
	setData(&d, data, replicaRegionKey, resp.GetAppSpace().GetReplicaRegion())
	setDBConnectionData(&d, data, resp.GetAppSpace().GetDbConnection())

	s := resp.GetAppSpace().GetIkgStatus().String()
	return s, d
}

func waitForActive(ctx context.Context, clientCtx *ClientContext, data *schema.ResourceData) diag.Diagnostics {
	const (
		maxWait = 20 * time.Minute
		target  = config.AppSpaceIKGStatus_APP_SPACE_IKG_STATUS_STATUS_ACTIVE
	)
	intervals := []time.Duration{
		0, // immediate
		10 * time.Second,
		1 * time.Minute,
		2 * time.Minute,
	}
	ctx, cancel := context.WithTimeout(ctx, maxWait)
	defer cancel()

	deadline := time.Now().Add(maxWait)

	for i := 0; ; i++ {
		// compute next wait
		var wait time.Duration
		if i < len(intervals) {
			wait = intervals[i]
		} else {
			wait = 10 * time.Second
		}

		// if we already passed the deadline, stop
		if time.Now().Add(wait).After(deadline) {
			return diag.Diagnostics{buildPluginError("timed out waiting for IKG status to become active")}
		}

		// sleep before the check unless it's the immediate one
		if wait > 0 {
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return diag.Diagnostics{buildPluginError("timed out waiting for IKG status to become active")}
			}
		}

		// do the check
		status, d := getStatus(ctx, clientCtx, data)
		if d != nil {
			continue
		}
		if status == target.String() {
			return d
		}
	}
}

func resAppSpaceReadContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()
	resp, err := clientCtx.GetClient().ReadApplicationSpace(ctx, &config.ReadApplicationSpaceRequest{
		Identifier: &config.ReadApplicationSpaceRequest_Id{
			Id: data.Id(),
		},
	})
	if readHasFailed(&d, err, data) {
		return d
	}

	if resp.GetAppSpace() == nil {
		return diag.Diagnostics{buildPluginError("empty ApplicationSpace response")}
	}

	data.SetId(resp.GetAppSpace().GetId())
	setData(&d, data, customerIDKey, resp.GetAppSpace().GetCustomerId())
	setData(&d, data, nameKey, resp.GetAppSpace().GetName())
	setData(&d, data, displayNameKey, resp.GetAppSpace().GetDisplayName())
	setData(&d, data, descriptionKey, resp.GetAppSpace().GetDescription())
	setData(&d, data, createTimeKey, resp.GetAppSpace().GetCreateTime())
	setData(&d, data, updateTimeKey, resp.GetAppSpace().GetUpdateTime())
	setData(&d, data, regionKey, resp.GetAppSpace().GetRegion())
	setData(&d, data, ikgSizeKey, resp.GetAppSpace().GetIkgSize())
	setData(&d, data, replicaRegionKey, resp.GetAppSpace().GetReplicaRegion())
	setDBConnectionData(&d, data, resp.GetAppSpace().GetDbConnection())

	return d
}

func resAppSpaceReadAfterCreateContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	const maxWait = 20 * time.Minute
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, maxWait)
	defer cancel()

	waitForActive(ctx, clientCtx, data)

	return d
}

func updateDBConnection(data *schema.ResourceData) *config.DBConnection {
	if !data.HasChange(dbConnectionKey) {
		return nil
	}
	return getDBConnection(data)
}

func resAppSpaceUpdateContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutUpdate))
	defer cancel()

	// If only change in plan is delete_protection, just ignore the request
	if !data.HasChangeExcept(deletionProtectionKey) {
		return d
	}

	req := &config.UpdateApplicationSpaceRequest{
		Id:           data.Id(),
		DisplayName:  updateOptionalString(data, displayNameKey),
		Description:  updateOptionalString(data, descriptionKey),
		DbConnection: updateDBConnection(data),
	}

	_, err := clientCtx.GetClient().UpdateApplicationSpace(ctx, req)
	if HasFailed(&d, err) {
		return d
	}

	return resAppSpaceReadContext(ctx, data, meta)
}

func resAppSpaceDeleteContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutDelete))
	defer cancel()
	if hasDeleteProtection(&d, data) {
		return d
	}
	_, err := clientCtx.GetClient().DeleteApplicationSpace(ctx, &config.DeleteApplicationSpaceRequest{
		Id: data.Id(),
	})
	HasFailed(&d, err)

	return d
}
