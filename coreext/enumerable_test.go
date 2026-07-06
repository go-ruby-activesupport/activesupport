// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package coreext

import (
	"reflect"
	"testing"
)

func TestIndexBy(t *testing.T) {
	got := IndexBy([]any{1, 2, 3, 4}, func(v any) any { return v.(int) % 2 })
	if !reflect.DeepEqual(got, map[any]any{1: 3, 0: 4}) {
		t.Errorf("IndexBy = %v", got)
	}
}

func TestMany(t *testing.T) {
	if !Many([]any{1, 2}) || Many([]any{1}) {
		t.Error("Many")
	}
	even := func(v any) bool { return v.(int)%2 == 0 }
	if !ManyBy([]any{2, 4, 5}, even) {
		t.Error("ManyBy true")
	}
	if ManyBy([]any{2, 5, 7}, even) {
		t.Error("ManyBy false")
	}
}

func TestExclude(t *testing.T) {
	if !Exclude([]any{1, 2, 3}, 5) || Exclude([]any{1, 2, 3}, 2) {
		t.Error("Exclude")
	}
}

func TestSum(t *testing.T) {
	if got := Sum([]any{1.0, 2.0, 3.0}, 0, nil); got != 6 {
		t.Errorf("Sum = %v", got)
	}
	if got := Sum([]any{1, 2, 3}, 10, func(v any) float64 { return float64(v.(int)) * 2 }); got != 22 {
		t.Errorf("Sum project = %v", got)
	}
}

func TestPluckPick(t *testing.T) {
	items := []any{
		map[any]any{"n": 1},
		map[any]any{"n": 2},
	}
	key := func(v any) any { return v.(map[any]any)["n"] }
	if got := Pluck(items, key); !reflect.DeepEqual(got, []any{1, 2}) {
		t.Errorf("Pluck = %v", got)
	}
	if v, ok := Pick(items, key); !ok || v != 1 {
		t.Errorf("Pick = %v %v", v, ok)
	}
	if _, ok := Pick(nil, key); ok {
		t.Error("Pick empty should be false")
	}
}
