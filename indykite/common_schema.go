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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	nameKey               = "name"
	displayNameKey        = "display_name"
	descriptionKey        = "description"
	locationKey           = "location"
	customerIDKey         = "customer_id"
	appSpaceIDKey         = "app_space_id"
	applicationIDKey      = "application_id"
	appAgentIDKey         = "app_agent_id"
	createTimeKey         = "create_time"
	updateTimeKey         = "update_time"
	deletionProtectionKey = "deletion_protection"
	filterKey             = "filter"
	regionKey             = "region"
	apiPermissionsKey     = "api_permissions"
)

const (
	locationDescription      = `Identifier of Location, where to create resource`
	customerIDDescription    = `Identifier of Customer`
	appSpaceIDDescription    = `Identifier of Application Space`
	applicationIDDescription = `Identifier of Application`
	appAgentIDDescription    = `Identifier of Application Agent`

	nameDescription = `Unique client assigned immutable identifier. Can not be updated without creating a new resource.`
)

func regionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: `Region where the application space is located.`,
		ValidateFunc: validation.StringInSlice([]string{
			"europe-west1", "us-east1",
		}, false),
	}
}

func convertToOptional(in *schema.Schema) *schema.Schema {
	in.Required = false
	in.Optional = true
	return in
}

func keysWithoutCurrent(currentKey string, allKeys []string) []string {
	newKeys := make([]string, 0, len(allKeys)-1)
	for _, k := range allKeys {
		if k != currentKey {
			newKeys = append(newKeys, k)
		}
	}
	return newKeys
}

func setExactlyOneOf(base *schema.Schema, currentKey string, oneOfKeys []string) *schema.Schema {
	base = convertToOptional(base)
	base.ConflictsWith = keysWithoutCurrent(currentKey, oneOfKeys)
	base.ExactlyOneOf = oneOfKeys
	return base
}

func setRequiredWith(base *schema.Schema, requiredWith ...string) *schema.Schema {
	base = convertToOptional(base)
	base.RequiredWith = requiredWith
	return base
}

func setComputed(base *schema.Schema) *schema.Schema {
	// required and optional must be false when compute to true, otherwise the change in TF files is allowed
	base.Required = false
	base.Optional = false
	base.Computed = true
	base.ForceNew = false
	base.ValidateDiagFunc = nil
	base.ValidateFunc = nil
	base.DiffSuppressFunc = nil
	base.MinItems = 0
	return base
}

func createTimeSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: `Timestamp when the Resource was created. Assigned by the server. A timestamp in RFC3339 UTC "Zulu" format, accurate to nanoseconds. Example: "2014-10-02T15:01:23.045123456Z".`,
	}
}

func updateTimeSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: `Timestamp when the Resource was last updated. Assigned by the server. A timestamp in RFC3339 UTC "Zulu" format, accurate to nanoseconds. Example: "2014-10-02T15:01:23.045123456Z".`,
	}
}

func displayNameSchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: DisplayNameDiffSuppress,
		Description:      `The display name for the instance. Can be updated without creating a new resource.`,
	}
}

func descriptionSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(0, 256),
		Description:  `Your own description of the resource. Must be less than or equal to 256 UTF-8 bytes.`,
	}
}

func nameSchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		ValidateDiagFunc: ValidateName,
		Required:         true,
		Description:      nameDescription,
	}
}

func baseIDSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		ValidateDiagFunc: ValidateGID,
		Required:         true,
		Description:      description,
	}
}

func exactNameFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		MinItems:    1,
		Description: `Filter customers based on given names. Using 'exact name match' strategy to find customer.`,
		Elem:        &schema.Schema{Type: schema.TypeString, ValidateDiagFunc: ValidateName},
	}
}

func apiPermissionsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringLenBetween(1, 64),
		},
		Description: `List of API permissions for the agent: Authorization, Capture, ContXIQ, EntityMatching, IKGRead and TrustedDataAccess.`,
	}
}

func locationSchema() *schema.Schema {
	return baseIDSchema(locationDescription)
}

func customerIDSchema() *schema.Schema {
	return baseIDSchema(customerIDDescription)
}

func appSpaceIDSchema() *schema.Schema {
	return baseIDSchema(appSpaceIDDescription)
}
func applicationIDSchema() *schema.Schema {
	return baseIDSchema(applicationIDDescription)
}

func appAgentIDSchema() *schema.Schema {
	return baseIDSchema(appAgentIDDescription)
}

func deletionProtectionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: `Whether or not to allow Terraform to destroy the instance. Unless this field is set to false in Terraform state, a terraform destroy or terraform apply that would delete the instance will fail.`,
	}
}
