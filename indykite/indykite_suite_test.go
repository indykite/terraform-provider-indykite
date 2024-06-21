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

	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	customerID     = "gid:AAAAAWluZHlraURlgAAAAAAAAA8"
	appSpaceID     = "gid:AAAAAmluZHlraURlgAABDwAAAAA"
	issuerID       = "gid:AAAAD2luZHlraURlgAAEDwAAAAA"
	applicationID  = "gid:AAAABGluZHlraURlgAACDwAAAAA"
	appAgentID     = "gid:AAAABWluZHlraURlgAAFDwAAAAA"
	appAgentCredID = "gid:AAAABt7z4hZzpkbAtZXbIEYsT9Q" // #nosec G101
	// sampleID is plain empty ID just for responses.
	sampleID  = "gid:AAAAAAAAAAAAAAAAAAAAAAAAAAA"
	sampleID2 = "gid:AAAAAAAAAAAAAAAAAAAAAAAAAAB"
	sampleID3 = "gid:AAAAAAAAAAAAAAAAAAAAAAAAAAC"
)

type terraformGomockTestReporter struct {
	ginkgoT GinkgoTInterface
}

type GomockTestCleanuper interface {
	gomock.TestHelper
	Cleanup(func())
}

func TestIndykite(t *testing.T) {
	RegisterFailHandler(Fail)
	_ = os.Setenv("TF_ACC", "ok")
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

func addStringArrayToKeys[T string | []byte](keys Keys, key string, value []T) {
	if len(value) == 0 {
		return
	}
	keys[key+".#"] = Equal(strconv.Itoa(len(value)))
	for i, v := range value {
		keys[key+"."+strconv.Itoa(i)] = Equal(string(v))
	}
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

func addStringMapMatcherToKeys(keys Keys, key string, data map[string]string, includeEmpty bool) {
	if len(data) == 0 && !includeEmpty {
		return
	}

	keys[key+".%"] = Equal(strconv.Itoa(len(data)))

	for k, v := range data {
		keys[key+"."+k] = Equal(v)
	}
}
