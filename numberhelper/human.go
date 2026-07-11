// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package numberhelper

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// storageUnits are the labels for number_to_human_size, indexed by exponent
// (Bytes, KB, MB, GB, TB). Rails only defines units up to TB by default.
var storageUnits = []string{"Bytes", "KB", "MB", "GB", "TB"}

// NumberToHumanSize formats a byte count in human-readable form, e.g. "1.18 MB"
// (number_to_human_size). Sizes are grouped in powers of 1024, up to TB.
func NumberToHumanSize(n any, opts ...Options) string {
	o := firstOpt(opts)
	if _, ok := validNumber(n); !ok {
		return numToString(n)
	}
	f := toFloat(n)
	strip := orBool(o.StripInsignificantZeros, true)

	const base = 1024.0
	var numberToFormat string
	unit := storageUnits[0]
	if int64(math.Abs(f)) < int64(base) {
		numberToFormat = strconv.FormatInt(int64(f), 10)
	} else {
		exp := int(math.Log(math.Abs(f)) / math.Log(base))
		if exp > len(storageUnits)-1 {
			exp = len(storageUnits) - 1
		}
		unit = storageUnits[exp]
		size := f / math.Pow(base, float64(exp))
		numberToFormat = roundedString(strconv.FormatFloat(size, 'f', -1, 64), size, resolved{
			precision:   orInt(o.Precision, 3),
			significant: orBool(o.Significant, true),
			separator:   orStr(o.Separator, "."),
			delimiter:   orStr(o.Delimiter, ""),
			stripZeros:  strip,
			roundMode:   o.RoundMode,
		})
	}
	format := orStr(o.Format, "%n %u")
	format = strings.ReplaceAll(format, "%n", numberToFormat)
	return strings.ReplaceAll(format, "%u", unit)
}

// decimalUnitExp maps a decimal-unit name to its power-of-ten exponent.
var decimalUnitExp = map[string]int{
	"unit": 0, "ten": 1, "hundred": 2, "thousand": 3, "million": 6,
	"billion": 9, "trillion": 12, "quadrillion": 15,
	"deci": -1, "centi": -2, "mili": -3, "micro": -6, "nano": -9,
	"pico": -12, "femto": -15,
}

// defaultDecimalUnits is the label table Rails quantifies by default.
var defaultDecimalUnits = map[int]string{
	0: "", 3: "Thousand", 6: "Million", 9: "Billion", 12: "Trillion", 15: "Quadrillion",
}

// NumberToHuman formats a number in human-readable quantity form, e.g. "1.23
// Million" (number_to_human).
func NumberToHuman(n any, opts ...Options) string {
	o := firstOpt(opts)
	numStr, ok := validNumber(n)
	if !ok {
		return numToString(n)
	}
	rnd := resolved{
		precision:   orInt(o.Precision, 3),
		significant: orBool(o.Significant, true),
		separator:   orStr(o.Separator, "."),
		delimiter:   orStr(o.Delimiter, ""),
		stripZeros:  orBool(o.StripInsignificantZeros, true),
		roundMode:   o.RoundMode,
	}

	// Round the original number first (RoundingHelper), then work from that.
	rounded := roundToF(numStr, roundPlacesFor(rnd, toFloat(n)), rnd.roundMode)
	f, _ := strconv.ParseFloat(rounded, 64)

	labels := defaultDecimalUnits
	if o.Units != nil {
		labels = map[int]string{}
		for name, label := range o.Units {
			if exp, okName := decimalUnitExp[name]; okName {
				labels[exp] = label
			}
		}
	}
	exps := make([]int, 0, len(labels))
	for e := range labels {
		exps = append(exps, e)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(exps)))

	numExp := 0
	if f != 0 {
		numExp = int(math.Floor(math.Log10(math.Abs(f))))
	}
	selected := 0
	for _, e := range exps {
		if numExp >= e {
			selected = e
			break
		}
	}

	scaled := f / math.Pow10(selected)
	number := roundedString(strconv.FormatFloat(scaled, 'f', -1, 64), scaled, rnd)
	unit := labels[selected]

	format := orStr(o.Format, "%n %u")
	out := strings.ReplaceAll(format, "%n", number)
	out = strings.ReplaceAll(out, "%u", unit)
	return strings.TrimSpace(out)
}

// roundPlacesFor computes the decimal places RoundingHelper#round would use.
func roundPlacesFor(r resolved, f float64) int {
	if r.significant && r.precision > 0 {
		return r.precision - digitCount(f)
	}
	return r.precision
}
