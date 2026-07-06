// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package coreext

import "testing"

type blankThing struct{ blank bool }

func (b blankThing) IsBlank() bool { return b.blank }

func TestBlankPresent(t *testing.T) {
	blanks := []any{nil, false, "", "  ", []any{}, map[any]any{}, blankThing{true}}
	for _, v := range blanks {
		if !Blank(v) {
			t.Errorf("Blank(%#v) = false", v)
		}
		if Present(v) {
			t.Errorf("Present(%#v) = true", v)
		}
	}
	presents := []any{true, "x", []any{1}, map[any]any{"a": 1}, 0, 42, blankThing{false}}
	for _, v := range presents {
		if Blank(v) {
			t.Errorf("Blank(%#v) = true", v)
		}
	}
}

func TestPresence(t *testing.T) {
	if Presence("") != nil {
		t.Error("Presence blank should be nil")
	}
	if Presence("x") != "x" {
		t.Error("Presence present")
	}
}

func TestTry(t *testing.T) {
	dispatch := func(recv any, method string, args []any) (any, bool) {
		if method == "double" {
			return recv.(int) * 2, true
		}
		return nil, false
	}
	if got := Try(5, dispatch, "double"); got != 10 {
		t.Errorf("Try double = %v", got)
	}
	if got := Try(5, dispatch, "unknown"); got != nil {
		t.Errorf("Try unknown = %v", got)
	}
	if got := Try(nil, dispatch, "double"); got != nil {
		t.Errorf("Try nil recv = %v", got)
	}
}

func TestOrdinalNumeric(t *testing.T) {
	if Ordinal(2) != "nd" || Ordinalize(3) != "3rd" {
		t.Error("Ordinal/Ordinalize")
	}
	if !MultipleOf(9, 3) || MultipleOf(9, 2) {
		t.Error("MultipleOf")
	}
	if !MultipleOf(0, 0) || MultipleOf(5, 0) {
		t.Error("MultipleOf zero divisor")
	}
}

func TestToS(t *testing.T) {
	if toS("x") != "x" || toS(Symbol("y")) != "y" || toS(7) != "7" {
		t.Error("toS")
	}
}
