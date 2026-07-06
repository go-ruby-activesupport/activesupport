// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package coreext is a pure-Go (no cgo), MRI-faithful reimplementation of the
// highest-value ActiveSupport core-extension helpers (String, Array, Hash,
// Object/Numeric, Enumerable). Each helper is a plain Go function that the rbgo
// binding maps onto the corresponding Ruby monkey-patch.
//
// Representation choices (documented so the rbgo binding can map cleanly):
//   - Ruby String  → Go string (helpers are rune-aware where Ruby is
//     character-aware).
//   - Ruby Array   → Go []any (nil padding is represented by Go nil).
//   - Ruby Hash    → Go map[any]any so string and Symbol keys can coexist.
//   - Ruby Symbol  → the Symbol type below.
//   - Behaviour that needs Ruby object semantics (Object#try, blank? on an
//     arbitrary object) is reached through an explicit seam (Dispatcher /
//     Blankable) the binding supplies.
//
// String inflection helpers (Pluralize, Camelize, …) delegate to the sibling
// inflector package, so they inherit its byte-for-byte MRI fidelity.
package coreext

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-ruby-activesupport/activesupport/inflector"
)

var reWhitespaceRun = regexp.MustCompile(`\s+`)

// StringBlank reports whether s is empty or contains only whitespace
// (ActiveSupport's String#blank?, matching /\A[[:space:]]*\z/).
func StringBlank(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// StringPresent is the inverse of StringBlank (String#present?).
func StringPresent(s string) bool { return !StringBlank(s) }

// StringPresence returns (s, true) when s is present, or ("", false) when it is
// blank (String#presence returns the string or nil).
func StringPresence(s string) (string, bool) {
	if StringBlank(s) {
		return "", false
	}
	return s, true
}

// Squish strips leading/trailing whitespace and collapses internal whitespace
// runs to a single space (String#squish).
func Squish(s string) string {
	return strings.Join(strings.FieldsFunc(s, unicode.IsSpace), " ")
}

// StripHeredoc removes the smallest common leading indentation from every line
// (String#strip_heredoc).
func StripHeredoc(s string) string {
	lines := strings.Split(s, "\n")
	minIndent := -1
	for _, ln := range lines {
		trimmed := strings.TrimLeft(ln, " \t")
		if trimmed == "" { // blank line: ignored when computing the minimum
			continue
		}
		indent := len(ln) - len(trimmed)
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent <= 0 {
		return s
	}
	for i, ln := range lines {
		if len(ln) >= minIndent {
			lines[i] = ln[minIndent:]
		}
	}
	return strings.Join(lines, "\n")
}

// Truncate shortens s to at most length characters, appending omission (default
// "..." when empty). When separator is non-empty, the cut is moved back to the
// last separator boundary (String#truncate).
func Truncate(s string, length int, omission, separator string) string {
	if omission == "" {
		omission = "..."
	}
	if utf8.RuneCountInString(s) <= length {
		return s
	}
	rs := []rune(s)
	room := length - utf8.RuneCountInString(omission)
	if room < 0 {
		room = 0
	}
	stop := room
	if separator != "" {
		if idx := lastRuneIndexBefore(rs, separator, room); idx >= 0 {
			stop = idx
		}
	}
	return string(rs[:stop]) + omission
}

// lastRuneIndexBefore returns the rune index of the last occurrence of sep that
// ends at or before limit, or -1. It mirrors Ruby's rindex(sep, limit).
func lastRuneIndexBefore(rs []rune, sep string, limit int) int {
	seps := []rune(sep)
	for i := limit; i >= 0; i-- {
		if i+len(seps) <= len(rs) && string(rs[i:i+len(seps)]) == sep {
			return i
		}
	}
	return -1
}

// TruncateWords keeps the first wordsCount whitespace-separated words, appending
// omission (default "...") when the text is longer (String#truncate_words).
func TruncateWords(s string, wordsCount int, omission string) string {
	if omission == "" {
		omission = "..."
	}
	if wordsCount <= 0 {
		return omission
	}
	seps := reWhitespaceRun.FindAllStringIndex(s, -1)
	if len(seps) < wordsCount {
		return s
	}
	return s[:seps[wordsCount-1][0]] + omission
}

// Pluralize returns the plural of s (String#pluralize).
func Pluralize(s string) string { return inflector.Pluralize(s) }

// Singularize returns the singular of s (String#singularize).
func Singularize(s string) string { return inflector.Singularize(s) }

// Titleize titleizes s (String#titleize).
func Titleize(s string) string { return inflector.Titleize(s) }

// Parameterize slugifies s with a "-" separator (String#parameterize).
func Parameterize(s string) string { return inflector.Parameterize(s, "-", false) }

// Camelize converts s to UpperCamelCase (String#camelize).
func Camelize(s string) string { return inflector.Camelize(s) }

// CamelizeLower converts s to lowerCamelCase (String#camelize(:lower)).
func CamelizeLower(s string) string { return inflector.CamelizeLower(s) }

// Underscore converts s to snake_case (String#underscore).
func Underscore(s string) string { return inflector.Underscore(s) }

// Dasherize replaces underscores with dashes (String#dasherize).
func Dasherize(s string) string { return inflector.Dasherize(s) }

// Classify turns a table name into a class name (String#classify).
func Classify(s string) string { return inflector.Classify(s) }

// Tableize turns a class name into a table name (String#tableize).
func Tableize(s string) string { return inflector.Tableize(s) }

// Humanize humanizes s (String#humanize).
func Humanize(s string) string { return inflector.Humanize(s) }

// StartsWith reports whether s begins with prefix (String#starts_with?).
func StartsWith(s, prefix string) bool { return strings.HasPrefix(s, prefix) }

// EndsWith reports whether s ends with suffix (String#ends_with?).
func EndsWith(s, suffix string) bool { return strings.HasSuffix(s, suffix) }

// First returns the first n characters of s (String#first). n defaults to 1 via
// FirstChar; values past the length return the whole string, 0 returns "".
func First(s string, n int) string {
	if n <= 0 {
		return ""
	}
	rs := []rune(s)
	if n >= len(rs) {
		return s
	}
	return string(rs[:n])
}

// FirstChar returns the first character of s (String#first with no argument).
func FirstChar(s string) string { return First(s, 1) }

// Last returns the last n characters of s (String#last).
func Last(s string, n int) string {
	if n <= 0 {
		return ""
	}
	rs := []rune(s)
	if n >= len(rs) {
		return s
	}
	return string(rs[len(rs)-n:])
}

// LastChar returns the last character of s (String#last with no argument).
func LastChar(s string) string { return Last(s, 1) }

// At returns the character at position pos (negative counts from the end),
// reporting false when pos is out of range (String#at / String#[]).
func At(s string, pos int) (string, bool) {
	rs := []rune(s)
	if pos < 0 {
		pos += len(rs)
	}
	if pos < 0 || pos >= len(rs) {
		return "", false
	}
	return string(rs[pos]), true
}

// From returns the substring from position pos to the end (negative counts from
// the end), reporting false when pos is past the end (String#from).
func From(s string, pos int) (string, bool) {
	rs := []rune(s)
	if pos < 0 {
		pos += len(rs)
	}
	if pos < 0 || pos > len(rs) {
		return "", false
	}
	return string(rs[pos:]), true
}

// To returns the substring from the start through position pos inclusive
// (negative counts from the end) (String#to).
func To(s string, pos int) string {
	rs := []rune(s)
	if pos < 0 {
		pos += len(rs)
	}
	if pos < 0 {
		return ""
	}
	if pos >= len(rs) {
		return s
	}
	return string(rs[:pos+1])
}

// Remove deletes every occurrence of each pattern from s (String#remove).
func Remove(s string, patterns ...string) string {
	for _, p := range patterns {
		if p != "" {
			s = strings.ReplaceAll(s, p, "")
		}
	}
	return s
}

// Indent prefixes each non-skipped line of s with amount copies of indentChar
// (default " "). Empty lines are indented only when indentEmptyLines is true
// (String#indent).
func Indent(s string, amount int, indentChar string, indentEmptyLines bool) string {
	if indentChar == "" {
		indentChar = " "
	}
	prefix := strings.Repeat(indentChar, amount)
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if ln == "" && !indentEmptyLines {
			continue
		}
		lines[i] = prefix + ln
	}
	return strings.Join(lines, "\n")
}
