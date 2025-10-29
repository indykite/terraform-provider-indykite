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
	"encoding/base64"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

var (
	nameCheck = regexp.MustCompile(`^[a-z]+[-a-z0-9]*[a-z0-9]+$`)

	// TODO improve the regexp pattern.
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$") //nolint:lll //nolint:lll
)

// ValidateName is Terraform validation helper to verify value is valid name.
func ValidateName(i any, path cty.Path) diag.Diagnostics {
	var ret diag.Diagnostics
	v, ok := i.(string)
	if !ok {
		return append(ret, buildPluginErrorWithPath(
			fmt.Sprintf("validateName failed, expected string, got %T", i),
			path,
		))
	}
	if l := len(v); l < 2 || l > 254 {
		ret = append(ret, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       fmt.Sprintf("expected string value betweem 2 and 254 runes but received %d", l),
			AttributePath: path,
		})
	}
	if !nameCheck.MatchString(v) {
		ret = append(ret, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Invalid name",
			Detail:        "Value can have lowercase letters, digits, or hyphens. It must start with a lowercase letter and end with a letter or number.",
			AttributePath: path,
		})
	}
	return ret
}

// ValidateEmail is Terraform validation helper to verify value is valid email.
func ValidateEmail(i any, path cty.Path) diag.Diagnostics {
	var ret diag.Diagnostics
	v, ok := i.(string)
	if !ok {
		return append(ret, buildPluginErrorWithPath(
			fmt.Sprintf("validateEmail failed, expected string, got %T", i),
			path,
		))
	}

	if l := len(v); l < 2 || l > 254 {
		ret = append(ret, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       fmt.Sprintf("expected string value betweem 2 and 254 runes but received %d", l),
			AttributePath: path,
		})
	}

	if !emailRegex.MatchString(v) {
		ret = append(ret, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Value is not valid email address",
			AttributePath: path,
		})
	}
	return ret
}

// ValidateGID is Terraform validation helper to verify value is valid gid.
func ValidateGID(i any, path cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	var errSummary string
	switch {
	case !ok:
		errSummary = "expected type to be string"
	case !strings.HasPrefix(v, "gid:"):
		errSummary = "expected to have 'gid:' prefix"
	case len(v) < 22, len(v) > 254:
		errSummary = "expected to have len between 22 and 254 characters"
	default:
		if _, err := base64.RawURLEncoding.DecodeString(v[4:]); err != nil {
			errSummary = "expected to be a valid Raw URL Base64 string with 'gid:' prefix, got " + err.Error()
		}
	}

	if errSummary != "" {
		return diag.Diagnostics{{
			Severity:      diag.Error,
			Detail:        errSummary,
			Summary:       "Invalid ID value",
			AttributePath: path,
		}}
	}
	return nil
}

// ValidateYaml is Terraform validation helper to verify value is valid YAML.
func ValidateYaml(i any, path cty.Path) diag.Diagnostics {
	var ret diag.Diagnostics
	v, ok := i.(string)
	if !ok {
		return append(ret, buildPluginErrorWithPath(
			fmt.Sprintf("validateYaml failed, expected string, got %T", i),
			path,
		))
	}
	var y any
	if err := yaml.Unmarshal([]byte(v), &y); err != nil {
		return diag.Diagnostics{{
			Severity:      diag.Error,
			Summary:       err.Error(),
			AttributePath: path,
		}}
	}
	return nil
}

// DisplayNameDiffSuppress suppress Terraform changes when it contains name returned from API.
func DisplayNameDiffSuppress(k, old, newVal string, d *schema.ResourceData) bool {
	if k == displayNameKey && old == d.Get(nameKey).(string) && newVal == "" {
		return true
	}
	return false
}

// DisplayNameCredentialDiffSuppress suppress Terraform changes when it contains KID returned from API.
func DisplayNameCredentialDiffSuppress(k, old, newVal string, d *schema.ResourceData) bool {
	if k == displayNameKey && old == d.Get(kidKey).(string) && newVal == "" {
		return true
	}
	return false
}

// SuppressYamlDiff verify that 2 YAML strings are the same in value and suppress Terraform changes.
func SuppressYamlDiff(_, old, newVal string, _ *schema.ResourceData) bool {
	var oldMap, newMap map[string]any

	if err := yaml.Unmarshal([]byte(old), &oldMap); err != nil {
		return false
	}

	if err := yaml.Unmarshal([]byte(newVal), &newMap); err != nil {
		return false
	}

	return reflect.DeepEqual(oldMap, newMap)
}

// SuppressDurationDiff compares duration written as string and compare if value is the same or not.
// So values like 1h or 60m is the same.
func SuppressDurationDiff(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}
	var oldDur, newDur time.Duration
	var err error

	if oldDur, err = time.ParseDuration(oldValue); err != nil {
		return false
	}
	if newDur, err = time.ParseDuration(newValue); err != nil {
		return false
	}

	return oldDur == newDur
}

func optionalString(data *schema.ResourceData, key string) *string {
	v, ok := data.Get(key).(string)
	if !ok || v == "" {
		return nil
	}
	return &v
}

func updateOptionalString(data *schema.ResourceData, key string) *string {
	if !data.HasChange(key) {
		return nil
	}
	v, ok := data.Get(key).(string)
	if !ok {
		return nil
	}
	return &v
}

func setData(d *diag.Diagnostics, data *schema.ResourceData, attr string, value any) {
	if valOf := reflect.ValueOf(value); value == nil || (valOf.Kind() == reflect.Ptr && valOf.IsNil()) {
		if err := data.Set(attr, nil); err != nil {
			*d = append(*d, diag.Diagnostic{
				Severity:      diag.Error,
				Detail:        err.Error(),
				Summary:       "Cannot add attribute",
				AttributePath: cty.Path{cty.GetAttrStep{Name: attr}},
			})
			return
		}
	}

	if v, ok := value.(time.Time); ok {
		if v.IsZero() {
			return
		}
		value = v.Format(time.RFC3339Nano)
	}

	if err := data.Set(attr, value); err != nil {
		*d = append(*d, diag.Diagnostic{
			Severity:      diag.Error,
			Detail:        err.Error(),
			Summary:       "Cannot add attribute",
			AttributePath: cty.Path{cty.GetAttrStep{Name: attr}},
		})
	}
}

// hasDeleteProtection returns true if resource is protected from deletion.
func hasDeleteProtection(d *diag.Diagnostics, data *schema.ResourceData) bool {
	if v, ok := data.GetOk(deletionProtectionKey); ok && v.(bool) {
		*d = append(*d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Delete Protection is enabled",
			Detail:   "Cannot destroy instance without setting deletion_protection=false and running `terraform apply`",
		})
		return true
	}
	return false
}

// HasFailed checks if error is not nil and if it is, it will add it to diagnostics.
func HasFailed(d *diag.Diagnostics, err error) bool {
	switch {
	case err == nil:
		return false

	case IsServiceError(err):
		*d = append(*d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Communication with IndyKite failed, please try again later",
			Detail:   err.Error(),
		})

	case IsNotFoundError(err):
		*d = append(*d, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Resource not found",
			Detail:   err.Error(),
		})

	default:
		*d = append(*d, buildPluginError(err.Error()))
	}

	return true
}

func readHasFailed(d *diag.Diagnostics, err error, data *schema.ResourceData) bool {
	if HasFailed(d, err) {
		if IsNotFoundError(err) {
			_ = schema.RemoveFromState(data, nil)
		}
		return true
	}
	return false
}

// rawArrayToTypedArray casts raw data to []any and next to []string.
func rawArrayToTypedArray[T string | []byte](rawData any) []T {
	strArr := make([]T, len(rawData.([]any)))
	if len(strArr) == 0 {
		return nil
	}
	for i, el := range rawData.([]any) {
		strArr[i] = T(el.(string)) // always cast to string and then to type, because Terraform returns string always
	}
	return strArr
}

func buildPluginError(summary string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  summary,
		Detail:   "This is IndyKite plugin error, please report this issue to us! Thank you",
	}
}

func buildPluginErrorWithAttrName(summary, attr string) diag.Diagnostic {
	d := buildPluginError(summary)
	d.AttributePath = cty.Path{cty.GetAttrStep{Name: attr}}
	return d
}

func buildPluginErrorWithPath(summary string, path cty.Path) diag.Diagnostic {
	d := buildPluginError(summary)
	d.AttributePath = path
	return d
}

func getMapStringKeys[V any](in map[string]V) []string {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	// Terraform builds errors based on that and also documentation.
	// To be consistent, we need to sort it.
	sort.Strings(keys)
	return keys
}

// ReverseProtoEnumMap create reverse map, where value is key and key is value of Proto Enum.
func ReverseProtoEnumMap[Key, Value comparable](in map[Key]Value) map[Value]Key {
	reversed := make(map[Value]Key)
	for k, v := range in {
		reversed[v] = k
	}
	return reversed
}

// AuthorizationPolicyStatusTypes defines all supported StatusTypes.
var AuthorizationPolicyStatusTypes = map[string]string{
	"active":   "active",
	"inactive": "inactive",
	"draft":    "draft",
}

// AuthorizationPolicyStatusToAPI maps Terraform status values to API values.
var AuthorizationPolicyStatusToAPI = map[string]string{
	"active":   "ACTIVE",
	"inactive": "INACTIVE",
	"draft":    "DRAFT",
}

// AuthorizationPolicyStatusFromAPI maps API status values to Terraform values.
var AuthorizationPolicyStatusFromAPI = map[string]string{
	"ACTIVE":   "active",
	"INACTIVE": "inactive",
	"DRAFT":    "draft",
}

// KnowledgeQueryStatusTypes defines all supported StatusTypes.
var KnowledgeQueryStatusTypes = map[string]string{
	"active":   "active",
	"inactive": "inactive",
	"draft":    "draft",
}

// KnowledgeQueryStatusToAPI maps Terraform status values to API values.
var KnowledgeQueryStatusToAPI = map[string]string{
	"active":   "ACTIVE",
	"inactive": "INACTIVE",
	"draft":    "DRAFT",
}

// KnowledgeQueryStatusFromAPI maps API status values to Terraform values.
var KnowledgeQueryStatusFromAPI = map[string]string{
	"ACTIVE":   "active",
	"INACTIVE": "inactive",
	"DRAFT":    "draft",
}

// ProtoValidateError tries to define interface for all Proto Validation errors,
// so we can generate better errors back to user.
type ProtoValidateError interface {
	// Field function returns field value.
	Field() string
	// Reason function returns reason value.
	Reason() string
	// Cause function returns cause value.
	Cause() error
	// Key function returns key value.
	Key() bool
	// ErrorName returns error name.
	ErrorName() string
}

// IngestPipelineOperationTypes defines all supported IngestPipelineOperationTypes and its mapping.
var IngestPipelineOperationTypes = map[string]string{
	"OPERATION_UPSERT_NODE":                  "OPERATION_UPSERT_NODE",
	"OPERATION_UPSERT_RELATIONSHIP":          "OPERATION_UPSERT_RELATIONSHIP",
	"OPERATION_DELETE_NODE":                  "OPERATION_DELETE_NODE",
	"OPERATION_DELETE_RELATIONSHIP":          "OPERATION_DELETE_RELATIONSHIP",
	"OPERATION_DELETE_NODE_PROPERTY":         "OPERATION_DELETE_NODE_PROPERTY",
	"OPERATION_DELETE_RELATIONSHIP_PROPERTY": "OPERATION_DELETE_RELATIONSHIP_PROPERTY",
}

// IngestPipelineOperationTypesReverse is reverse mapping of IngestPipelineOperationTypes.
var IngestPipelineOperationTypesReverse = ReverseProtoEnumMap(IngestPipelineOperationTypes)

// ExternalDataResolverConfigContentType defines all supported ContentTypes and its mapping.
var ExternalDataResolverConfigContentType = map[string]string{
	"json": "json",
}

// TrustScoreProfileScheduleFrequencies defines all supported frequencies for trust score.
var TrustScoreProfileScheduleFrequencies = map[string]string{
	"UPDATE_FREQUENCY_INVALID":      "UPDATE_FREQUENCY_INVALID",
	"UPDATE_FREQUENCY_THREE_HOURS":  "UPDATE_FREQUENCY_THREE_HOURS",
	"UPDATE_FREQUENCY_SIX_HOURS":    "UPDATE_FREQUENCY_SIX_HOURS",
	"UPDATE_FREQUENCY_TWELVE_HOURS": "UPDATE_FREQUENCY_TWELVE_HOURS",
	"UPDATE_FREQUENCY_DAILY":        "UPDATE_FREQUENCY_DAILY",
}

// TrustScoreProfileScheduleToAPI maps Terraform schedule values to API values.
var TrustScoreProfileScheduleToAPI = map[string]string{
	"UPDATE_FREQUENCY_INVALID":      "INVALID",
	"UPDATE_FREQUENCY_THREE_HOURS":  "THREE_HOURS",
	"UPDATE_FREQUENCY_SIX_HOURS":    "SIX_HOURS",
	"UPDATE_FREQUENCY_TWELVE_HOURS": "TWELVE_HOURS",
	"UPDATE_FREQUENCY_DAILY":        "DAILY",
}

// TrustScoreProfileScheduleFromAPI maps API schedule values to Terraform values.
var TrustScoreProfileScheduleFromAPI = map[string]string{
	"INVALID":      "UPDATE_FREQUENCY_INVALID",
	"THREE_HOURS":  "UPDATE_FREQUENCY_THREE_HOURS",
	"SIX_HOURS":    "UPDATE_FREQUENCY_SIX_HOURS",
	"TWELVE_HOURS": "UPDATE_FREQUENCY_TWELVE_HOURS",
	"DAILY":        "UPDATE_FREQUENCY_DAILY",
}

// TrustScoreDimensionNames defines all supported dimensions names for trust score.
// Note: NAME_INVALID is not included as it's not accepted by the API.
var TrustScoreDimensionNames = map[string]string{
	"NAME_FRESHNESS":    "NAME_FRESHNESS",
	"NAME_COMPLETENESS": "NAME_COMPLETENESS",
	"NAME_VALIDITY":     "NAME_VALIDITY",
	"NAME_ORIGIN":       "NAME_ORIGIN",
	"NAME_VERIFICATION": "NAME_VERIFICATION",
}

// TrustScoreDimensionToAPI maps Terraform dimension names to API values.
var TrustScoreDimensionToAPI = map[string]string{
	"NAME_FRESHNESS":    "FRESHNESS",
	"NAME_COMPLETENESS": "COMPLETENESS",
	"NAME_VALIDITY":     "VALIDITY",
	"NAME_ORIGIN":       "ORIGIN",
	"NAME_VERIFICATION": "VERIFICATION",
}

// TrustScoreDimensionFromAPI maps API dimension names to Terraform values.
var TrustScoreDimensionFromAPI = map[string]string{
	"FRESHNESS":    "NAME_FRESHNESS",
	"COMPLETENESS": "NAME_COMPLETENESS",
	"VALIDITY":     "NAME_VALIDITY",
	"ORIGIN":       "NAME_ORIGIN",
	"VERIFICATION": "NAME_VERIFICATION",
}

func contains[E comparable](arr []E, el E) bool {
	for _, v := range arr {
		if v == el {
			return true
		}
	}
	return false
}
