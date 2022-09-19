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

// Package test contains helper functions for unit testing.
package test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type matcherWrapper struct {
	matcher types.GomegaMatcher
	// This is used to save variable between calls to Matches and String in case of error
	// to be able to print better messages on failure
	actual interface{}
}

func WrapMatcher(matcher types.GomegaMatcher) gomock.Matcher {
	return &matcherWrapper{matcher: matcher}
}

func (m *matcherWrapper) Matches(x interface{}) (ok bool) {
	m.actual = x
	var err error
	if ok, err = m.matcher.Match(x); err != nil {
		ok = false
	}
	return
}

func (m *matcherWrapper) String() string {
	return fmt.Sprintf("Wrapped Gomega fail message: %s", m.matcher.FailureMessage(m.actual))
}

func WrapTestifyMatcher(matcher types.GomegaMatcher) interface{} {
	return mock.MatchedBy(func(param interface{}) bool {
		result, _ := matcher.Match(param)
		return result
	})
}

func Base58(min ...int) types.GomegaMatcher {
	if len(min) == 1 && min[0] > 0 {
		return gomega.MatchRegexp(fmt.Sprintf("^.[1-9a-zABCDEFGHJKLMNPQRSTUVWXYZ]{%d,}$", min[0]))
	}
	return gomega.MatchRegexp("^.[1-9a-zABCDEFGHJKLMNPQRSTUVWXYZ]$")
}

// BeTemporally compares time.Time's like BeNumerically
// Actual and expected must be time.Time. The comparators are the same as for BeNumerically
//
//	Expect(time.Now()).Should(BeTemporally(">", time.Time{}))
//	Expect(timestamppb.Now()).Should(BeTemporally("~", time.Now(), time.Second))
func BeTemporally(comparator string, compareTo time.Time, threshold ...time.Duration) types.GomegaMatcher {
	return &BeTemporallyMatcher{
		OmegaMatcher: gomega.BeTemporally(comparator, compareTo, threshold...),
	}
}

type BeTemporallyMatcher struct {
	gomega.OmegaMatcher
}

func (matcher *BeTemporallyMatcher) Match(actual interface{}) (bool, error) {
	switch t := actual.(type) {
	case *timestamppb.Timestamp:
		return matcher.OmegaMatcher.Match(t.AsTime())
	default:
		return matcher.OmegaMatcher.Match(actual)
	}
}

type IsJSONMatcher struct{}

func IsJSON() types.GomegaMatcher {
	return &IsJSONMatcher{}
}

func (matcher *IsJSONMatcher) Match(actual interface{}) (success bool, err error) {
	var jsonByteArray []byte
	switch val := actual.(type) {
	case string:
		jsonByteArray = []byte(val)
	case []byte:
		jsonByteArray = val
	default:
		return false, fmt.Errorf("IsJSONString matcher expects a string/[]byte.  Got:\n%s", format.Object(actual, 1))
	}

	var jsonObj interface{}
	return json.Unmarshal(jsonByteArray, &jsonObj) == nil, nil
}

func (matcher *IsJSONMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n%s\nto be JSON", format.Object(actual, 1))
}

func (matcher *IsJSONMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n%s\nnot to be JSON", format.Object(actual, 1))
}

// EqualProto uses proto.Equal to compare actual with expected.  Equal is strict about
// types when performing comparisons.
// It is an error for both actual and expected to be nil.  Use BeNil() instead.
func EqualProto(expected protoreflect.ProtoMessage) types.GomegaMatcher {
	return &EqualProtoMatcher{
		Expected: expected,
	}
}

type EqualProtoMatcher struct {
	Expected proto.Message
}

func (matcher *EqualProtoMatcher) GomegaString() string {
	op := protojson.MarshalOptions{AllowPartial: true, Indent: "  "}
	ex, _ := op.Marshal(matcher.Expected)
	return string(ex)
}

func (matcher *EqualProtoMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil && matcher.Expected == nil {
		// nolint
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead.  This is to avoid mistakes where both sides of an assertion are erroneously uninitialized.")
	}

	if a, ok := actual.(*anypb.Any); ok {
		var pa proto.Message
		pa, err = a.UnmarshalNew()
		if err != nil {
			return false, err
		}
		return proto.Equal(pa, matcher.Expected), nil
	}

	pa, ok := actual.(proto.Message)
	if !ok {
		return false, fmt.Errorf("expected a proto.Message.  Got:\n%s", format.Object(actual, 1))
	}
	return proto.Equal(pa, matcher.Expected), nil
}

func (matcher *EqualProtoMatcher) FailureMessage(actual interface{}) (message string) {
	actualMessage, actualOK := actual.(proto.Message)
	if actualOK {
		op := protojson.MarshalOptions{AllowPartial: true}
		ac := op.Format(actualMessage)
		ex := op.Format(matcher.Expected)
		return format.MessageWithDiff(ac, "to equal", ex)
	}

	return format.Message(actual, "to equal", matcher.Expected)
}

func (matcher *EqualProtoMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	actualMessage, actualOK := actual.(proto.Message)
	if actualOK {
		op := protojson.MarshalOptions{AllowPartial: true}
		ac, _ := op.Marshal(actualMessage)
		ex, _ := op.Marshal(matcher.Expected)
		return format.MessageWithDiff(string(ac), "not to equal", string(ex))
	}
	return format.Message(actual, "not to equal", matcher.Expected)
}
