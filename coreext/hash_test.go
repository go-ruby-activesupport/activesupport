// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package coreext

import (
	"reflect"
	"testing"
)

func TestHashBlank(t *testing.T) {
	if !HashBlank(map[any]any{}) || HashBlank(map[any]any{"a": 1}) {
		t.Error("HashBlank")
	}
}

func TestDeepMerge(t *testing.T) {
	h1 := map[any]any{"a": 1, "b": map[any]any{"c": 2, "d": 3}}
	h2 := map[any]any{"b": map[any]any{"d": 4, "e": 5}, "f": 6}
	want := map[any]any{"a": 1, "b": map[any]any{"c": 2, "d": 4, "e": 5}, "f": 6}
	if got := DeepMerge(h1, h2); !reflect.DeepEqual(got, want) {
		t.Errorf("DeepMerge = %v", got)
	}
	// existing value is not a hash ⇒ other wins.
	if got := DeepMerge(map[any]any{"a": 1}, map[any]any{"a": map[any]any{"x": 1}}); !reflect.DeepEqual(got, map[any]any{"a": map[any]any{"x": 1}}) {
		t.Errorf("DeepMerge existing-scalar = %v", got)
	}
	// new value is not a hash while existing is ⇒ other wins.
	if got := DeepMerge(map[any]any{"a": map[any]any{"x": 1}}, map[any]any{"a": 9}); !reflect.DeepEqual(got, map[any]any{"a": 9}) {
		t.Errorf("DeepMerge new-scalar = %v", got)
	}
	// h1/h2 must be untouched by the pure merge.
	if _, ok := h1["f"]; ok {
		t.Error("DeepMerge mutated receiver")
	}

	h := map[any]any{"a": map[any]any{"x": 1}}
	DeepMergeInto(h, map[any]any{"a": map[any]any{"y": 2}, "b": 3})
	if !reflect.DeepEqual(h, map[any]any{"a": map[any]any{"x": 1, "y": 2}, "b": 3}) {
		t.Errorf("DeepMergeInto = %v", h)
	}
}

func TestDeepDup(t *testing.T) {
	orig := map[any]any{"a": map[any]any{"b": 1}, "c": []any{map[any]any{"d": 2}}, "e": 3}
	dup := DeepDup(orig)
	dup["a"].(map[any]any)["b"] = 99
	dup["c"].([]any)[0].(map[any]any)["d"] = 88
	if orig["a"].(map[any]any)["b"] != 1 {
		t.Error("DeepDup nested hash not isolated")
	}
	if orig["c"].([]any)[0].(map[any]any)["d"] != 2 {
		t.Error("DeepDup nested slice not isolated")
	}
	if dup["e"] != 3 {
		t.Error("DeepDup scalar")
	}
}

func TestExceptSlice(t *testing.T) {
	h := map[any]any{"a": 1, "b": 2, "c": 3}
	if got := Except(h, "a", "c"); !reflect.DeepEqual(got, map[any]any{"b": 2}) {
		t.Errorf("Except = %v", got)
	}
	if got := Slice(h, "a", "c", "zz"); !reflect.DeepEqual(got, map[any]any{"a": 1, "c": 3}) {
		t.Errorf("Slice = %v", got)
	}
}

func TestReverseMerge(t *testing.T) {
	got := ReverseMerge(map[any]any{"a": 1, "b": 2}, map[any]any{"b": 3, "c": 4})
	if !reflect.DeepEqual(got, map[any]any{"a": 1, "b": 2, "c": 4}) {
		t.Errorf("ReverseMerge = %v", got)
	}
}

func TestDeepTransformValues(t *testing.T) {
	h := map[any]any{"a": 1, "b": map[any]any{"c": 2}, "d": []any{3, map[any]any{"e": 4}}}
	got := DeepTransformValues(h, func(v any) any { return v.(int) * 10 })
	want := map[any]any{"a": 10, "b": map[any]any{"c": 20}, "d": []any{30, map[any]any{"e": 40}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DeepTransformValues = %v", got)
	}
}

func TestKeyTransforms(t *testing.T) {
	if got := SymbolizeKeys(map[any]any{"a": 1, 2: "x"}); !reflect.DeepEqual(got, map[any]any{Symbol("a"): 1, 2: "x"}) {
		t.Errorf("SymbolizeKeys = %v", got)
	}
	if got := StringifyKeys(map[any]any{Symbol("a"): 1, "b": 2, 3: "z"}); !reflect.DeepEqual(got, map[any]any{"a": 1, "b": 2, "3": "z"}) {
		t.Errorf("StringifyKeys = %v", got)
	}
	// Symbolize keeps an existing Symbol key.
	if got := SymbolizeKeys(map[any]any{Symbol("k"): 1}); !reflect.DeepEqual(got, map[any]any{Symbol("k"): 1}) {
		t.Errorf("SymbolizeKeys sym = %v", got)
	}
	if got := DeepSymbolizeKeys(map[any]any{"a": map[any]any{"b": 1}, "c": []any{map[any]any{"d": 2}}}); !reflect.DeepEqual(
		got, map[any]any{Symbol("a"): map[any]any{Symbol("b"): 1}, Symbol("c"): []any{map[any]any{Symbol("d"): 2}}}) {
		t.Errorf("DeepSymbolizeKeys = %v", got)
	}
	if got := DeepStringifyKeys(map[any]any{Symbol("a"): map[any]any{Symbol("b"): 1}}); !reflect.DeepEqual(
		got, map[any]any{"a": map[any]any{"b": 1}}) {
		t.Errorf("DeepStringifyKeys = %v", got)
	}
	// deepKeyTransform default (scalar leaf) branch.
	if got := DeepSymbolizeKeys(map[any]any{"a": 5}); !reflect.DeepEqual(got, map[any]any{Symbol("a"): 5}) {
		t.Errorf("DeepSymbolizeKeys scalar = %v", got)
	}
}

func TestAssertValidKeys(t *testing.T) {
	if err := AssertValidKeys(map[any]any{"a": 1, "b": 2}, "a", "b", "c"); err != nil {
		t.Errorf("AssertValidKeys valid = %v", err)
	}
	if err := AssertValidKeys(map[any]any{"z": 1}, "a", "b"); err == nil {
		t.Error("AssertValidKeys invalid should error")
	}
}
