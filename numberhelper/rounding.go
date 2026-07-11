// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package numberhelper

import (
	"math/big"
	"strings"
)

var bigTen = big.NewInt(10)

// pow10 returns 10^n as a big.Int for n >= 0.
func pow10(n int) *big.Int {
	return new(big.Int).Exp(bigTen, big.NewInt(int64(n)), nil)
}

// roundRatToInt rounds a signed rational to the nearest integer using the given
// BigDecimal round mode, returning a big.Int.
func roundRatToInt(x *big.Rat, mode string) *big.Int {
	neg := x.Sign() < 0
	abs := new(big.Rat).Abs(x)
	num := abs.Num()
	den := abs.Denom()
	q := new(big.Int)
	r := new(big.Int)
	q.QuoRem(num, den, r) // q = floor(abs), r = remainder numerator (>= 0)

	roundUp := false
	if r.Sign() != 0 {
		// Compare fraction r/den to 1/2 via 2*r vs den.
		twoR := new(big.Int).Lsh(r, 1)
		cmp := twoR.Cmp(den) // <0: frac<0.5, 0: ==0.5, >0: >0.5
		switch normalizeMode(mode) {
		case "up":
			roundUp = true
		case "down":
			roundUp = false
		case "ceiling":
			roundUp = !neg // toward +inf
		case "floor":
			roundUp = neg // toward -inf
		case "half_up":
			roundUp = cmp >= 0
		case "half_down":
			roundUp = cmp > 0
		case "half_even":
			if cmp > 0 {
				roundUp = true
			} else if cmp < 0 {
				roundUp = false
			} else {
				roundUp = q.Bit(0) == 1 // round to even
			}
		}
	}
	if roundUp {
		q.Add(q, big.NewInt(1))
	}
	if neg {
		q.Neg(q)
	}
	return q
}

// normalizeMode maps mode aliases to the canonical set.
func normalizeMode(mode string) string {
	switch mode {
	case "", "default":
		return "half_up"
	case "banker":
		return "half_even"
	case "ceil":
		return "ceiling"
	case "to_zero", "truncate":
		return "down"
	default:
		return mode
	}
}

// roundToF rounds numStr (a decimal string) to `places` decimal places (which may
// be negative) using mode, and returns the minimal BigDecimal "F" string: always
// carrying a fractional part ("13.0", "1200.0", "111.235") and never a negative
// zero.
func roundToF(numStr string, places int, mode string) string {
	r := new(big.Rat)
	r.SetString(numStr)

	var scaled *big.Rat
	if places >= 0 {
		scaled = new(big.Rat).Mul(r, new(big.Rat).SetInt(pow10(places)))
	} else {
		scaled = new(big.Rat).Quo(r, new(big.Rat).SetInt(pow10(-places)))
	}
	n := roundRatToInt(scaled, mode) // integer = value * 10^places

	if places <= 0 {
		value := new(big.Int).Mul(n, pow10(-places))
		if value.Sign() == 0 {
			return "0.0"
		}
		return value.String() + ".0"
	}

	neg := n.Sign() < 0
	digits := new(big.Int).Abs(n).String()
	for len(digits) <= places {
		digits = "0" + digits
	}
	intPart := digits[:len(digits)-places]
	fracPart := strings.TrimRight(digits[len(digits)-places:], "0")
	if fracPart == "" {
		fracPart = "0"
	}
	sign := ""
	if neg {
		sign = "-"
	}
	return sign + intPart + "." + fracPart
}
