// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package coreext

import (
	"fmt"

	"github.com/go-ruby-activesupport/activesupport/inflector"
)

// Blankable is the seam for objects that define their own blank? via an
// empty?-like predicate. The rbgo binding implements it for Ruby objects that
// respond to empty?.
type Blankable interface{ IsBlank() bool }

// Dispatcher is the method-dispatch seam used by Try: it invokes method on recv
// and reports whether recv responds to it. The rbgo binding backs this with the
// Ruby object model.
type Dispatcher func(recv any, method string, args []any) (result any, responded bool)

// Blank reports whether v is "blank" in the ActiveSupport sense: nil, false,
// an empty/whitespace string, an empty array or hash, or a Blankable that says
// so. Everything else (including any number) is present (Object#blank?).
func Blank(v any) bool {
	switch t := v.(type) {
	case nil:
		return true
	case bool:
		return !t
	case string:
		return StringBlank(t)
	case []any:
		return len(t) == 0
	case map[any]any:
		return len(t) == 0
	case Blankable:
		return t.IsBlank()
	default:
		return false
	}
}

// Present is the inverse of Blank (Object#present?).
func Present(v any) bool { return !Blank(v) }

// Presence returns v when it is present, or nil when it is blank
// (Object#presence).
func Presence(v any) any {
	if Blank(v) {
		return nil
	}
	return v
}

// Try invokes method on recv through dispatch, returning nil when recv is nil or
// does not respond to method (Object#try). This is the safe-navigation helper;
// dispatch is the Ruby-semantics seam.
func Try(recv any, dispatch Dispatcher, method string, args ...any) any {
	if recv == nil {
		return nil
	}
	if result, ok := dispatch(recv, method, args); ok {
		return result
	}
	return nil
}

// Ordinal returns the ordinal suffix for n (Integer#ordinal).
func Ordinal(n int) string { return inflector.Ordinal(n) }

// Ordinalize renders n with its ordinal suffix (Integer#ordinalize).
func Ordinalize(n int) string { return inflector.Ordinalize(n) }

// MultipleOf reports whether n is an integer multiple of m
// (Integer#multiple_of?). When m is zero, only zero is a multiple.
func MultipleOf(n, m int) bool {
	if m == 0 {
		return n == 0
	}
	return n%m == 0
}

// toS renders a value the way Ruby's to_s would for the common cases used by
// Array#to_sentence.
func toS(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case Symbol:
		return string(t)
	default:
		return fmt.Sprintf("%v", v)
	}
}
