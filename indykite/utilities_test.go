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

package indykite_test

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utilities", func() {
	invalidBase64Error := "expected to be a valid Raw URL Base64 string with 'gid:' prefix, got illegal base64 data"
	DescribeTable("ValidateBase64ID",
		func(input any, errStringMatcher OmegaMatcher) {
			path := cty.IndexIntPath(22)
			d := indykite.ValidateGID(input, path)

			if errStringMatcher == nil {
				Expect(d).To(BeEmpty())
			} else {
				Expect(d).To(HaveLen(1))
				Expect(d[0].Detail).To(errStringMatcher)
				Expect(d[0].Summary).To(Equal("Invalid ID value"))
				Expect(d[0].AttributePath).To(Equal(path))
			}
		},
		Entry("Not a string", 22, Equal("expected type to be string")),
		Entry("No prefix", "abc", Equal("expected to have 'gid:' prefix")),
		Entry("Empty gid", "gid:", Equal("expected to have len between 22 and 254 characters")),
		Entry("Not valid Base64", "gid:", Equal("expected to have len between 22 and 254 characters")),
		Entry("Not valid Base64", "gid:##################", HavePrefix(invalidBase64Error)),
		Entry("Not valid - got Std Base64", "gid:SGVsbG8gSW5keUtpdGU=", HavePrefix(invalidBase64Error)),
		Entry("Not valid - got Raw Std Base64", "gid:SGVsbG9+SW5keUtpdGU", HavePrefix(invalidBase64Error)),
		Entry("Not valid - got Raw Std Base64", "gid:SGVsbG/CsEluZHlLaXRlIQ", HavePrefix(invalidBase64Error)),
		Entry("Not valid - got URL Base64", "gid:SGVsbG9-SW5keUtpdGU=", HavePrefix(invalidBase64Error)),
		Entry("Valid Raw URL Base64", "gid:SGVsbG_CsEluZHlLaXRlIQ", nil),
	)

	DescribeTable("DisplayNameDiffSuppress",
		func(k, currentName, old, newVal string, expected OmegaMatcher) {
			resourceData := schema.TestResourceDataRaw(GinkgoT(),
				map[string]*schema.Schema{"name": {Type: schema.TypeString}},
				map[string]any{"name": currentName},
			)
			Expect(indykite.DisplayNameDiffSuppress(k, old, newVal, resourceData)).To(expected)
		},
		Entry("Different key", "description", "a", "b", "c", BeFalse()),
		// If new and old are the same, tested function is not called at all
		// We care only about situation, when new is empty and old is same as name
		Entry("Display name is same as name", "display_name", "abc", "abc", "", BeTrue()),
		Entry("Display name is different than name", "display_name", "abc", "jkl", "", BeFalse()),
	)

	DescribeTable("ExpireTimeDiffSuppress",
		func(old, newVal string, expected OmegaMatcher) {
			Expect(indykite.ExpireTimeDiffSuppress("expire_time", old, newVal, nil)).To(expected)
		},
		// Same instant expressed with a different timezone offset must be suppressed,
		// because the API normalizes the value to UTC (this is the drift being fixed).
		Entry("Same instant, UTC vs offset",
			"2026-12-31T13:34:56Z", "2026-12-31T12:34:56-01:00", BeTrue()),
		Entry("Same instant, extra nanoseconds",
			"2026-12-31T13:34:56Z", "2026-12-31T13:34:56.000Z", BeTrue()),
		Entry("Identical values", "2026-12-31T13:34:56Z", "2026-12-31T13:34:56Z", BeTrue()),
		// Genuinely different instants must NOT be suppressed so ForceNew still triggers.
		Entry("Different instant, one hour later",
			"2026-12-31T13:34:56Z", "2026-12-31T14:34:56Z", BeFalse()),
		Entry("Different instant, different day",
			"2026-12-31T13:34:56Z", "2026-11-30T12:34:56-01:00", BeFalse()),
		// Empty (adding or removing an expiry) does not parse and must NOT be suppressed.
		Entry("Adding expiry from empty", "", "2026-12-31T12:34:56-01:00", BeFalse()),
		Entry("Removing expiry", "2026-12-31T13:34:56Z", "", BeFalse()),
		// Unparseable value falls back to showing the diff.
		Entry("Unparseable new value", "2026-12-31T13:34:56Z", "not-a-time", BeFalse()),
	)

	DescribeTable("ValidateName",
		func(input any, expectedSummaries ...string) {
			path := cty.GetAttrPath("name")
			d := indykite.ValidateName(input, path)
			summaries := make([]string, 0, len(d))
			for _, diagnostic := range d {
				Expect(diagnostic.AttributePath).To(Equal(path))
				summaries = append(summaries, diagnostic.Summary)
			}
			Expect(summaries).To(ConsistOf(expectedSummaries))
		},
		Entry("2 characters", "ab"),
		Entry("63 characters", "a"+strings.Repeat("b", 62)),
		Entry("Double hyphen", "tf--name"),
		Entry("64 characters", "a"+strings.Repeat("b", 63),
			"expected string value between 2 and 63 characters but received 64"),
		Entry("1 character", "a",
			"expected string value between 2 and 63 characters but received 1", "Invalid name"),
		Entry("Uppercase", "tf-Name", "Invalid name"),
		Entry("Trailing hyphen", "tf-name-", "Invalid name"),
		Entry("Leading digit", "1tf-name", "Invalid name"),
		Entry("Not a string", 22, "validateName failed, expected string, got int"),
	)

	DescribeTable("CheckOptionalTextLength",
		func(maxLen int, input any, expectedErr OmegaMatcher) {
			_, errs := indykite.CheckOptionalTextLength(maxLen)(input, "description")
			if expectedErr == nil {
				Expect(errs).To(BeEmpty())
				return
			}
			Expect(errs).To(ConsistOf(MatchError(expectedErr)))
		},
		Entry("Empty clears the value", 254, "", nil),
		Entry("2 characters", 254, "ab", nil),
		Entry("Maximum length", 254, strings.Repeat("d", 254), nil),
		Entry("1 character", 254, "x",
			Equal("expected description to be empty or between 2 and 254 characters, got 1")),
		Entry("Above maximum", 254, strings.Repeat("d", 255),
			Equal("expected description to be empty or between 2 and 254 characters, got 255")),
		// The API counts characters: 32500 two-byte characters are 65000 bytes, 32501 are accepted too.
		Entry("Multi-byte characters count once", 65000, strings.Repeat("é", 32501), nil),
		Entry("Multi-byte above maximum", 10, strings.Repeat("é", 11),
			Equal("expected description to be empty or between 2 and 10 characters, got 11")),
		Entry("Not a string", 254, 22, Equal("expected type of description to be string")),
	)

	DescribeTable("CheckCharLength",
		func(input any, expectedErr OmegaMatcher) {
			_, errs := indykite.CheckCharLength(1, 150)(input, "audience")
			if expectedErr == nil {
				Expect(errs).To(BeEmpty())
				return
			}
			Expect(errs).To(ConsistOf(MatchError(expectedErr)))
		},
		Entry("Minimum length", "a", nil),
		Entry("Maximum length", strings.Repeat("a", 150), nil),
		Entry("Multi-byte characters count once", strings.Repeat("é", 150), nil),
		Entry("Empty", "", Equal("expected audience to be between 1 and 150 characters, got 0")),
		Entry("Above maximum", strings.Repeat("a", 151),
			Equal("expected audience to be between 1 and 150 characters, got 151")),
		Entry("Not a string", 22, Equal("expected type of audience to be string")),
	)

	DescribeTable("CheckMapValueCharLength",
		func(input any, expectedSummary string) {
			path := cty.GetAttrPath("claims_mapping")
			d := indykite.CheckMapValueCharLength(1, 200)(input, path)
			if expectedSummary == "" {
				Expect(d).To(BeEmpty())
				return
			}
			Expect(d).To(HaveLen(1))
			Expect(d[0].Summary).To(Equal(expectedSummary))
		},
		Entry("Values within limits", map[string]any{"email": "e", "name": strings.Repeat("n", 200)}, ""),
		Entry("Empty value", map[string]any{"email": ""},
			`expected value of "email" to be between 1 and 200 characters, got 0`),
		Entry("Value above maximum", map[string]any{"email": strings.Repeat("e", 201)},
			`expected value of "email" to be between 1 and 200 characters, got 201`),
		Entry("Not a map", "email", "expected type to be map, got string"),
	)
})

var _ = Describe("HasFailed", func() {
	var diagnostics *diag.Diagnostics

	BeforeEach(func() {
		diagnostics = new(diag.Diagnostics)
	})

	Context("when the error is a service error", func() {
		It("should log a service error and return true", func() {
			err := &indykite.RestError{
				StatusCode: http.StatusInternalServerError,
				Message:    "internal server error",
			}
			Expect(indykite.HasFailed(diagnostics, err)).To(BeTrue())
			Expect(*diagnostics).To(HaveLen(1))
			Expect((*diagnostics)[0].Severity).To(Equal(diag.Error))
			Expect((*diagnostics)[0].Summary).To(Equal("Communication with IndyKite failed, please try again later"))
		})
	})

	Context("when the error is a not-found error", func() {
		It("should log a not-found warning and return true", func() {
			err := &indykite.RestError{
				StatusCode: http.StatusNotFound,
				Message:    "not found",
			}
			Expect(indykite.HasFailed(diagnostics, err)).To(BeTrue())
			Expect(*diagnostics).To(HaveLen(1))
			Expect((*diagnostics)[0].Severity).To(Equal(diag.Warning))
			Expect((*diagnostics)[0].Summary).To(Equal("Resource not found"))
		})
	})

	Context("when the error is of another type", func() {
		It("should log a generic error and return true", func() {
			err := errors.New("generic error")
			Expect(indykite.HasFailed(diagnostics, err)).To(BeTrue())
			Expect(*diagnostics).To(HaveLen(1))
		})
	})

	Context("when the error is nil", func() {
		It("should not log any errors and return false", func() {
			Expect(indykite.HasFailed(diagnostics, nil)).To(BeFalse())
			Expect(*diagnostics).To(BeEmpty())
		})
	})
})
