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
)

func resourceApplicationSpace() *schema.Resource {
	orgIdentifiers := []string{customerIDKey, organizationIDKey}

	return &schema.Resource{
		Description:   "It is workspace or environment for your applications.  ",
		CreateContext: resAppSpaceCreateContext,
		ReadContext:   resAppSpaceReadContext,
		UpdateContext: resAppSpaceUpdateContext,
		DeleteContext: resAppSpaceDeleteContext,
		Importer: &schema.ResourceImporter{
			StateContext: basicStateImporter,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(20 * time.Minute),
			Create:  schema.DefaultTimeout(20 * time.Minute),
			Read:    schema.DefaultTimeout(4 * time.Minute),
			Update:  schema.DefaultTimeout(4 * time.Minute),
			Delete:  schema.DefaultTimeout(4 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			customerIDKey:         setExactlyOneOf(customerIDSchema(), customerIDKey, orgIdentifiers),
			organizationIDKey:     setExactlyOneOf(organizationIDSchema(), organizationIDKey, orgIdentifiers),
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

func getDBConnection(data *schema.ResourceData) *DBConnection {
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

	return &DBConnection{
		URL:      url,
		Username: username,
		Password: password,
		Name:     name,
	}
}

func setDBConnectionData(d *diag.Diagnostics, data *schema.ResourceData, dbConn *DBConnection) {
	if dbConn == nil {
		setData(d, data, dbConnectionKey, []map[string]any{})
		return
	}

	oldDBConn := getDBConnection(data)
	oldPassword := ""
	if oldDBConn != nil {
		oldPassword = oldDBConn.Password
	}
	dbConnData := []map[string]any{
		{
			dbURLKey:      dbConn.URL,
			dbUsernameKey: dbConn.Username,
			dbPasswordKey: oldPassword,
			dbNameKey:     dbConn.Name,
		},
	}
	setData(d, data, dbConnectionKey, dbConnData)
}

func resAppSpaceCreateContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	// Get organization ID from either organization_id or customer_id (backward compatibility)
	orgID := data.Get(organizationIDKey).(string)
	if orgID == "" {
		orgID = data.Get(customerIDKey).(string)
	}

	req := CreateApplicationSpaceRequest{
		OrganizationID: orgID,
		Name:           data.Get(nameKey).(string),
		DisplayName:    stringValue(optionalString(data, displayNameKey)),
		Description:    stringValue(optionalString(data, descriptionKey)),
		Region:         data.Get(regionKey).(string),
		IKGSize:        data.Get(ikgSizeKey).(string),
		ReplicaRegion:  data.Get(replicaRegionKey).(string),
		DBConnection:   getDBConnection(data),
	}

	var resp ApplicationSpaceResponse
	err := clientCtx.GetClient().Post(ctx, "/projects", req, &resp)
	if HasFailed(&d, err) {
		return d
	}
	data.SetId(resp.ID)
	return resAppSpaceReadAfterCreateContext(ctx, data, meta)
}

func getStatus(ctx context.Context, clientCtx *ClientContext, data *schema.ResourceData) (string, diag.Diagnostics) {
	var d diag.Diagnostics
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutRead))
	defer cancel()

	var resp ApplicationSpaceResponse
	err := clientCtx.GetClient().Get(ctx, "/projects/"+data.Id(), &resp)

	if err != nil {
		// If we get a 404, the resource might not be ready yet, return empty status to retry
		if IsNotFoundError(err) {
			return "", d // Return empty status without error to trigger retry
		}
		return "", diag.Diagnostics{buildPluginError("read application space failed")}
	}

	data.SetId(resp.ID)
	setData(&d, data, customerIDKey, resp.CustomerID)
	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, regionKey, resp.Region)
	setData(&d, data, ikgSizeKey, resp.IKGSize)
	setData(&d, data, replicaRegionKey, resp.ReplicaRegion)
	setDBConnectionData(&d, data, resp.DBConnection)

	return resp.IKGStatus, d
}

func waitForActive(ctx context.Context, clientCtx *ClientContext, data *schema.ResourceData) diag.Diagnostics {
	// Accept multiple possible status values for active state
	activeStatuses := map[string]bool{
		"APP_SPACE_IKG_STATUS_STATUS_ACTIVE": true,
		"ACTIVE":                             true,
		"":                                   true, // Empty status might mean already active
	}

	intervals := []time.Duration{
		0, // immediate
		10 * time.Second,
		1 * time.Minute,
		2 * time.Minute,
	}

	for i := 0; ; i++ {
		// compute next wait
		var wait time.Duration
		if i < len(intervals) {
			wait = intervals[i]
		} else {
			wait = 10 * time.Second
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
		if len(d) > 0 {
			return d
		}
		if activeStatuses[status] {
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

	var resp ApplicationSpaceResponse
	err := clientCtx.GetClient().Get(ctx, "/projects/"+data.Id(), &resp)
	if readHasFailed(&d, err, data) {
		return d
	}

	data.SetId(resp.ID)

	// Set both customer_id and organization_id for backward compatibility
	// Determine which field was used in the configuration
	if _, ok := data.GetOk(organizationIDKey); ok {
		setData(&d, data, organizationIDKey, resp.CustomerID)
	} else {
		setData(&d, data, customerIDKey, resp.CustomerID)
	}

	setData(&d, data, nameKey, resp.Name)
	setData(&d, data, displayNameKey, resp.DisplayName)
	setData(&d, data, descriptionKey, resp.Description)
	setData(&d, data, createTimeKey, resp.CreateTime)
	setData(&d, data, updateTimeKey, resp.UpdateTime)
	setData(&d, data, regionKey, resp.Region)
	setData(&d, data, ikgSizeKey, resp.IKGSize)
	setData(&d, data, replicaRegionKey, resp.ReplicaRegion)
	setDBConnectionData(&d, data, resp.DBConnection)

	return d
}

func resAppSpaceReadAfterCreateContext(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	var d diag.Diagnostics
	clientCtx := getClientContext(&d, meta)
	if clientCtx == nil {
		return d
	}
	ctx, cancel := context.WithTimeout(ctx, data.Timeout(schema.TimeoutCreate))
	defer cancel()

	return waitForActive(ctx, clientCtx, data)
}

func updateDBConnection(data *schema.ResourceData) *DBConnection {
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

	req := UpdateApplicationSpaceRequest{
		DisplayName:  updateOptionalString(data, displayNameKey),
		Description:  updateOptionalString(data, descriptionKey),
		DBConnection: updateDBConnection(data),
	}

	var resp ApplicationSpaceResponse
	err := clientCtx.GetClient().Put(ctx, "/projects/"+data.Id(), req, &resp)
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
	err := clientCtx.GetClient().Delete(ctx, "/projects/"+data.Id())
	HasFailed(&d, err)

	return d
}
