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
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/onsi/gomega/matchers"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	customerID      = "gid:AAAAAWluZHlraURlgAAAAAAAAA8"
	appSpaceID      = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
	applicationID   = "gid:AAAABGluZHlraURlgAACDwAAAAA"
	appAgentID      = "gid:AAAABWluZHlraURlgAAFDwAAAAA"
	appAgentCredID  = "gid:AAAABt7z4hZzpkbAtZXbIEYsT9Q" // #nosec G101
	appAgentCredID2 = "gid:AAAABkMsC87ROQ8mlK-Q6PSoTuw" // #nosec G101
	// sampleID is plain empty ID just for responses.
	sampleID  = "gid:AAAAAAAAAAAAAAAAAAAAAAAAAAA"
	sampleID2 = "gid:AAAAAAAAAAAAAAAAAAAAAAAAAAB"
)

type terraformGomockTestReporter struct {
	ginkgoT GinkgoTInterface
}

type GomockTestCleanuper interface {
	gomock.TestHelper
	Cleanup(fn func())
}

func TestIndykite(t *testing.T) {
	RegisterFailHandler(Fail)
	t.Setenv("TF_ACC", "ok")
	RunSpecs(t, "Indykite Suite")
}

// TerraformGomockT should be used inside gomock.NewController instead of pure GinkgoT or testing.T.
//
// This is not the best solution, but Terraform execute provider as s separate program.
// So os.Stderr and os.Stdout are only way how to easily communicate error out.
// And as this is required currently only for Gomock, it is enough to support gomock.TestReporter.
func TerraformGomockT(ginkgoT GinkgoTInterface) GomockTestCleanuper {
	return terraformGomockTestReporter{
		ginkgoT: ginkgoT,
	}
}

func (terraformGomockTestReporter) Errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	panic("Error, see stderr")
}

func (terraformGomockTestReporter) Fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	panic("Fatal, see stderr")
}

func (t terraformGomockTestReporter) Helper() {
	t.ginkgoT.Helper()
}

func (t terraformGomockTestReporter) Cleanup(callback func()) {
	t.ginkgoT.Cleanup(callback)
}

func JSONEquals(oldValue, newValue string) bool {
	match, err := MatchJSON(newValue).Match(oldValue)
	if err != nil {
		return false
	}
	return match
}

func convertOmegaMatcherToError(matcher OmegaMatcher, actual any) error {
	success, err := matcher.Match(actual)
	if err != nil {
		return err
	}
	if !success {
		return errors.New(matcher.FailureMessage(actual))
	}

	return nil
}

type NumericalTerraformMatcher struct {
	matchers.BeNumericallyMatcher
}

// BeTerraformNumerically performs numerical assertions in a type-agnostic way.
// Expected should be numbers, though the specific type of number is irrelevant (float32, float64, uint8, etc...).
// Actual has same rules as Expected, but can be also string, which is returned by Terraform.
//
// There are six, self-explanatory, supported comparators:
//
//	Expect(1.0).Should(BeTerraformNumerically("==", 1))
//	Expect(1.0).Should(BeTerraformNumerically("~", 0.999, 0.01))
//	Expect(1.0).Should(BeTerraformNumerically(">", 0.9))
//	Expect(1.0).Should(BeTerraformNumerically(">=", 1.0))
//	Expect(1.0).Should(BeTerraformNumerically("<", 3))
//	Expect(1.0).Should(BeTerraformNumerically("<=", 1.0))
func BeTerraformNumerically(comparator string, compareTo ...any) OmegaMatcher {
	if len(compareTo) == 1 {
		compareTo = append(compareTo, 1e-6)
	}
	return &NumericalTerraformMatcher{
		BeNumericallyMatcher: matchers.BeNumericallyMatcher{Comparator: comparator, CompareTo: compareTo},
	}
}

func (matcher *NumericalTerraformMatcher) Match(actual any) (bool, error) {
	if s, isStr := actual.(string); isStr {
		var err error
		actual, err = strconv.ParseFloat(s, 64)
		if err != nil {
			return false, err
		}
	}
	return matcher.BeNumericallyMatcher.Match(actual)
}
