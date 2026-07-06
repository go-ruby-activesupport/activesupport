// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package inflector

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Pre-compiled patterns that carry no lookaround (RE2-safe). The camel-boundary
// and acronym-aware transforms that need Ruby lookbehind/lookahead are done by
// hand-written scanners below, so their observable behaviour still matches MRI.
var (
	reAllLowerDigit     = regexp.MustCompile(`^[a-z\d]*$`)
	reLeadingLowerDigit = regexp.MustCompile(`^[a-z\d]*`)
	reCamelSeg          = regexp.MustCompile(`(?i)(?:_|(/))([a-z\d]*)`)
	reNeedsUnderscore   = regexp.MustCompile(`[A-Z-]|::`)
	reAlnumRun          = regexp.MustCompile(`[\p{L}0-9]+`)
	reNonURLChar        = regexp.MustCompile(`(?i)[^a-z0-9\-_]+`)
)

// subFirst replaces only the first match of re in s using repl (Go Expand
// syntax), mirroring Ruby's String#sub!. It reports whether a substitution
// happened.
func subFirst(re *regexp.Regexp, repl, s string) (string, bool) {
	loc := re.FindStringSubmatchIndex(s)
	if loc == nil {
		return s, false
	}
	dst := re.ExpandString(nil, repl, s, loc)
	return s[:loc[0]] + string(dst) + s[loc[1]:], true
}

// gsubFunc replaces every match of re, calling f with the submatch strings
// (index 0 is the whole match; unmatched groups are ""). It mirrors Ruby's
// String#gsub with a block.
func gsubFunc(re *regexp.Regexp, s string, f func(groups []string) string) string {
	var b strings.Builder
	last := 0
	for _, idx := range re.FindAllStringSubmatchIndex(s, -1) {
		b.WriteString(s[last:idx[0]])
		groups := make([]string, len(idx)/2)
		for i := range groups {
			if idx[2*i] >= 0 {
				groups[i] = s[idx[2*i]:idx[2*i+1]]
			}
		}
		b.WriteString(f(groups))
		last = idx[1]
	}
	b.WriteString(s[last:])
	return b.String()
}

// ---- Pluralize / Singularize ------------------------------------------------

func (inf *Inflections) apply(word string, rules []rule) string {
	if word == "" || inf.isUncountable(word) {
		return word
	}
	for _, r := range rules {
		if out, ok := subFirst(r.re, r.repl, word); ok {
			return out
		}
	}
	return word
}

// Pluralize returns the plural form of word.
func (inf *Inflections) Pluralize(word string) string { return inf.apply(word, inf.plurals) }

// Singularize returns the singular form of word.
func (inf *Inflections) Singularize(word string) string { return inf.apply(word, inf.singulars) }

// ---- Camelize ---------------------------------------------------------------

// Camelize converts snake_case to UpperCamelCase (or lowerCamelCase when
// upperFirstLetter is false). It also turns "/" into "::". Registered acronyms
// are preserved.
func (inf *Inflections) Camelize(term string, upperFirstLetter bool) string {
	s := term
	switch {
	case !upperFirstLetter:
		s = inf.lowerFirstToken(s)
	case reAllLowerDigit.MatchString(s):
		if a, ok := inf.acronyms[s]; ok {
			return a
		}
		return rubyCapitalize(s)
	default:
		s = inf.upperFirstToken(s)
	}
	return gsubFunc(reCamelSeg, s, func(g []string) string {
		word := g[2]
		sub, ok := inf.acronyms[word]
		if !ok {
			sub = rubyCapitalize(word)
		}
		if g[1] == "/" {
			return "::" + sub
		}
		return sub
	})
}

// upperFirstToken replaces the leading [a-z0-9]* run with its acronym form or
// its capitalization (Ruby: string.sub(/^[a-z\d]*/){ acronyms[m] || m.capitalize }).
func (inf *Inflections) upperFirstToken(s string) string {
	loc := reLeadingLowerDigit.FindStringIndex(s)
	run := s[loc[0]:loc[1]]
	if a, ok := inf.acronyms[run]; ok {
		return a + s[loc[1]:]
	}
	return rubyCapitalize(run) + s[loc[1]:]
}

// lowerFirstToken lowercases the leading acronym (if any) or the first word
// character, mirroring the acronyms_camelize_regex sub in Ruby's camelize.
func (inf *Inflections) lowerFirstToken(s string) string {
	if s == "" {
		return s
	}
	for _, a := range inf.sortedAcronyms() {
		if strings.HasPrefix(s, a) {
			rest := s[len(a):]
			// (?=\b|[A-Z_]) — next char must not be a lowercase letter or digit.
			if rest == "" || !(isLowerAlpha(rest[0]) || isDigit(rest[0])) {
				return strings.ToLower(a) + rest
			}
		}
	}
	if !isWordByte(s[0]) { // Ruby \w (ASCII) did not match ⇒ no change
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return strings.ToLower(string(r)) + s[size:]
}

// ---- Underscore -------------------------------------------------------------

// Underscore converts UpperCamelCase to snake_case, turning "::" into "/".
func (inf *Inflections) Underscore(camelCased string) string {
	if !reNeedsUnderscore.MatchString(camelCased) {
		return camelCased
	}
	w := strings.ReplaceAll(camelCased, "::", "/")
	w = inf.acronymUnderscore(w)
	w = insertUnderscores(w)
	w = strings.ReplaceAll(w, "-", "_")
	return strings.ToLower(w)
}

// acronymUnderscore replays gsub(acronyms_underscore_regex){ ... } without
// lookbehind/lookahead. It is a no-op when no acronyms are registered.
func (inf *Inflections) acronymUnderscore(s string) string {
	if len(inf.acronyms) == 0 {
		return s
	}
	acrs := inf.sortedAcronyms()
	var b strings.Builder
	i := 0
	for i < len(s) {
		matched := false
		for _, a := range acrs {
			if !strings.HasPrefix(s[i:], a) {
				continue
			}
			after := i + len(a)
			// (?=\b|[^a-z]) — the char after the acronym must not be a-z.
			if after < len(s) && isLowerAlpha(s[after]) {
				continue
			}
			// Prefix: (?:(?<=[A-Za-z\d]) $1 | \b). '_' before satisfies neither.
			if i == 0 {
				b.WriteString(strings.ToLower(a))
			} else if prev := s[i-1]; isAlnumByte(prev) {
				b.WriteByte('_')
				b.WriteString(strings.ToLower(a))
			} else if prev == '_' {
				continue // no match at this position; try next acronym
			} else {
				b.WriteString(strings.ToLower(a))
			}
			i = after
			matched = true
			break
		}
		if !matched {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}

// insertUnderscores inserts "_" at CamelCase boundaries, replicating the Ruby
// lookbehind/lookahead pattern
// /(?<=[A-Z])(?=[A-Z][a-z])|(?<=[a-z\d])(?=[A-Z])/.
func insertUnderscores(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if i > 0 {
			prev, cur := s[i-1], s[i]
			switch {
			case isUpper(prev) && isUpper(cur) && i+1 < len(s) && isLowerAlpha(s[i+1]):
				b.WriteByte('_')
			case (isLowerAlpha(prev) || isDigit(prev)) && isUpper(cur):
				b.WriteByte('_')
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// ---- Humanize / Titleize ----------------------------------------------------

// Humanize turns an attribute name into a human-friendly phrase. capitalize
// controls whether the first letter is upper-cased; keepIDSuffix keeps a trailing
// "_id".
func (inf *Inflections) Humanize(word string, capitalize, keepIDSuffix bool) string {
	result := word
	for _, r := range inf.humans {
		if out, ok := subFirst(r.re, r.repl, result); ok {
			result = out
			break
		}
	}
	result = strings.ReplaceAll(result, "_", " ")
	result = strings.TrimLeft(result, "\x00\t\n\v\f\r ")
	if !keepIDSuffix && strings.HasSuffix(word, "_id") {
		result = strings.TrimSuffix(result, " id")
	}
	result = gsubFunc(reAlnumRun, result, func(g []string) string {
		lower := strings.ToLower(g[0])
		if a, ok := inf.acronyms[lower]; ok {
			return a
		}
		return lower
	})
	if capitalize {
		result = upcaseFirstAlpha(result)
	}
	return result
}

// Titleize produces a pretty, capitalized title.
func (inf *Inflections) Titleize(word string, keepIDSuffix bool) string {
	h := inf.Humanize(inf.Underscore(word), true, keepIDSuffix)
	return titlecaseInitials(h)
}

// titlecaseInitials capitalizes the first letter of each word, skipping letters
// right after an in-word apostrophe/quote/paren — Ruby
// /\b(?<!\w['’`()])[a-z]/.
func titlecaseInitials(s string) string {
	rs := []rune(s)
	for i, c := range rs {
		if c < 'a' || c > 'z' {
			continue
		}
		if i > 0 && isWordRune(rs[i-1]) { // not a word boundary
			continue
		}
		if i >= 2 && isWordRune(rs[i-2]) && isTitleSkipPunct(rs[i-1]) {
			continue
		}
		rs[i] = unicode.ToUpper(c)
	}
	return string(rs)
}

// ---- Tableize / Classify ----------------------------------------------------

// Tableize turns a model name into its table name.
func (inf *Inflections) Tableize(className string) string {
	return inf.Pluralize(inf.Underscore(className))
}

// Classify turns a (possibly schema-qualified) table name into a class name.
func (inf *Inflections) Classify(tableName string) string {
	s := tableName
	if i := strings.LastIndex(s, "."); i >= 0 {
		s = s[i+1:]
	}
	return inf.Camelize(inf.Singularize(s), true)
}

// ForeignKey builds a foreign-key column name from a class name.
func (inf *Inflections) ForeignKey(className string, separateWithUnderscore bool) string {
	base := inf.Underscore(Demodulize(className))
	if separateWithUnderscore {
		return base + "_id"
	}
	return base + "id"
}

// ---- Stateless helpers ------------------------------------------------------

// Dasherize replaces underscores with dashes.
func Dasherize(word string) string { return strings.ReplaceAll(word, "_", "-") }

// Demodulize strips the module part of a constant path.
func Demodulize(path string) string {
	if i := strings.LastIndex(path, "::"); i >= 0 {
		return path[i+2:]
	}
	return path
}

// Deconstantize removes the rightmost segment of a constant path.
func Deconstantize(path string) string {
	if i := strings.LastIndex(path, "::"); i >= 0 {
		return path[:i]
	}
	return ""
}

// Ordinal returns the ordinal suffix ("st","nd","rd","th") for number.
func Ordinal(number int) string {
	abs := number
	if abs < 0 {
		abs = -abs
	}
	if m := abs % 100; m >= 11 && m <= 13 {
		return "th"
	}
	switch abs % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

// Ordinalize renders number followed by its ordinal suffix (e.g. "1st").
func Ordinalize(number int) string { return strconv.Itoa(number) + Ordinal(number) }

// Transliterate replaces non-ASCII characters with an ASCII approximation, or
// replacement when none exists. Input is assumed to be NFC-normalised UTF-8
// (the realistic case for precomposed Latin text), matching the :en locale of
// ActiveSupport::Inflector.transliterate.
func Transliterate(s, replacement string) string {
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			ascii = false
			break
		}
	}
	if ascii {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r < utf8.RuneSelf:
			b.WriteRune(r)
		default:
			if v, ok := transliterations[r]; ok {
				b.WriteString(v)
			} else {
				b.WriteString(replacement)
			}
		}
	}
	return b.String()
}

// Parameterize turns a string into a URL-safe slug.
func Parameterize(s, separator string, preserveCase bool) string {
	p := Transliterate(s, "?")
	p = reNonURLChar.ReplaceAllLiteralString(p, separator)
	if separator != "" {
		if len(separator) == 1 {
			p = squeezeByte(p, separator[0])
		} else {
			re := regexp.MustCompile(regexp.QuoteMeta(separator) + "{2,}")
			p = re.ReplaceAllLiteralString(p, separator)
		}
		p = strings.TrimPrefix(p, separator)
		p = strings.TrimSuffix(p, separator)
	}
	if !preserveCase {
		p = strings.ToLower(p)
	}
	return p
}

// squeezeByte collapses runs of byte c to a single c (Ruby String#squeeze).
func squeezeByte(s string, c byte) string {
	var b strings.Builder
	prev := false
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			if prev {
				continue
			}
			prev = true
		} else {
			prev = false
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// ---- Constant resolution seam ----------------------------------------------

// Resolver looks up a (possibly namespaced) constant name and reports whether it
// exists. The rbgo binding supplies one backed by the Ruby object space; tests
// supply a fake. This is the seam standing in for Ruby's Object.const_get.
type Resolver func(name string) (value any, ok bool)

// NameError models Ruby's NameError raised by constantize for an unknown or
// malformed constant name.
type NameError struct{ Name string }

func (e *NameError) Error() string { return "uninitialized constant " + e.Name }

// Constantize resolves name via resolve, returning a *NameError when it is
// unknown, mirroring Inflector.constantize.
func Constantize(name string, resolve Resolver) (any, error) {
	if v, ok := resolve(name); ok {
		return v, nil
	}
	return nil, &NameError{Name: name}
}

// SafeConstantize resolves name, returning nil when it is unknown (never
// erroring), mirroring Inflector.safe_constantize.
func SafeConstantize(name string, resolve Resolver) any {
	if v, ok := resolve(name); ok {
		return v
	}
	return nil
}

// ---- small char helpers -----------------------------------------------------

func isUpper(b byte) bool      { return b >= 'A' && b <= 'Z' }
func isLowerAlpha(b byte) bool { return b >= 'a' && b <= 'z' }
func isDigit(b byte) bool      { return b >= '0' && b <= '9' }
func isAlnumByte(b byte) bool  { return isUpper(b) || isLowerAlpha(b) || isDigit(b) }
func isWordByte(b byte) bool   { return isAlnumByte(b) || b == '_' }

// isWordRune matches Ruby's \w on a UTF-8 string, which is Unicode-aware: a word
// character is any letter or digit, or an underscore. This governs the \b word
// boundary used by titleize.
func isWordRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isTitleSkipPunct(r rune) bool {
	switch r {
	case '\'', '’', '`', '(', ')':
		return true
	}
	return false
}

func rubyCapitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + strings.ToLower(s[size:])
}

func upcaseFirstAlpha(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if unicode.IsLetter(r) {
		return string(unicode.ToUpper(r)) + s[size:]
	}
	return s
}
