// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package numberhelper is a pure-Go, MRI-faithful port of
// ActiveSupport::NumberHelper.
//
// It formats numbers the way Rails does — delimited thousands, fixed/significant
// rounding with every BigDecimal round mode, currency, percentage, human-readable
// sizes and quantities, and phone numbers — matching the gem's observable output
// byte-for-byte. Decimal rounding is done with math/big rationals (not binary
// floats) so half-up/half-even/etc. behave exactly like BigDecimal#round.
//
// Numbers may be passed as int, int64, float64 or string; a value that is not a
// valid number is returned formatted around its raw string, exactly as Rails
// does (e.g. NumberToCurrency("abc") == "$abc").
package numberhelper

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// Options carries the formatting overrides. A nil pointer means "use the
// per-helper default"; a set pointer overrides it. This mirrors the option
// hashes Rails' number_to_* helpers accept.
type Options struct {
	Precision               *int
	Significant             *bool
	Separator               *string
	Delimiter               *string
	StripInsignificantZeros *bool
	RoundMode               string // "" uses the default (half up)

	// Currency / percentage / human.
	Unit           *string
	Format         *string
	NegativeFormat *string

	// number_to_human unit table override (exponent name -> label).
	Units map[string]string

	// number_to_phone.
	AreaCode    bool
	CountryCode any
	Extension   any
	Pattern     *regexp.Regexp
}

func IntPtr(i int) *int       { return &i }
func BoolPtr(b bool) *bool    { return &b }
func StrPtr(s string) *string { return &s }
func orInt(p *int, d int) int {
	if p != nil {
		return *p
	}
	return d
}
func orBool(p *bool, d bool) bool {
	if p != nil {
		return *p
	}
	return d
}
func orStr(p *string, d string) string {
	if p != nil {
		return *p
	}
	return d
}

// firstOpt returns the first Options in opts, or a zero Options.
func firstOpt(opts []Options) Options {
	if len(opts) > 0 {
		return opts[0]
	}
	return Options{}
}

// --- number normalisation ---------------------------------------------------

// numToString renders a number the way Ruby's Numeric#to_s / String does.
func numToString(n any) string {
	switch t := n.(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		if math.IsInf(t, 0) || math.IsNaN(t) {
			return rubyNonFinite(t)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return ""
	}
}

// rubyNonFinite renders Float::INFINITY / NAN the way Ruby's "%f" does.
func rubyNonFinite(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case f > 0:
		return "Inf"
	default:
		return "-Inf"
	}
}

// validNumber reports whether n is a valid finite number and returns its decimal
// string (valid_bigdecimal). Strings containing d/e exponent markers are treated
// as invalid, matching Rails.
func validNumber(n any) (string, bool) {
	switch t := n.(type) {
	case int:
		return strconv.Itoa(t), true
	case int64:
		return strconv.FormatInt(t, 10), true
	case float64:
		if math.IsInf(t, 0) || math.IsNaN(t) {
			return "", false
		}
		return strconv.FormatFloat(t, 'f', -1, 64), true
	case string:
		if strings.ContainsAny(t, "dDeE") {
			return "", false
		}
		if _, ok := new(big.Rat).SetString(t); !ok {
			return "", false
		}
		return t, true
	default:
		return "", false
	}
}

// toFloat parses n's decimal string to a float64 for the log10-based helpers.
func toFloat(n any) float64 {
	f, _ := strconv.ParseFloat(numToString(n), 64)
	return f
}

// digitCount replicates RoundingHelper#digit_count: floor(log10(abs)) + 1, or 1
// for zero — including its float-log quirks, matching Rails exactly.
func digitCount(f float64) int {
	if f == 0 {
		return 1
	}
	return int(math.Floor(math.Log10(math.Abs(f)))) + 1
}
