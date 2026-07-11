// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package duration is a pure-Go, MRI-faithful port of ActiveSupport::Duration.
//
// A Duration is an amount of time expressed as an ordered set of parts (years,
// months, weeks, days, hours, minutes, seconds) plus a total value in seconds.
// It mirrors Rails' observable behaviour byte-for-byte: the humanised Inspect
// output, the ISO 8601 serialisation and parser, the Build/decompose algorithm,
// arithmetic (which preserves and merges parts), and the In<Unit> conversions.
//
// The gregorian-average constants match Rails: a month is 1/12 of a gregorian
// year (2629746 s) and a year is 31556952 s.
package duration

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// Seconds-per-unit constants, matching ActiveSupport::Duration.
const (
	SecondsPerMinute = 60
	SecondsPerHour   = 3600
	SecondsPerDay    = 86400
	SecondsPerWeek   = 604800
	SecondsPerMonth  = 2629746  // 1/12 of a gregorian year
	SecondsPerYear   = 31556952 // length of a gregorian year (365.2425 days)
)

// Unit identifies a calendar/clock unit.
type Unit int

// The units, in canonical PARTS order.
const (
	UnitYears Unit = iota
	UnitMonths
	UnitWeeks
	UnitDays
	UnitHours
	UnitMinutes
	UnitSeconds
)

// unitNames are the plural Ruby names; the singular is the plural with its last
// byte chopped (year, month, …), exactly as Duration#inspect does.
var unitNames = map[Unit]string{
	UnitYears: "years", UnitMonths: "months", UnitWeeks: "weeks", UnitDays: "days",
	UnitHours: "hours", UnitMinutes: "minutes", UnitSeconds: "seconds",
}

var secondsPerUnit = map[Unit]float64{
	UnitYears: SecondsPerYear, UnitMonths: SecondsPerMonth, UnitWeeks: SecondsPerWeek,
	UnitDays: SecondsPerDay, UnitHours: SecondsPerHour, UnitMinutes: SecondsPerMinute,
	UnitSeconds: 1,
}

// partsOrder is the canonical PARTS ordering used by Inspect and normalisation.
var partsOrder = []Unit{UnitYears, UnitMonths, UnitWeeks, UnitDays, UnitHours, UnitMinutes, UnitSeconds}

// variableUnits are the parts whose real length depends on the calendar.
var variableUnits = map[Unit]bool{UnitYears: true, UnitMonths: true, UnitWeeks: true, UnitDays: true}

// Part is a single (unit, scalar) component of a Duration.
type Part struct {
	Unit   Unit
	Scalar float64
}

// Duration is an amount of time. The zero value is an empty 0-second duration.
type Duration struct {
	value float64
	parts []Part // insertion order, with zero parts removed unless value == 0
}

// newDuration constructs a Duration, dropping zero parts unless value == 0
// (matching Duration#initialize).
func newDuration(value float64, parts []Part) Duration {
	if value != 0 {
		kept := make([]Part, 0, len(parts))
		for _, p := range parts {
			if p.Scalar != 0 {
				kept = append(kept, p)
			}
		}
		parts = kept
	}
	return Duration{value: value, parts: parts}
}

// scalar builds a single-unit Duration (the n.days / n.hours constructors).
func scalar(u Unit, n float64) Duration {
	return newDuration(n*secondsPerUnit[u], []Part{{u, n}})
}

// Years returns a Duration of n years (Numeric#years).
func Years(n float64) Duration { return scalar(UnitYears, n) }

// Months returns a Duration of n months (Numeric#months).
func Months(n float64) Duration { return scalar(UnitMonths, n) }

// Weeks returns a Duration of n weeks (Numeric#weeks).
func Weeks(n float64) Duration { return scalar(UnitWeeks, n) }

// Days returns a Duration of n days (Numeric#days).
func Days(n float64) Duration { return scalar(UnitDays, n) }

// Hours returns a Duration of n hours (Numeric#hours).
func Hours(n float64) Duration { return scalar(UnitHours, n) }

// Minutes returns a Duration of n minutes (Numeric#minutes).
func Minutes(n float64) Duration { return scalar(UnitMinutes, n) }

// Seconds returns a Duration of n seconds (Numeric#seconds).
func Seconds(n float64) Duration { return scalar(UnitSeconds, n) }

// Value returns the total length of the duration in seconds (Duration#value).
func (d Duration) Value() float64 { return d.value }

// ToI returns the duration in whole seconds, truncated toward zero
// (Duration#to_i).
func (d Duration) ToI() int64 { return int64(d.value) }

// Parts returns the duration's parts in canonical (Years→Seconds) order
// (Duration#parts, canonically sorted).
func (d Duration) Parts() []Part {
	out := make([]Part, 0, len(d.parts))
	for _, u := range partsOrder {
		for _, p := range d.parts {
			if p.Unit == u {
				out = append(out, p)
			}
		}
	}
	return out
}

// IsVariable reports whether any part is calendar-variable (Duration#variable?).
func (d Duration) IsVariable() bool {
	for _, p := range d.parts {
		if variableUnits[p.Unit] {
			return true
		}
	}
	return false
}

// Build decomposes a number of seconds into a normalised Duration, greedily
// filling the largest units first (ActiveSupport::Duration.build).
func Build(value float64) Duration {
	parts := make([]Part, 0, len(partsOrder))
	if value != 0 {
		sign := 1.0
		if value < 0 {
			sign = -1.0
		}
		remainder := round9(value) * sign // abs
		for _, u := range partsOrder {
			if u == UnitSeconds {
				continue
			}
			per := secondsPerUnit[u]
			q := float64(int64(remainder / per)) // integer division of non-negative remainder
			parts = append(parts, Part{u, q * sign})
			remainder = round9(remainder - q*per)
		}
		parts = append(parts, Part{UnitSeconds, remainder * sign})
	} else {
		parts = append(parts, Part{UnitSeconds, 0})
	}
	return newDuration(value, parts)
}

// round9 rounds to 9 decimal places, matching Ruby's Float#round(9) used by Build.
func round9(x float64) float64 {
	const scale = 1e9
	if x < 0 {
		return -round9(-x)
	}
	// round half up
	return float64(int64(x*scale+0.5)) / scale
}

// InSeconds returns the duration in seconds (Duration#in_seconds).
func (d Duration) InSeconds() float64 { return d.value }

// InMinutes returns the duration in minutes (Duration#in_minutes).
func (d Duration) InMinutes() float64 { return d.value / SecondsPerMinute }

// InHours returns the duration in hours (Duration#in_hours).
func (d Duration) InHours() float64 { return d.value / SecondsPerHour }

// InDays returns the duration in days (Duration#in_days).
func (d Duration) InDays() float64 { return d.value / SecondsPerDay }

// InWeeks returns the duration in weeks (Duration#in_weeks).
func (d Duration) InWeeks() float64 { return d.value / SecondsPerWeek }

// InMonths returns the duration in months (Duration#in_months).
func (d Duration) InMonths() float64 { return d.value / SecondsPerMonth }

// InYears returns the duration in years (Duration#in_years).
func (d Duration) InYears() float64 { return d.value / SecondsPerYear }

// Add returns the sum of two durations, merging matching parts and appending new
// ones in the receiver-then-other order (Duration#+).
func (d Duration) Add(other Duration) Duration {
	parts := make([]Part, len(d.parts))
	copy(parts, d.parts)
	for _, p := range other.parts {
		merged := false
		for i := range parts {
			if parts[i].Unit == p.Unit {
				parts[i].Scalar += p.Scalar
				merged = true
				break
			}
		}
		if !merged {
			parts = append(parts, p)
		}
	}
	return newDuration(d.value+other.value, parts)
}

// Sub returns d minus other (Duration#-).
func (d Duration) Sub(other Duration) Duration { return d.Add(other.Neg()) }

// Neg returns the negation of the duration (Duration#-@).
func (d Duration) Neg() Duration { return d.Mul(-1) }

// Mul scales the duration by a scalar factor (Duration#*).
func (d Duration) Mul(factor float64) Duration {
	parts := make([]Part, len(d.parts))
	for i, p := range d.parts {
		parts[i] = Part{p.Unit, p.Scalar * factor}
	}
	return newDuration(d.value*factor, parts)
}

// Cmp compares two durations by total value: -1, 0 or +1 (Duration#<=>).
func (d Duration) Cmp(other Duration) int {
	switch {
	case d.value < other.value:
		return -1
	case d.value > other.value:
		return 1
	default:
		return 0
	}
}

// Equal reports whether two durations have the same total value (Duration#==).
func (d Duration) Equal(other Duration) bool { return d.value == other.value }

// Inspect renders the human-readable duration, e.g. "1 year, 2 months, and 3
// days" (Duration#inspect / to_sentence with locale: false).
func (d Duration) Inspect() string {
	if len(d.parts) == 0 {
		return formatNum(d.value) + " seconds"
	}
	sorted := d.Parts()
	terms := make([]string, len(sorted))
	for i, p := range sorted {
		name := unitNames[p.Unit]
		if p.Scalar == 1 {
			name = name[:len(name)-1] // chop trailing "s"
		}
		terms[i] = formatNum(p.Scalar) + " " + name
	}
	return toSentence(terms)
}

// String is an alias for Inspect so a Duration prints usefully with %v.
func (d Duration) String() string { return d.Inspect() }

// toSentence joins terms the way Array#to_sentence does with locale: false:
// "a", "a and b", "a, b, and c".
func toSentence(terms []string) string {
	switch len(terms) {
	case 0:
		return ""
	case 1:
		return terms[0]
	case 2:
		return terms[0] + " and " + terms[1]
	default:
		return strings.Join(terms[:len(terms)-1], ", ") + ", and " + terms[len(terms)-1]
	}
}

// Iso8601 serialises the duration to an ISO 8601 duration string
// (Duration#iso8601 / ISO8601Serializer).
func (d Duration) Iso8601() string {
	// Sum parts, dropping zeros.
	sum := map[Unit]float64{}
	for _, p := range d.parts {
		if p.Scalar == 0 {
			continue
		}
		sum[p.Unit] += p.Scalar
	}
	if len(sum) == 0 {
		return "PT0S"
	}
	// Convert weeks to days when mixed with year/month/day date parts.
	if _, hasWeeks := sum[UnitWeeks]; hasWeeks {
		if _, y := sum[UnitYears]; y {
			mixWeeks(sum)
		} else if _, m := sum[UnitMonths]; m {
			mixWeeks(sum)
		} else if _, dd := sum[UnitDays]; dd {
			mixWeeks(sum)
		}
	}
	var b strings.Builder
	b.WriteByte('P')
	writeComp(&b, sum, UnitYears, "Y")
	writeComp(&b, sum, UnitMonths, "M")
	writeComp(&b, sum, UnitDays, "D")
	writeComp(&b, sum, UnitWeeks, "W")
	var t strings.Builder
	writeComp(&t, sum, UnitHours, "H")
	writeComp(&t, sum, UnitMinutes, "M")
	writeComp(&t, sum, UnitSeconds, "S")
	if t.Len() > 0 {
		b.WriteByte('T')
		b.WriteString(t.String())
	}
	return b.String()
}

func mixWeeks(sum map[Unit]float64) {
	sum[UnitDays] += sum[UnitWeeks] * SecondsPerWeek / SecondsPerDay
	delete(sum, UnitWeeks)
}

func writeComp(b *strings.Builder, sum map[Unit]float64, u Unit, designator string) {
	if v, ok := sum[u]; ok {
		b.WriteString(formatNum(v))
		b.WriteString(designator)
	}
}

// ErrParse is returned by Parse for any malformed ISO 8601 duration.
var ErrParse = errors.New("invalid ISO 8601 duration")

// Since returns the instant t advanced by the duration, applying each part in
// order and using calendar arithmetic for years/months/weeks/days
// (Duration#since / Time#+). Fractional calendar parts contribute their exact
// second-equivalent.
func (d Duration) Since(t time.Time) time.Time { return d.apply(t, 1) }

// Ago returns the instant t moved back by the duration (Duration#ago).
func (d Duration) Ago(t time.Time) time.Time { return d.apply(t, -1) }

func (d Duration) apply(t time.Time, sign float64) time.Time {
	if len(d.parts) == 0 {
		return t.Add(time.Duration(sign * d.value * float64(time.Second)))
	}
	for _, p := range d.parts {
		s := sign * p.Scalar
		whole := float64(int64(s))
		frac := s - whole
		switch p.Unit {
		case UnitYears:
			t = t.AddDate(int(whole), 0, 0)
		case UnitMonths:
			t = t.AddDate(0, int(whole), 0)
		case UnitWeeks:
			t = t.AddDate(0, 0, int(whole)*7)
		case UnitDays:
			t = t.AddDate(0, 0, int(whole))
		default:
			whole, frac = 0, s // hours/minutes/seconds are exact
		}
		if frac != 0 {
			t = t.Add(time.Duration(frac * secondsPerUnit[p.Unit] * float64(time.Second)))
		}
	}
	return t
}

// formatNum renders a float the way Ruby's Numeric#to_s does for the values
// durations carry: integral values print without a decimal point, others use
// the shortest round-tripping decimal.
func formatNum(x float64) string {
	if x == float64(int64(x)) {
		return strconv.FormatInt(int64(x), 10)
	}
	return strconv.FormatFloat(x, 'f', -1, 64)
}
