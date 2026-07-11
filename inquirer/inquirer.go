// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package inquirer ports ActiveSupport::StringInquirer and
// ActiveSupport::ArrayInquirer.
//
// Ruby exposes these through method_missing (e.g. Rails.env.production?); Go has
// no such hook, so the dynamic "<name>?" predicate becomes an explicit Is(name)
// method, and ArrayInquirer's variadic any? becomes Any(candidates...). The
// observable truth values match Rails exactly, including the string/symbol
// equivalence ArrayInquirer applies to its members.
package inquirer

// StringInquirer wraps a string so equality reads as a predicate
// (ActiveSupport::StringInquirer).
type StringInquirer string

// New wraps s in a StringInquirer.
func New(s string) StringInquirer { return StringInquirer(s) }

// Is reports whether the wrapped string equals name (the "<name>?" predicate).
func (s StringInquirer) Is(name string) bool { return string(s) == name }

// String returns the underlying string.
func (s StringInquirer) String() string { return string(s) }

// ArrayInquirer wraps a set of string/symbol-like members so membership reads as
// a predicate (ActiveSupport::ArrayInquirer). Members are stored as strings; the
// string/symbol distinction Ruby collapses on inquiry is preserved by comparing
// stringified forms.
type ArrayInquirer []string

// NewArray wraps items in an ArrayInquirer.
func NewArray(items ...string) ArrayInquirer { return ArrayInquirer(items) }

// Is reports whether name is a member (the "<name>?" predicate, i.e. any?(name)).
func (a ArrayInquirer) Is(name string) bool { return a.Any(name) }

// Any reports whether any candidate is a member. With no candidates it reports
// whether the collection is non-empty, matching Array#any?.
func (a ArrayInquirer) Any(candidates ...string) bool {
	if len(candidates) == 0 {
		return len(a) > 0
	}
	for _, c := range candidates {
		for _, m := range a {
			if m == c {
				return true
			}
		}
	}
	return false
}
