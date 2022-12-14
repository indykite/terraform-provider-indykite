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
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utilities", func() {
	invalidBase64Error := "expected to be a valid Raw URL Base64 string with 'gid:' prefix, got illegal base64 data"
	DescribeTable("ValidateBase64ID",
		func(input interface{}, errStringMatcher OmegaMatcher) {
			path := cty.IndexIntPath(22)
			d := indykite.ValidateGID(input, path)

			if errStringMatcher == nil {
				Expect(d).To(HaveLen(0))
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
		func(k, currentName, old, new string, expected OmegaMatcher) {
			resourceData := schema.TestResourceDataRaw(GinkgoT(),
				map[string]*schema.Schema{"name": {Type: schema.TypeString}},
				map[string]interface{}{"name": currentName},
			)
			Expect(indykite.DisplayNameDiffSuppress(k, old, new, resourceData)).To(expected)
		},
		Entry("Different key", "description", "a", "b", "c", BeFalse()),
		// If new and old are the same, tested function is not called at all
		// We care only about situation, when new is empty and old is same as name
		Entry("Display name is same as name", "display_name", "abc", "abc", "", BeTrue()),
		Entry("Display name is different than name", "display_name", "abc", "jkl", "", BeFalse()),
	)
})
