// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package hwia

import (
	"reflect"
	"testing"

	"github.com/go-ruby-activesupport/activesupport/coreext"
)

func TestIndifferentAccess(t *testing.T) {
	h := New()
	h.Set("a", 1)
	if h.Get("a") != 1 {
		t.Errorf("Get(a) = %v", h.Get("a"))
	}
	// Overwrite keeps position.
	h.Set("a", 2)
	if h.Get("a") != 2 || h.Len() != 1 {
		t.Errorf("overwrite: %v len %d", h.Get("a"), h.Len())
	}
}

func TestNestedConversion(t *testing.T) {
	h := NewFrom(map[string]any{"b": map[string]any{"c": 2}})
	nested, ok := h.Get("b").(*Hash)
	if !ok {
		t.Fatalf("nested not *Hash: %T", h.Get("b"))
	}
	if nested.Get("c") != 2 {
		t.Errorf("nested c = %v", nested.Get("c"))
	}
	// map[any]any with symbol keys and array of hashes.
	h2 := New()
	h2.Set("x", map[any]any{coreext.Symbol("y"): []any{map[string]any{"z": 1}}})
	arr := h2.Get("x").(*Hash).Get("y").([]any)
	if arr[0].(*Hash).Get("z") != 1 {
		t.Errorf("deep nested = %v", arr[0])
	}
	// map[any]any duplicate-normalised key path: string + symbol collapse.
	h3 := New()
	h3.Set("m", map[any]any{"k": 1})
	if h3.Get("m").(*Hash).Get("k") != 1 {
		t.Error("map[any]any string key")
	}
	// Unsupported key type is dropped to "".
	h4 := New()
	h4.Set("m", map[any]any{42: 1})
	if !h4.Get("m").(*Hash).KeyQ("") {
		t.Error("unsupported key should map to empty string")
	}
	// Setting an already-converted *Hash passes it through unchanged.
	inner := NewFrom(map[string]any{"p": 9})
	h5 := New()
	h5.Set("n", inner)
	if h5.Get("n").(*Hash).Get("p") != 9 {
		t.Error("existing *Hash should pass through")
	}
	// map[any]any with a repeated key (string then symbol) keeps one entry.
	h6 := New()
	h6.Set("d", map[any]any{"k": 1})
	h6.Get("d").(*Hash).Set("k", 2) // exercise overwrite via converted hash
	if h6.Get("d").(*Hash).Get("k") != 2 {
		t.Error("converted hash overwrite")
	}
}

func TestFetchDeleteKeyQ(t *testing.T) {
	h := NewFrom(map[string]any{"a": 1})
	if v, ok := h.Fetch("a"); !ok || v != 1 {
		t.Error("fetch present")
	}
	if _, ok := h.Fetch("z"); ok {
		t.Error("fetch absent")
	}
	if h.FetchDefault("z", 9) != 9 {
		t.Error("fetch default")
	}
	if h.FetchDefault("a", 9) != 1 {
		t.Error("fetch default present")
	}
	if !h.KeyQ("a") || h.KeyQ("z") {
		t.Error("keyq")
	}
	if v, ok := h.Delete("a"); !ok || v != 1 {
		t.Error("delete present")
	}
	if _, ok := h.Delete("a"); ok {
		t.Error("delete absent")
	}
}

func TestOrderAndIteration(t *testing.T) {
	h := New()
	h.Set("a", 1)
	h.Set("b", 2)
	h.Set("c", 3)
	if !reflect.DeepEqual(h.Keys(), []string{"a", "b", "c"}) {
		t.Errorf("keys = %v", h.Keys())
	}
	if !reflect.DeepEqual(h.Values(), []any{1, 2, 3}) {
		t.Errorf("values = %v", h.Values())
	}
	var seen []string
	h.Each(func(k string, _ any) { seen = append(seen, k) })
	if !reflect.DeepEqual(seen, []string{"a", "b", "c"}) {
		t.Errorf("each = %v", seen)
	}
	h.Delete("b")
	if !reflect.DeepEqual(h.Keys(), []string{"a", "c"}) {
		t.Errorf("after delete = %v", h.Keys())
	}
}

func TestMergeUpdateDupSliceExcept(t *testing.T) {
	h := NewFrom(map[string]any{"a": 1, "b": 2})
	m := h.Merge(NewFrom(map[string]any{"b": 20, "c": 3}))
	if m.Get("b") != 20 || m.Get("c") != 3 || h.Get("b") != 2 {
		t.Errorf("merge = %v (orig b %v)", m.ToHash(), h.Get("b"))
	}
	d := h.Dup()
	d.Set("a", 99)
	if h.Get("a") != 1 {
		t.Error("dup should not alias")
	}
	h.Update(NewFrom(map[string]any{"a": 5}))
	if h.Get("a") != 5 {
		t.Error("update in place")
	}
	sl := h.Slice("a", "z")
	if sl.Len() != 1 || sl.Get("a") != 5 {
		t.Errorf("slice = %v", sl.ToHash())
	}
	ex := h.Except("a")
	if ex.KeyQ("a") || !ex.KeyQ("b") {
		t.Errorf("except = %v", ex.ToHash())
	}
	va := h.ValuesAt("a", "z", "b")
	if !reflect.DeepEqual(va, []any{5, nil, 2}) {
		t.Errorf("values_at = %v", va)
	}
}

func TestToHashDeep(t *testing.T) {
	h := NewFrom(map[string]any{"x": map[string]any{"y": []any{map[string]any{"z": 1}}}})
	plain := h.ToHash()
	inner := plain["x"].(map[string]any)["y"].([]any)[0].(map[string]any)
	if inner["z"] != 1 {
		t.Errorf("to_hash deep = %v", plain)
	}
}

func TestOrderedOptions(t *testing.T) {
	o := NewOrderedOptions()
	o.Set("boy", "John")
	if o.Get("boy") != "John" {
		t.Errorf("get = %v", o.Get("boy"))
	}
	if o.Get("dog") != nil {
		t.Error("absent should be nil")
	}
	if _, err := o.GetBang("boy"); err != nil {
		t.Errorf("bang present: %v", err)
	}
	if _, err := o.GetBang("dog"); err == nil || err.Error() != ":dog is blank" {
		t.Errorf("bang blank err = %v", err)
	}
	o.Set("empty", "")
	if _, err := o.GetBang("empty"); err == nil {
		t.Error("blank string should error")
	}
	if !o.KeyQ("boy") {
		t.Error("keyq")
	}
	o.Set("x", 1)
	if !reflect.DeepEqual(o.Keys(), []string{"boy", "empty", "x"}) {
		t.Errorf("keys = %v", o.Keys())
	}
	if _, ok := o.Delete("x"); !ok {
		t.Error("delete")
	}
}
