// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package coreext

// The Enumerable helpers operate on []any with explicit projection functions
// standing in for the Ruby blocks / symbol-to-proc arguments.

// IndexBy builds a map from key(elem) to the last element with that key
// (Enumerable#index_by).
func IndexBy(items []any, key func(any) any) map[any]any {
	out := make(map[any]any, len(items))
	for _, it := range items {
		out[key(it)] = it
	}
	return out
}

// Many reports whether items has more than one element (Enumerable#many?).
func Many(items []any) bool { return len(items) > 1 }

// ManyBy reports whether more than one element satisfies pred
// (Enumerable#many? with a block).
func ManyBy(items []any, pred func(any) bool) bool {
	count := 0
	for _, it := range items {
		if pred(it) {
			count++
			if count > 1 {
				return true
			}
		}
	}
	return false
}

// Exclude reports whether v is not present in items (Enumerable#exclude?).
func Exclude(items []any, v any) bool {
	for _, it := range items {
		if it == v {
			return false
		}
	}
	return true
}

// Sum adds init to the sum of project(elem) over items (Enumerable#sum). Pass a
// nil project to sum the elements themselves when they are float64.
func Sum(items []any, init float64, project func(any) float64) float64 {
	total := init
	for _, it := range items {
		if project != nil {
			total += project(it)
		} else {
			total += it.(float64)
		}
	}
	return total
}

// Pluck returns key(elem) for every element (Enumerable#pluck).
func Pluck(items []any, key func(any) any) []any {
	out := make([]any, len(items))
	for i, it := range items {
		out[i] = key(it)
	}
	return out
}

// Pick returns key(elem) for the first element, reporting false when items is
// empty (Enumerable#pick).
func Pick(items []any, key func(any) any) (any, bool) {
	if len(items) == 0 {
		return nil, false
	}
	return key(items[0]), true
}
