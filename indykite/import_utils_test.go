// Copyright (c) 2026 IndyKite
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
	"github.com/indykite/terraform-provider-indykite/indykite"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildReadPath", func() {
	const parentGID = "gid:AAAAAmluZHlraURlgAABDwAAAAA"

	DescribeTable("translates the import location to the Config API query parameter",
		func(resourcePath, id, expected string) {
			Expect(indykite.ReadPathForID(resourcePath, id)).To(Equal(expected))
		},
		Entry("direct ID is used as is",
			"/knowledge-queries", "gid:AAAAFezuHu4AAAAAAAAAAAA",
			"/knowledge-queries/gid:AAAAFezuHu4AAAAAAAAAAAA"),
		Entry("project scoped name uses project_id, not the deprecated location",
			"/knowledge-queries", "wonka-query?location="+parentGID,
			"/knowledge-queries/wonka-query?project_id="+parentGID),
		Entry("application name uses project_id",
			"/applications", "wonka-app?location="+parentGID,
			"/applications/wonka-app?project_id="+parentGID),
		Entry("audit signing name uses project_id",
			"/audit-signings", "wonka-audit?location="+parentGID,
			"/audit-signings/wonka-audit?project_id="+parentGID),
		Entry("mcp server name uses project_id",
			"/mcp-servers", "wonka-mcp?location="+parentGID,
			"/mcp-servers/wonka-mcp?project_id="+parentGID),
		Entry("project name uses organization_id",
			"/projects", "acme?location="+parentGID,
			"/projects/acme?organization_id="+parentGID),
		Entry("service account name uses organization_id",
			"/service-accounts", "my-service-account?location="+parentGID,
			"/service-accounts/my-service-account?organization_id="+parentGID),
	)
})

var _ = Describe("ValidateAliasMapping", func() {
	DescribeTable("accepts valid alias mappings",
		func(value string) {
			warnings, errs := indykite.CheckAliasMapping(value)
			Expect(warnings).To(BeEmpty())
			Expect(errs).To(BeEmpty())
		},
		Entry("only global", "global=db1"),
		Entry("global with other locations", "global=testdb1&east=testdb2&west=testdb3"),
		Entry("global not first", "east=testdb2&global=testdb1"),
	)

	// Rules added by the API later only warn, so configurations accepted before still plan.
	DescribeTable("warns about alias mappings the API rejects on update",
		func(value string, expectedWarnings ...string) {
			warnings, errs := indykite.CheckAliasMapping(value)
			Expect(errs).To(BeEmpty())
			Expect(warnings).To(HaveLen(len(expectedWarnings)))
			for _, expected := range expectedWarnings {
				Expect(warnings).To(ContainElement(ContainSubstring(expected)))
			}
		},
		Entry("missing global", "east=testdb2&west=testdb3", `"alias_mapping" should contain location "global"`),
		Entry("reserved __default", "global=testdb1&__default=testdb2",
			`"alias_mapping" should not contain reserved location "__default"`),
		Entry("reserved __default and missing global", "__default=testdb2",
			`should not contain reserved location "__default"`, `should contain location "global"`),
	)

	DescribeTable("rejects invalid alias mappings",
		func(value string, expectedErrs ...string) {
			_, errs := indykite.CheckAliasMapping(value)
			Expect(errs).To(HaveLen(len(expectedErrs)))
			for _, expected := range expectedErrs {
				Expect(errs).To(ContainElement(MatchError(ContainSubstring(expected))))
			}
		},
		Entry("empty value", "", `"alias_mapping" must not be empty`),
		Entry("duplicate location", "global=testdb1&east=testdb2&east=testdb3",
			`contains location "east" more than once`),
		Entry("empty alias", "global=testdb1&east=", `is missing an alias for location "east"`),
		Entry("empty location", "global=testdb1&=testdb2", "contains an entry with empty location"),
		Entry("malformed query", "global=testdb1&east=%zz", "must be a URL-query-encoded map"),
	)

	It("rejects non string values", func() {
		_, errs := indykite.CheckAliasMapping(42)
		Expect(errs).To(ConsistOf(MatchError(ContainSubstring("to be string"))))
	})
})
