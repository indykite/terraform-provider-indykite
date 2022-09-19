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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	gidBase64Regex = regexp.MustCompile("^gid:[A-Za-z0-9_-]{22,}$")
)

func basicStateImporter(_ context.Context, data *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	if err := parseImportID(data); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func parseImportID(d *schema.ResourceData) error {
	if gidBase64Regex.MatchString(d.Id()) {
		return nil
	}

	return errors.New("Unimplemented id format: " + d.Id())
}
