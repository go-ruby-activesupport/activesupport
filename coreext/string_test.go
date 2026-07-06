// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package coreext

import "testing"

func TestStringBlankPresent(t *testing.T) {
	for _, s := range []string{"", "   ", "\t\n"} {
		if !StringBlank(s) {
			t.Errorf("StringBlank(%q) = false", s)
		}
		if StringPresent(s) {
			t.Errorf("StringPresent(%q) = true", s)
		}
		if _, ok := StringPresence(s); ok {
			t.Errorf("StringPresence(%q) present", s)
		}
	}
	if StringBlank("x") {
		t.Error("StringBlank(x) = true")
	}
	if v, ok := StringPresence(" x "); !ok || v != " x " {
		t.Errorf("StringPresence = %q, %v", v, ok)
	}
}

func TestSquish(t *testing.T) {
	if got := Squish("  foo\n\t bar   baz  "); got != "foo bar baz" {
		t.Errorf("Squish = %q", got)
	}
}

func TestStripHeredoc(t *testing.T) {
	if got := StripHeredoc("    line1\n      line2\n    line3\n"); got != "line1\n  line2\nline3\n" {
		t.Errorf("StripHeredoc = %q", got)
	}
	// No indentation ⇒ unchanged (minIndent <= 0 path).
	if got := StripHeredoc("a\nb"); got != "a\nb" {
		t.Errorf("StripHeredoc no-indent = %q", got)
	}
	// A line shorter than the common indent (a truly empty line) is preserved.
	if got := StripHeredoc("  a\n\n  b"); got != "a\n\nb" {
		t.Errorf("StripHeredoc short-line = %q", got)
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		s             string
		length        int
		om, sep, want string
	}{
		{"Once upon a time in a world far far away", 27, "", "", "Once upon a time in a wo..."},
		{"Once upon a time in a world far far away", 27, "", " ", "Once upon a time in a..."},
		{"Once upon a time", 10, "…", " ", "Once upon…"},
		{"short", 20, "", "", "short"},
		{"hello world", 2, "", "", "..."},    // room < 0 ⇒ omission only
		{"abcdefghij", 6, "", "-", "abc..."}, // separator absent before cut ⇒ plain cut
	}
	for _, c := range cases {
		if got := Truncate(c.s, c.length, c.om, c.sep); got != c.want {
			t.Errorf("Truncate(%q,%d,%q,%q) = %q, want %q", c.s, c.length, c.om, c.sep, got, c.want)
		}
	}
}

func TestTruncateWords(t *testing.T) {
	if got := TruncateWords("Once upon a time in a world far far away", 4, ""); got != "Once upon a time..." {
		t.Errorf("TruncateWords = %q", got)
	}
	if got := TruncateWords("Once upon a time", 2, "... (continued)"); got != "Once upon... (continued)" {
		t.Errorf("TruncateWords om = %q", got)
	}
	if got := TruncateWords("only three words here", 10, ""); got != "only three words here" {
		t.Errorf("TruncateWords fewer = %q", got)
	}
	if got := TruncateWords("anything", 0, "X"); got != "X" {
		t.Errorf("TruncateWords zero = %q", got)
	}
}

func TestStringInflectionDelegation(t *testing.T) {
	checks := []struct {
		name string
		got  string
		want string
	}{
		{"Pluralize", Pluralize("post"), "posts"},
		{"Singularize", Singularize("posts"), "post"},
		{"Titleize", Titleize("man from the boondocks"), "Man From The Boondocks"},
		{"Parameterize", Parameterize("Donald E. Knuth"), "donald-e-knuth"},
		{"Camelize", Camelize("active_model"), "ActiveModel"},
		{"CamelizeLower", CamelizeLower("active_model"), "activeModel"},
		{"Underscore", Underscore("ActiveModel"), "active_model"},
		{"Dasherize", Dasherize("puni_puni"), "puni-puni"},
		{"Classify", Classify("posts"), "Post"},
		{"Tableize", Tableize("RawScaledScorer"), "raw_scaled_scorers"},
		{"Humanize", Humanize("employee_salary"), "Employee salary"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestStartsEndsWith(t *testing.T) {
	if !StartsWith("hello", "he") || StartsWith("hello", "lo") {
		t.Error("StartsWith")
	}
	if !EndsWith("hello", "lo") || EndsWith("hello", "he") {
		t.Error("EndsWith")
	}
}

func TestFirstLast(t *testing.T) {
	if First("hello", 3) != "hel" || First("hello", 0) != "" || First("hi", 9) != "hi" {
		t.Error("First")
	}
	if FirstChar("hello") != "h" {
		t.Error("FirstChar")
	}
	if Last("hello", 3) != "llo" || Last("hello", 0) != "" || Last("hi", 9) != "hi" {
		t.Error("Last")
	}
	if LastChar("hello") != "o" {
		t.Error("LastChar")
	}
}

func TestAtFromTo(t *testing.T) {
	if v, ok := At("hello", 1); !ok || v != "e" {
		t.Errorf("At 1 = %q %v", v, ok)
	}
	if v, ok := At("hello", -1); !ok || v != "o" {
		t.Errorf("At -1 = %q %v", v, ok)
	}
	if _, ok := At("hello", 99); ok {
		t.Error("At 99 should be OOB")
	}
	if _, ok := At("hello", -99); ok {
		t.Error("At -99 should be OOB")
	}
	if v, ok := From("hello", 2); !ok || v != "llo" {
		t.Errorf("From 2 = %q %v", v, ok)
	}
	if v, ok := From("hello", -2); !ok || v != "lo" {
		t.Errorf("From -2 = %q %v", v, ok)
	}
	if v, ok := From("hello", 5); !ok || v != "" {
		t.Errorf("From 5 = %q %v", v, ok)
	}
	if _, ok := From("hello", 99); ok {
		t.Error("From 99 should be OOB")
	}
	if _, ok := From("hello", -99); ok {
		t.Error("From -99 should be OOB")
	}
	if To("hello", 2) != "hel" || To("hello", -2) != "hell" || To("hello", 99) != "hello" || To("hello", -99) != "" {
		t.Error("To")
	}
}

func TestRemove(t *testing.T) {
	if got := Remove("foo bar foo", "foo "); got != "bar foo" {
		t.Errorf("Remove = %q", got)
	}
	if got := Remove("keepme", ""); got != "keepme" { // empty pattern skipped
		t.Errorf("Remove empty = %q", got)
	}
	if got := Remove("a1b2c3", "1", "3"); got != "ab2c" {
		t.Errorf("Remove multi = %q", got)
	}
}

func TestIndent(t *testing.T) {
	if got := Indent("foo\n  bar", 2, "", false); got != "  foo\n    bar" {
		t.Errorf("Indent = %q", got)
	}
	if got := Indent("foo\n\nbar", 2, "*", true); got != "**foo\n**\n**bar" {
		t.Errorf("Indent empty-lines = %q", got)
	}
	if got := Indent("foo\n\nbar", 2, " ", false); got != "  foo\n\n  bar" {
		t.Errorf("Indent skip-empty = %q", got)
	}
}
