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
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	gidBase64Regex = regexp.MustCompile("^gid:[A-Za-z0-9_-]{22,}$")
)

func basicStateImporter(_ context.Context, data *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	if err := parseImportID(data); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func parseImportID(d *schema.ResourceData) error {
	importID := d.Id()

	// Check if it's a GID (direct ID import)
	if gidBase64Regex.MatchString(importID) {
		return nil
	}

	// Check if it's a name with location parameter (e.g., "my-resource?location=gid:xxx")
	// The ID will be kept as-is and used in the read function
	if strings.Contains(importID, "?location=") {
		return nil
	}

	return errors.New("Unimplemented id format: " + importID +
		". Expected either 'gid:xxx' or 'resource-name?location=gid:xxx'")
}

// buildReadPath constructs the API path for reading a resource.
// It supports both:
//   - Direct ID: data.Id() = "gid:xxx" -> returns "/resource/gid:xxx"
//   - Name with location: data.Id() = "my-name?location=gid:xxx" -> returns "/resource/my-name?<param>=gid:xxx"
//     where <param> is translated to the correct API parameter (project_id or organization_id)
func buildReadPath(resourcePath string, data *schema.ResourceData) string {
	id := data.Id()

	// If the ID contains a query parameter, it's a name+location format
	// Translate the generic "location" parameter to the correct API parameter
	if strings.Contains(id, "?location=") {
		// Determine the correct parameter name based on the resource path
		var apiParam string
		switch {
		case strings.Contains(resourcePath, "/projects"):
			// Application spaces use organization_id
			apiParam = "organization_id"
		case strings.Contains(resourcePath, "/service-accounts"):
			// Service accounts use organization_id
			apiParam = "organization_id"
		default:
			// Most other resources use project_id (applications, agents, policies, etc.)
			apiParam = "project_id"
		}

		// Replace "location=" with the correct API parameter
		translatedID := strings.Replace(id, "?location=", "?"+apiParam+"=", 1)
		return resourcePath + "/" + translatedID
	}

	// Otherwise, it's a direct ID
	return resourcePath + "/" + id
}
