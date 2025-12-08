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
	createdByKey          = "created_by"
	updatedByKey          = "updated_by"
	ikgStatusKey          = "ikg_status"
	filterKey             = "filter"
	regionKey             = "region"
	apiPermissionsKey     = "api_permissions"
	ikgSizeKey            = "ikg_size"
	replicaRegionKey      = "replica_region"
	dbConnectionKey       = "db_connection"
	dbURLKey              = "url"
	dbUsernameKey         = "username"
	dbPasswordKey         = "password"
	dbNameKey             = "name"
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
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
		Description: `Region where the application space is located.
		Valid values are: europe-west1, us-east1.`,
		ValidateFunc: validation.StringInSlice([]string{
			"europe-west1", "us-east1",
		}, false),
	}
}

func ikgSizeSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Default:  "2GB",
		ForceNew: true,
		Description: `IKG size that will be allocated, which corresponds also to number of CPU nodes (default 2GB).
		Valid values are: 2GB (1 CPU), 4GB (1 CPU), 8GB (2 CPUs), 16GB (3 CPUs), 32GB (6 CPUs), 64GB (12 CPUs),
		128GB (24 CPUs), 192GB (36 CPUs), 256GB (48 CPUs), 384GB (82 CPUs), and 512GB (96 CPUs).`,
		ValidateFunc: validation.StringInSlice([]string{
			"2GB", "4GB", "8GB", "16GB", "32GB", "64GB", "128GB", "192GB", "256GB", "384GB", "512GB",
		}, false),
	}
}

func ikgSizeComputedSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
		Description: `IKG size that will be allocated, which corresponds also to number of CPU nodes.
		Valid values are: 2GB (1 CPU), 4GB (1 CPU), 8GB (2 CPUs), 16GB (3 CPUs), 32GB (6 CPUs), 64GB (12 CPUs),
		128GB (24 CPUs), 192GB (36 CPUs), 256GB (48 CPUs), 384GB (82 CPUs), and 512GB (96 CPUs).`,
	}
}

func replicaRegionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: true,
		Description: `Replica region specifies where the replica IKG is created.
		Replica must be a different region than the master, but also on the same geographical continent.
		Valid values are: europe-west1, us-east1, us-west1.`,
		ValidateFunc: validation.StringInSlice([]string{
			"europe-west1", "us-east1", "us-west1",
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
		ValidateFunc: validation.StringLenBetween(0, 65000),
		Description:  `Your own description of the resource. Must be less than or equal to 65000 UTF-8 bytes.`,
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
		Required: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{
				"Authorization", "Capture", "ContXIQ", "EntityMatching", "IKGRead", "TrustedDataAccess",
			}, false),
		},
		Description: `List of API permissions for the agent: Authorization, Capture, ContXIQ, EntityMatching, IKGRead and TrustedDataAccess.`,
	}
}

func dbConnectionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Description: "DBConnection",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				dbURLKey: {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "Connection URL for the database",
					ValidateFunc: validation.StringIsNotEmpty,
				},
				dbUsernameKey: {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "Username for database authentication",
					ValidateFunc: validation.StringIsNotEmpty,
				},
				dbPasswordKey: {
					Type:         schema.TypeString,
					Required:     true,
					Sensitive:    true,
					Description:  "Password for database authentication",
					ValidateFunc: validation.StringIsNotEmpty,
				},
				dbNameKey: {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Optional database name",
				},
			},
		},
	}
}

func dbConnectionComputedSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "DBConnection",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				dbURLKey: {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Connection URL for the database",
				},
				dbUsernameKey: {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Username for database authentication",
				},
				dbPasswordKey: {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "Password for database authentication",
				},
				dbNameKey: {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Optional database name",
				},
			},
		},
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

func createdBySchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Identifier of the user who created the resource",
	}
}

func updatedBySchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Identifier of the user who last updated the resource",
	}
}

func ikgStatusSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Status of the Identity Knowledge Graph",
	}
}
