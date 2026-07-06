// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package coreext

import (
	"reflect"
	"testing"
)

func TestArrayBlank(t *testing.T) {
	if !ArrayBlank(nil) || !ArrayBlank([]any{}) {
		t.Error("ArrayBlank empty")
	}
	if ArrayBlank([]any{1}) {
		t.Error("ArrayBlank non-empty")
	}
}

func TestArrayFromTo(t *testing.T) {
	a := []any{1, 2, 3, 4, 5}
	if v, ok := ArrayFrom(a, 2); !ok || !reflect.DeepEqual(v, []any{3, 4, 5}) {
		t.Errorf("ArrayFrom 2 = %v %v", v, ok)
	}
	if v, ok := ArrayFrom(a, -2); !ok || !reflect.DeepEqual(v, []any{4, 5}) {
		t.Errorf("ArrayFrom -2 = %v %v", v, ok)
	}
	if v, ok := ArrayFrom(a, 5); !ok || !reflect.DeepEqual(v, []any{}) {
		t.Errorf("ArrayFrom 5 = %v %v", v, ok)
	}
	if _, ok := ArrayFrom(a, 99); ok {
		t.Error("ArrayFrom 99 should be OOB")
	}
	if _, ok := ArrayFrom(a, -99); ok {
		t.Error("ArrayFrom -99 should be OOB")
	}
	if !reflect.DeepEqual(ArrayTo(a, 2), []any{1, 2, 3}) {
		t.Error("ArrayTo 2")
	}
	if !reflect.DeepEqual(ArrayTo(a, -2), []any{1, 2, 3, 4}) {
		t.Error("ArrayTo -2")
	}
	if !reflect.DeepEqual(ArrayTo(a, 99), a) {
		t.Error("ArrayTo 99")
	}
	if !reflect.DeepEqual(ArrayTo(a, -99), []any{}) {
		t.Error("ArrayTo -99")
	}
}

func TestNthAccessors(t *testing.T) {
	a := []any{1, 2, 3, 4, 5}
	if Second(a) != 2 || Third(a) != 3 || Fourth(a) != 4 || Fifth(a) != 5 {
		t.Error("nth accessors")
	}
	if Fifth([]any{1}) != nil {
		t.Error("Fifth OOB should be nil")
	}
}

func TestInGroups(t *testing.T) {
	a := []any{1, 2, 3, 4, 5, 6, 7}
	if !reflect.DeepEqual(InGroups(a, 3, nil),
		[][]any{{1, 2, 3}, {4, 5, nil}, {6, 7, nil}}) {
		t.Errorf("InGroups = %v", InGroups(a, 3, nil))
	}
	if !reflect.DeepEqual(InGroupsNoFill(a, 3),
		[][]any{{1, 2, 3}, {4, 5}, {6, 7}}) {
		t.Errorf("InGroupsNoFill = %v", InGroupsNoFill(a, 3))
	}
	// Evenly divisible ⇒ modulo 0, no fill branch.
	if !reflect.DeepEqual(InGroups([]any{1, 2, 3, 4}, 2, nil),
		[][]any{{1, 2}, {3, 4}}) {
		t.Errorf("InGroups even = %v", InGroups([]any{1, 2, 3, 4}, 2, nil))
	}
}

func TestInGroupsOf(t *testing.T) {
	a := []any{1, 2, 3, 4, 5, 6, 7}
	if !reflect.DeepEqual(InGroupsOf(a, 3, nil),
		[][]any{{1, 2, 3}, {4, 5, 6}, {7, nil, nil}}) {
		t.Errorf("InGroupsOf = %v", InGroupsOf(a, 3, nil))
	}
	if !reflect.DeepEqual(InGroupsOfNoFill(a, 3),
		[][]any{{1, 2, 3}, {4, 5, 6}, {7}}) {
		t.Errorf("InGroupsOfNoFill = %v", InGroupsOfNoFill(a, 3))
	}
}

func TestSplit(t *testing.T) {
	if !reflect.DeepEqual(Split([]any{1, 2, 0, 3, 4, 0, 5}, 0),
		[][]any{{1, 2}, {3, 4}, {5}}) {
		t.Errorf("Split = %v", Split([]any{1, 2, 0, 3, 4, 0, 5}, 0))
	}
}

func TestExtractOptions(t *testing.T) {
	opts := map[any]any{"a": 1}
	rest, got := ExtractOptions([]any{1, 2, opts})
	if !reflect.DeepEqual(rest, []any{1, 2}) || !reflect.DeepEqual(got, opts) {
		t.Errorf("ExtractOptions = %v %v", rest, got)
	}
	rest, got = ExtractOptions([]any{1, 2})
	if !reflect.DeepEqual(rest, []any{1, 2}) || len(got) != 0 {
		t.Errorf("ExtractOptions no-opts = %v %v", rest, got)
	}
	rest, got = ExtractOptions(nil)
	if rest != nil || len(got) != 0 {
		t.Errorf("ExtractOptions empty = %v %v", rest, got)
	}
}

func TestToSentence(t *testing.T) {
	if ToSentence(nil) != "" {
		t.Error("ToSentence empty")
	}
	if ToSentence([]any{"a"}) != "a" {
		t.Error("ToSentence one")
	}
	if ToSentence([]any{"a", "b"}) != "a and b" {
		t.Error("ToSentence two")
	}
	if ToSentence([]any{"a", "b", "c"}) != "a, b, and c" {
		t.Error("ToSentence three")
	}
	if ToSentenceWith([]any{"a", "b", "c"}, "; ", " & ", " or ") != "a; b or c" {
		t.Error("ToSentenceWith custom")
	}
	// Non-string element uses to_s.
	if ToSentence([]any{1, 2}) != "1 and 2" {
		t.Errorf("ToSentence ints = %q", ToSentence([]any{1, 2}))
	}
}
