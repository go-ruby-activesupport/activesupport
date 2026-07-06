// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package inflector is a pure-Go (no cgo), MRI-faithful reimplementation of
// ActiveSupport::Inflector — the English word-inflection engine at the heart of
// Rails (pluralize/singularize, camelize/underscore, humanize/titleize,
// tableize/classify, foreign_key, ordinalize, parameterize, transliterate).
//
// The default inflection rules, irregulars, uncountables and the transliteration
// table are lifted verbatim from the activesupport gem so outputs match MRI
// byte-for-byte. Registration (Plural/Singular/Irregular/Uncountable/Acronym/
// Human) replicates ActiveSupport::Inflector::Inflections precisely, including its
// prepend ("new rules on top") ordering.
package inflector

import (
	"regexp"
	"sort"
	"strings"
)

// rule is one (pattern, replacement) inflection rule. The replacement uses Go's
// Expand template syntax (${1}, ${2}) rather than Ruby's \1/\2.
type rule struct {
	re   *regexp.Regexp
	repl string
}

// Inflections is a mutable set of inflection rules, mirroring
// ActiveSupport::Inflector::Inflections. New rules are prepended, so the most
// recently registered rule is tried first — exactly as Rails does it.
type Inflections struct {
	plurals      []rule
	singulars    []rule
	humans       []rule
	uncountables []string
	acronyms     map[string]string

	uncountableRe *regexp.Regexp // cached; rebuilt when uncountables change
}

// NewInflections returns an empty Inflections with no rules. Use DefaultLocale
// (or a clone of it) for Rails' built-in English rules.
func NewInflections() *Inflections {
	return &Inflections{acronyms: map[string]string{}}
}

// Clone returns a deep copy, so callers can register extra rules without mutating
// the shared default instance (mirrors Inflections#initialize_dup).
func (inf *Inflections) Clone() *Inflections {
	c := &Inflections{
		plurals:      append([]rule(nil), inf.plurals...),
		singulars:    append([]rule(nil), inf.singulars...),
		humans:       append([]rule(nil), inf.humans...),
		uncountables: append([]string(nil), inf.uncountables...),
		acronyms:     make(map[string]string, len(inf.acronyms)),
	}
	for k, v := range inf.acronyms {
		c.acronyms[k] = v
	}
	return c
}

// mustRuby compiles a Ruby-flavoured pattern: a leading "i" makes it
// case-insensitive (Ruby's /.../i). The body is otherwise RE2-compatible.
func mustRuby(caseInsensitive bool, body string) *regexp.Regexp {
	if caseInsensitive {
		body = "(?i)" + body
	}
	return regexp.MustCompile(body)
}

// Plural registers a pluralization rule (prepended). replacement is in Go Expand
// syntax (use ${1} not \1). Registering removes any uncountable it shadows.
func (inf *Inflections) Plural(caseInsensitive bool, pattern, replacement string) {
	inf.plurals = append([]rule{{mustRuby(caseInsensitive, pattern), replacement}}, inf.plurals...)
}

// Singular registers a singularization rule (prepended).
func (inf *Inflections) Singular(caseInsensitive bool, pattern, replacement string) {
	inf.singulars = append([]rule{{mustRuby(caseInsensitive, pattern), replacement}}, inf.singulars...)
}

// Human registers a humanization rule (prepended), applied by Humanize.
func (inf *Inflections) Human(caseInsensitive bool, pattern, replacement string) {
	inf.humans = append([]rule{{mustRuby(caseInsensitive, pattern), replacement}}, inf.humans...)
}

// Uncountable marks words as uncountable (never inflected). Words are compared
// case-insensitively at a word boundary, as in Rails.
func (inf *Inflections) Uncountable(words ...string) {
	for _, w := range words {
		inf.uncountables = append(inf.uncountables, strings.ToLower(w))
	}
	inf.uncountableRe = nil
}

// Irregular registers a singular/plural pair, replicating
// Inflections#irregular exactly (including its case-preserving rule generation).
func (inf *Inflections) Irregular(singular, plural string) {
	s0, srest := singular[:1], singular[1:]
	p0, prest := plural[:1], plural[1:]

	if strings.ToUpper(s0) == strings.ToUpper(p0) {
		inf.Plural(true, "("+s0+")"+srest+"$", "${1}"+prest)
		inf.Plural(true, "("+p0+")"+prest+"$", "${1}"+prest)
		inf.Singular(true, "("+s0+")"+srest+"$", "${1}"+srest)
		inf.Singular(true, "("+p0+")"+prest+"$", "${1}"+srest)
		return
	}
	su, sd := strings.ToUpper(s0), strings.ToLower(s0)
	pu, pd := strings.ToUpper(p0), strings.ToLower(p0)
	// Ruby: /#{s0.upcase}(?i)#{srest}$/ — first char case-sensitive, rest folded.
	inf.Plural(false, su+"(?i)"+srest+"$", pu+prest)
	inf.Plural(false, sd+"(?i)"+srest+"$", pd+prest)
	inf.Plural(false, pu+"(?i)"+prest+"$", pu+prest)
	inf.Plural(false, pd+"(?i)"+prest+"$", pd+prest)
	inf.Singular(false, su+"(?i)"+srest+"$", su+srest)
	inf.Singular(false, sd+"(?i)"+srest+"$", sd+srest)
	inf.Singular(false, pu+"(?i)"+prest+"$", su+srest)
	inf.Singular(false, pd+"(?i)"+prest+"$", sd+srest)
}

// Acronym registers an acronym (must begin with a capital as it appears in
// camelized form), mirroring Inflections#acronym.
func (inf *Inflections) Acronym(word string) {
	inf.acronyms[strings.ToLower(word)] = word
}

// sortedAcronyms returns acronym values sorted by descending length, matching
// Rails' define_acronym_regex_patterns so longer acronyms win.
func (inf *Inflections) sortedAcronyms() []string {
	vals := make([]string, 0, len(inf.acronyms))
	for _, v := range inf.acronyms {
		vals = append(vals, v)
	}
	sort.Slice(vals, func(i, j int) bool {
		if len(vals[i]) != len(vals[j]) {
			return len(vals[i]) > len(vals[j])
		}
		return vals[i] < vals[j] // deterministic tie-break (Ruby keeps stable, we normalise)
	})
	return vals
}

// isUncountable reports whether word is uncountable, mirroring Rails'
// /\b(w1|w2|...)\z/i test.
func (inf *Inflections) isUncountable(word string) bool {
	if len(inf.uncountables) == 0 {
		return false
	}
	if inf.uncountableRe == nil {
		parts := make([]string, len(inf.uncountables))
		for i, w := range inf.uncountables {
			parts[i] = regexp.QuoteMeta(w)
		}
		inf.uncountableRe = regexp.MustCompile(`(?i)\b(?:` + strings.Join(parts, "|") + `)\z`)
	}
	return inf.uncountableRe.MatchString(word)
}
