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
	nameCheck = regexp.MustCompile(`^[a-z]+(?:[-a-z0-9]*)*[a-z0-9]+$`)

	// TODO improve the regexp pattern
	//nolint:lll
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	// pemRegex defines regex to match PEM Private key format.
	pemRegex = regexp.MustCompile(`^-----BEGIN PRIVATE KEY-----(?:(?s).*)-----END PRIVATE KEY-----(?:\n)?$`)
)

var supportedSigningAlgs = []string{
	"RS256", "RS384", "RS512", "PS256", "PS384", "PS512", "ES256", "ES384",
	"ES512", "ES256K", "HS256", "HS384", "HS512", "EdDSA",
}

// ValidateName is Terraform validation helper to verify value is valid name.
func ValidateName(i interface{}, path cty.Path) (ret diag.Diagnostics) {
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
func ValidateEmail(i interface{}, path cty.Path) (ret diag.Diagnostics) {
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
func ValidateGID(i interface{}, path cty.Path) (ret diag.Diagnostics) {
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
func ValidateYaml(i interface{}, path cty.Path) (ret diag.Diagnostics) {
	v, ok := i.(string)
	if !ok {
		return append(ret, buildPluginErrorWithPath(
			fmt.Sprintf("validateYaml failed, expected string, got %T", i),
			path,
		))
	}
	var y interface{}
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
func DisplayNameDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	if k == displayNameKey && old == d.Get(nameKey).(string) && new == "" {
		return true
	}
	return false
}

// DisplayNameCredentialDiffSuppress suppress Terraform changes when it contains KID returned from API.
func DisplayNameCredentialDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	if k == displayNameKey && old == d.Get(kidKey).(string) && new == "" {
		return true
	}
	return false
}

// SuppressYamlDiff verify that 2 YAML strings are the same in value and suppress Terraform changes.
func SuppressYamlDiff(k, old, new string, _ *schema.ResourceData) bool {
	var oldMap, newMap map[string]interface{}

	if err := yaml.Unmarshal([]byte(old), &oldMap); err != nil {
		return false
	}

	if err := yaml.Unmarshal([]byte(new), &newMap); err != nil {
		return false
	}

	return reflect.DeepEqual(oldMap, newMap)
}

func optionalString(data *schema.ResourceData, key string) *wrapperspb.StringValue {
	v, ok := data.Get(key).(string)
	if !ok || v == "" {
		return nil
	}
	return wrapperspb.String(v)
}

func stringOrEmpty(data *schema.ResourceData, key string) string {
	v, _ := data.Get(key).(string)
	return v
}

// flattenOptionalString returns String if v is not nil and v is not empty else returns nil.
func flattenOptionalString(v *wrapperspb.StringValue) interface{} {
	if v != nil && v.Value != "" {
		return v.Value
	}
	return nil
}

func flattenOptionalMap(data map[string]string) map[string]string {
	if len(data) == 0 {
		return nil
	}
	return data
}

func flattenOptionalArray(data []string) []string {
	if len(data) == 0 {
		return nil
	}
	return data
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

func setData(d *diag.Diagnostics, data *schema.ResourceData, attr string, value interface{}) {
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

func hasFailed(d *diag.Diagnostics, err error) bool {
	if err != nil {
		if sdkerrors.IsServiceError(err) {
			*d = append(*d, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Communication with IndyKite failed, please try again later",
				Detail:   err.Error(),
			})
		} else {
			*d = append(*d, buildPluginError(err.Error()))
		}

		return true
	}
	return false
}

// rawArrayToStringArray casts raw data to []interface{} and next to []string.
func rawArrayToStringArray(rawData interface{}) []string {
	strArr := make([]string, len(rawData.([]interface{})))
	if len(strArr) == 0 {
		return nil
	}
	for i, el := range rawData.([]interface{}) {
		strArr[i] = el.(string)
	}
	return strArr
}

// rawMapToStringMap casts raw data to map[string]interface{} and next convert to map[string]string.
func rawMapToStringMap(rawData interface{}) map[string]string {
	out := make(map[string]string)
	for i, el := range rawData.(map[string]interface{}) {
		out[i] = el.(string)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func stringToOptionalStringWrapper(in string) *wrapperspb.StringValue {
	if len(in) == 0 {
		return nil
	}
	return wrapperspb.String(in)
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

// OAuth2GrantTypes defines all supported GrantTypes and its mapping.
var OAuth2GrantTypes = map[string]configpb.GrantType{
	"authorization_code": configpb.GrantType_GRANT_TYPE_AUTHORIZATION_CODE,
	"implicit":           configpb.GrantType_GRANT_TYPE_IMPLICIT,
	"password":           configpb.GrantType_GRANT_TYPE_PASSWORD,
	"client_credentials": configpb.GrantType_GRANT_TYPE_CLIENT_CREDENTIALS,
	"refresh_token":      configpb.GrantType_GRANT_TYPE_REFRESH_TOKEN,
}

// OAuth2ResponseTypes defines all supported ResponseTypes and its mapping.
var OAuth2ResponseTypes = map[string]configpb.ResponseType{
	"token":    configpb.ResponseType_RESPONSE_TYPE_TOKEN,
	"code":     configpb.ResponseType_RESPONSE_TYPE_CODE,
	"id_token": configpb.ResponseType_RESPONSE_TYPE_ID_TOKEN,
}

// OAuth2TokenEndpointAuthMethods defines all supported Token Endpoint Auth Methods and its mapping.
var OAuth2TokenEndpointAuthMethods = map[string]configpb.TokenEndpointAuthMethod{
	"client_secret_basic": configpb.TokenEndpointAuthMethod_TOKEN_ENDPOINT_AUTH_METHOD_CLIENT_SECRET_BASIC,
	"client_secret_post":  configpb.TokenEndpointAuthMethod_TOKEN_ENDPOINT_AUTH_METHOD_CLIENT_SECRET_POST,
	"private_key_jwt":     configpb.TokenEndpointAuthMethod_TOKEN_ENDPOINT_AUTH_METHOD_PRIVATE_KEY_JWT,
	"none":                configpb.TokenEndpointAuthMethod_TOKEN_ENDPOINT_AUTH_METHOD_NONE,
}

// OAuth2TokenEndpointAuthMethodsReverse defines all supported Token Endpoint Auth Methods and its reversed mapping.
var OAuth2TokenEndpointAuthMethodsReverse = ReverseProtoEnumMap(OAuth2TokenEndpointAuthMethods)

// OAuth2ClientSubjectTypes defines all supported Client Subjects and its mapping.
var OAuth2ClientSubjectTypes = map[string]configpb.ClientSubjectType{
	"public":   configpb.ClientSubjectType_CLIENT_SUBJECT_TYPE_PUBLIC,
	"pairwise": configpb.ClientSubjectType_CLIENT_SUBJECT_TYPE_PAIRWISE,
}

// OAuth2ClientSubjectTypesReverse defines all supported Client Subjects and its reversed mapping.
var OAuth2ClientSubjectTypesReverse = ReverseProtoEnumMap(OAuth2ClientSubjectTypes)

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
