// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package numberhelper

import (
	"math/big"
	"strconv"
	"strings"
)

// resolved is the fully-defaulted rounding/formatting option set.
type resolved struct {
	precision   int
	significant bool
	separator   string
	delimiter   string
	stripZeros  bool
	roundMode   string
}

// groupThousands inserts delim between groups of three digits from the right,
// mirroring the default delimiter regex. A leading sign and any non-grouped
// short value are passed through unchanged.
func groupThousands(left, delim string) string {
	sign := ""
	if strings.HasPrefix(left, "-") {
		sign, left = "-", left[1:]
	}
	if delim == "" || len(left) <= 3 {
		return sign + left
	}
	off := len(left) % 3
	var parts []string
	if off > 0 {
		parts = append(parts, left[:off])
	}
	for i := off; i < len(left); i += 3 {
		parts = append(parts, left[i:i+3])
	}
	return sign + strings.Join(parts, delim)
}

// delimit splits numStr on "." and re-joins the delimited integer part with the
// fraction using separator (NumberToDelimitedConverter over a "."-decimal value).
func delimit(numStr, separator, delimiter string) string {
	left, right, hasDot := strings.Cut(numStr, ".")
	left = groupThousands(left, delimiter)
	if hasDot {
		return left + separator + right
	}
	return left
}

// stripInsignificantZeros removes trailing zeros after the separator, and the
// bare separator if the fraction becomes empty (RoundingConverter#format_number).
func stripInsignificantZeros(s, separator string) string {
	idx := strings.LastIndex(s, separator)
	if idx < 0 {
		return s
	}
	frac := strings.TrimRight(s[idx+len(separator):], "0")
	if frac == "" {
		return s[:idx]
	}
	return s[:idx+len(separator)] + frac
}

// roundedString reproduces NumberToRoundedConverter#convert: round the value,
// format to the (significant-adjusted) display precision, delimit and optionally
// strip insignificant zeros.
func roundedString(numStr string, floatVal float64, r resolved) string {
	roundPlaces := r.precision
	if r.significant && r.precision > 0 {
		roundPlaces = r.precision - digitCount(floatVal)
	}
	fstr := roundToF(numStr, roundPlaces, r.roundMode)

	displayPrec := r.precision
	if r.significant && r.precision > 0 {
		roundedFloat, _ := strconv.ParseFloat(fstr, 64)
		displayPrec = r.precision - digitCount(roundedFloat)
		if displayPrec < 0 {
			displayPrec = 0
		}
	}

	a, b, _ := strings.Cut(fstr, ".")
	if displayPrec != 0 {
		b += strings.Repeat("0", displayPrec)
		a = a + "." + b[:displayPrec]
	}

	out := delimit(a, r.separator, r.delimiter)
	if r.stripZeros {
		out = stripInsignificantZeros(out, r.separator)
	}
	return out
}

// roundedApply validates n and either rounds+formats it or returns its raw
// string (NumberToRoundedConverter#execute).
func roundedApply(n any, r resolved) string {
	numStr, ok := validNumber(n)
	if !ok {
		return numToString(n)
	}
	return roundedString(numStr, toFloat(n), r)
}

// --- public helpers ---------------------------------------------------------

// NumberToDelimited formats a number with grouped thousands
// (ActiveSupport::NumberHelper#number_to_delimited).
func NumberToDelimited(n any, opts ...Options) string {
	o := firstOpt(opts)
	numStr, ok := validNumber(n)
	if !ok {
		return numToString(n)
	}
	return delimit(numStr, orStr(o.Separator, "."), orStr(o.Delimiter, ","))
}

// NumberToRounded rounds a number to a precision (or significant digits)
// (number_to_rounded).
func NumberToRounded(n any, opts ...Options) string {
	o := firstOpt(opts)
	return roundedApply(n, resolved{
		precision:   orInt(o.Precision, 3),
		significant: orBool(o.Significant, false),
		separator:   orStr(o.Separator, "."),
		delimiter:   orStr(o.Delimiter, ""),
		stripZeros:  orBool(o.StripInsignificantZeros, false),
		roundMode:   o.RoundMode,
	})
}

// NumberToPercentage formats a number as a percentage (number_to_percentage).
func NumberToPercentage(n any, opts ...Options) string {
	o := firstOpt(opts)
	s := roundedApply(n, resolved{
		precision:   orInt(o.Precision, 3),
		significant: orBool(o.Significant, false),
		separator:   orStr(o.Separator, "."),
		delimiter:   orStr(o.Delimiter, ""),
		stripZeros:  orBool(o.StripInsignificantZeros, false),
		roundMode:   o.RoundMode,
	})
	return strings.ReplaceAll(orStr(o.Format, "%n%"), "%n", s)
}

// NumberToCurrency formats a number as currency (number_to_currency).
func NumberToCurrency(n any, opts ...Options) string {
	o := firstOpt(opts)
	unit := orStr(o.Unit, "$")
	precision := orInt(o.Precision, 2)

	activeFormat := orStr(o.Format, "%u%n")
	negativeFmt := "-" + activeFormat
	if o.NegativeFormat != nil {
		negativeFmt = *o.NegativeFormat
	}

	rnd := resolved{
		precision:   precision,
		significant: orBool(o.Significant, false),
		separator:   orStr(o.Separator, "."),
		delimiter:   orStr(o.Delimiter, ","),
		stripZeros:  orBool(o.StripInsignificantZeros, false),
		roundMode:   o.RoundMode,
	}

	numStr, ok := validNumber(n)
	var numberS string
	if ok {
		val := new(big.Rat)
		val.SetString(numStr)
		if val.Sign() < 0 {
			val.Abs(val)
			// abs * 10^precision >= 0.5  <=>  abs * 10^precision * 2 >= 1
			scaled := new(big.Rat).Mul(val, new(big.Rat).SetInt(pow10(precision)))
			scaled.Mul(scaled, big.NewRat(2, 1))
			if scaled.Cmp(big.NewRat(1, 1)) >= 0 {
				activeFormat = negativeFmt
			}
		}
		numberS = roundedString(val.FloatString(precision+2), absFloat(toFloat(n)), rnd)
	} else {
		numberS = strings.TrimSpace(numToString(n))
		if strings.HasPrefix(numberS, "-") {
			numberS = numberS[1:]
			activeFormat = negativeFmt
		}
	}
	out := strings.ReplaceAll(activeFormat, "%n", numberS)
	return strings.ReplaceAll(out, "%u", unit)
}

func absFloat(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
