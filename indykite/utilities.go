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
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkerrors "github.com/indykite/indykite-sdk-go/errors"
	configpb "github.com/indykite/indykite-sdk-go/gen/indykite/config/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gopkg.in/yaml.v3"
)

var (
	nameCheck = regexp.MustCompile(`^[a-z]+[-a-z0-9]*[a-z0-9]+$`)

	// TODO improve the regexp pattern.
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$") //nolint:lll
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

func optionalString(data *schema.ResourceData, key string) *wrapperspb.StringValue {
	v, ok := data.Get(key).(string)
	if !ok || v == "" {
		return nil
	}
	return wrapperspb.String(v)
}

// flattenOptionalString returns String if v is not nil and v is not empty else returns nil.
func flattenOptionalString(v *wrapperspb.StringValue) any {
	if v != nil && v.Value != "" {
		return v.Value
	}
	return nil
}

func updateOptionalString(data *schema.ResourceData, key string) *wrapperspb.StringValue {
	if !data.HasChange(key) {
		return nil
	}
	v, ok := data.Get(key).(string)
	if !ok {
		return nil
	}
	return wrapperspb.String(v)
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

	switch v := value.(type) {
	case *wrapperspb.StringValue:
		value = v.GetValue()
	case *wrapperspb.Int32Value:
		value = v.GetValue()
	case *wrapperspb.Int64Value:
		value = v.GetValue()
	case *wrapperspb.UInt32Value:
		value = v.GetValue()
	case *wrapperspb.UInt64Value:
		value = v.GetValue()
	case *wrapperspb.BoolValue:
		value = v.GetValue()
	case *wrapperspb.BytesValue:
		value = v.GetValue()
	case *wrapperspb.FloatValue:
		value = v.GetValue()
	case *timestamppb.Timestamp:
		if v == nil {
			return
		}
		t := v.AsTime()
		if t.IsZero() {
			return
		}
		value = t.Format(time.RFC3339Nano)
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

	case sdkerrors.IsServiceError(err):
		*d = append(*d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Communication with IndyKite failed, please try again later",
			Detail:   err.Error(),
		})

	case sdkerrors.IsNotFoundError(err):
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
		if sdkerrors.IsNotFoundError(err) {
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

// AuthorizationPolicyStatusTypes defines all supported StatusTypes and its mapping.
var AuthorizationPolicyStatusTypes = map[string]configpb.AuthorizationPolicyConfig_Status{
	"active":   configpb.AuthorizationPolicyConfig_STATUS_ACTIVE,
	"inactive": configpb.AuthorizationPolicyConfig_STATUS_INACTIVE,
}

// ProtoValidateError tries to define interface for all Proto Validation errors,
// so we can generate better errors back to user.
type ProtoValidateError interface {
	Field() string
	Reason() string
	Cause() error
	Key() bool
	ErrorName() string
}

// IngestPipelineOperationTypes defines all supported IngestPipelineOperationTypes and its mapping.
//
//nolint:lll
var IngestPipelineOperationTypes = map[string]configpb.IngestPipelineOperation{
	"OPERATION_UPSERT_NODE":                  configpb.IngestPipelineOperation_INGEST_PIPELINE_OPERATION_UPSERT_NODE,
	"OPERATION_UPSERT_RELATIONSHIP":          configpb.IngestPipelineOperation_INGEST_PIPELINE_OPERATION_UPSERT_RELATIONSHIP,
	"OPERATION_DELETE_NODE":                  configpb.IngestPipelineOperation_INGEST_PIPELINE_OPERATION_DELETE_NODE,
	"OPERATION_DELETE_RELATIONSHIP":          configpb.IngestPipelineOperation_INGEST_PIPELINE_OPERATION_DELETE_RELATIONSHIP,
	"OPERATION_DELETE_NODE_PROPERTY":         configpb.IngestPipelineOperation_INGEST_PIPELINE_OPERATION_DELETE_NODE_PROPERTY,
	"OPERATION_DELETE_RELATIONSHIP_PROPERTY": configpb.IngestPipelineOperation_INGEST_PIPELINE_OPERATION_DELETE_RELATIONSHIP_PROPERTY,
}

// IngestPipelineOperationTypesReverse is reverse mapping of IngestPipelineOperationTypes.
var IngestPipelineOperationTypesReverse = ReverseProtoEnumMap(IngestPipelineOperationTypes)

// ExternalDataResolverConfigContentType defines all supported ContentTypes and its mapping.
var ExternalDataResolverConfigContentType = map[string]configpb.ExternalDataResolverConfig_ContentType{
	"json": configpb.ExternalDataResolverConfig_CONTENT_TYPE_JSON,
}

var externalDataResolverContentTypeToString = map[configpb.ExternalDataResolverConfig_ContentType]string{
	configpb.ExternalDataResolverConfig_CONTENT_TYPE_INVALID: "invalid",
	configpb.ExternalDataResolverConfig_CONTENT_TYPE_JSON:    "json",
}

func betterValidationErrorWithPath(err error) error {
	var protoValidErr ProtoValidateError
	if errors.As(err, &protoValidErr) {
		err = handleProtoValidationError(protoValidErr, true)
	}
	return err
}

func handleProtoValidationError(err ProtoValidateError, withPath bool) error {
	path := []string{
		err.Field(),
	}
	for err.Cause() != nil {
		var causeErr ProtoValidateError
		if !errors.As(err.Cause(), &causeErr) {
			break
		}
		path = append(path, causeErr.Field())
		err = causeErr
	}

	cause := ""
	if err.Cause() != nil {
		cause = " caused by: " + err.Cause().Error()
	}

	var attribute string
	if withPath {
		attribute = strings.Join(path, ".")
	} else {
		attribute = path[len(attribute)-1]
	}

	return fmt.Errorf("invalid %s: %s%s", attribute, err.Reason(), cause)
}

func contains[E comparable](arr []E, el E) bool {
	for _, v := range arr {
		if v == el {
			return true
		}
	}
	return false
}
