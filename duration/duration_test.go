// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package duration

import (
	"errors"
	"testing"
	"time"
)

func TestConstructorsAndValue(t *testing.T) {
	cases := []struct {
		d    Duration
		want float64
	}{
		{Years(1), SecondsPerYear},
		{Months(1), SecondsPerMonth},
		{Weeks(1), SecondsPerWeek},
		{Days(1), SecondsPerDay},
		{Hours(1), SecondsPerHour},
		{Minutes(1), SecondsPerMinute},
		{Seconds(1), 1},
		{Hours(1.5), 5400},
	}
	for _, c := range cases {
		if c.d.Value() != c.want {
			t.Errorf("value = %v, want %v", c.d.Value(), c.want)
		}
	}
}

func TestToI(t *testing.T) {
	if got := Days(1).Add(Hours(2)).ToI(); got != 93600 {
		t.Errorf("ToI = %d, want 93600", got)
	}
	if got := Hours(1.5).ToI(); got != 5400 {
		t.Errorf("ToI = %d, want 5400", got)
	}
}

func TestParts(t *testing.T) {
	// Canonical ordering: append in reverse, expect Years→Seconds order.
	d := Seconds(6).Add(Hours(4)).Add(Years(1))
	got := d.Parts()
	want := []Part{{UnitYears, 1}, {UnitHours, 4}, {UnitSeconds, 6}}
	if len(got) != len(want) {
		t.Fatalf("Parts len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Parts[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestIsVariable(t *testing.T) {
	if !Days(1).IsVariable() {
		t.Error("days should be variable")
	}
	if Hours(1).IsVariable() {
		t.Error("hours should not be variable")
	}
	if Hours(1).Add(Days(1)).IsVariable() != true {
		t.Error("mixed should be variable")
	}
}

func TestBuild(t *testing.T) {
	if got := Build(0).Inspect(); got != "0 seconds" {
		t.Errorf("Build(0) = %q", got)
	}
	if got := Build(90).Inspect(); got != "1 minute and 30 seconds" {
		t.Errorf("Build(90) = %q", got)
	}
	if got := Build(3661).Inspect(); got != "1 hour, 1 minute, and 1 second" {
		t.Errorf("Build(3661) = %q", got)
	}
	if got := Build(-3661).Inspect(); got != "-1 hours, -1 minutes, and -1 seconds" {
		t.Errorf("Build(-3661) = %q", got)
	}
	if got := Build(SecondsPerYear).Inspect(); got != "1 year" {
		t.Errorf("Build(year) = %q", got)
	}
}

func TestRound9(t *testing.T) {
	if got := round9(-1.2345678904); got != -1.23456789 {
		t.Errorf("round9(neg) = %v", got)
	}
	if got := round9(1.2345678906); got != 1.234567891 {
		t.Errorf("round9 = %v", got)
	}
}

func TestInUnits(t *testing.T) {
	d := Minutes(90)
	checks := []struct {
		got, want float64
		name      string
	}{
		{d.InSeconds(), 5400, "seconds"},
		{d.InMinutes(), 90, "minutes"},
		{d.InHours(), 1.5, "hours"},
		{Days(1).InDays(), 1, "days"},
		{Weeks(2).InWeeks(), 2, "weeks"},
		{Months(1).InMonths(), 1, "months"},
		{Years(1).InYears(), 1, "years"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("In%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestArithmetic(t *testing.T) {
	if got := Days(1).Add(Days(1)).Inspect(); got != "2 days" {
		t.Errorf("1d+1d = %q", got)
	}
	if got := Hours(2).Sub(Minutes(30)).Inspect(); got != "2 hours and -30 minutes" {
		t.Errorf("2h-30m = %q", got)
	}
	if got := Days(1).Mul(3).Inspect(); got != "3 days" {
		t.Errorf("1d*3 = %q", got)
	}
	if got := Days(1).Neg().Value(); got != -SecondsPerDay {
		t.Errorf("neg = %v", got)
	}
	// Zero value keeps its (zero) part.
	z := Minutes(30).Add(Minutes(-30))
	if got := z.Inspect(); got != "0 minutes" {
		t.Errorf("30m-30m = %q", got)
	}
}

func TestCmpAndEqual(t *testing.T) {
	if Days(1).Cmp(Hours(25)) != -1 {
		t.Error("1d < 25h")
	}
	if Hours(25).Cmp(Days(1)) != 1 {
		t.Error("25h > 1d")
	}
	if Days(1).Cmp(Hours(24)) != 0 {
		t.Error("1d == 24h cmp")
	}
	if !Days(1).Equal(Hours(24)) {
		t.Error("1d == 24h")
	}
	if Days(1).Equal(Hours(23)) {
		t.Error("1d != 23h")
	}
}

func TestInspectAndString(t *testing.T) {
	// Empty-parts branch: the zero Duration.
	if got := (Duration{}).Inspect(); got != "0 seconds" {
		t.Errorf("empty inspect = %q", got)
	}
	if got := Hours(1.5).String(); got != "1.5 hours" {
		t.Errorf("String = %q", got)
	}
	if got := Seconds(2).Inspect(); got != "2 seconds" {
		t.Errorf("2s = %q", got)
	}
}

func TestToSentence(t *testing.T) {
	if got := toSentence(nil); got != "" {
		t.Errorf("empty = %q", got)
	}
	if got := toSentence([]string{"a"}); got != "a" {
		t.Errorf("one = %q", got)
	}
	if got := toSentence([]string{"a", "b"}); got != "a and b" {
		t.Errorf("two = %q", got)
	}
	if got := toSentence([]string{"a", "b", "c"}); got != "a, b, and c" {
		t.Errorf("three = %q", got)
	}
}

func TestIso8601(t *testing.T) {
	cases := []struct {
		d    Duration
		want string
	}{
		{Duration{}, "PT0S"},
		{Build(0), "PT0S"}, // parts carry a zero seconds entry
		{Days(30), "P30D"},
		{Seconds(90), "PT90S"},
		{Hours(1.5), "PT1.5H"},
		{Seconds(-90), "PT-90S"},
		{Weeks(3).Add(Days(4)), "P25D"},    // weeks mixed with days
		{Years(1).Add(Weeks(2)), "P1Y14D"}, // weeks mixed with years
		{Months(1).Add(Weeks(1)), "P1M7D"}, // weeks mixed with months
		{Weeks(2), "P2W"},                  // pure weeks kept
		{Years(1).Add(Months(2)).Add(Days(3)).Add(Hours(4)).Add(Minutes(5)).Add(Seconds(6)), "P1Y2M3DT4H5M6S"},
	}
	for _, c := range cases {
		if got := c.d.Iso8601(); got != c.want {
			t.Errorf("Iso8601(%s) = %q, want %q", c.d.Inspect(), got, c.want)
		}
	}
}

func TestSinceAgo(t *testing.T) {
	base := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	// Calendar advance.
	if got := Years(1).Add(Months(1)).Add(Days(1)).Since(base); !got.Equal(time.Date(2027, 2, 16, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("since = %v", got)
	}
	// Weeks + hours.
	if got := Weeks(1).Add(Hours(2)).Since(base); !got.Equal(time.Date(2026, 1, 22, 14, 0, 0, 0, time.UTC)) {
		t.Errorf("weeks since = %v", got)
	}
	// Ago.
	if got := Days(1).Ago(base); !got.Equal(time.Date(2026, 1, 14, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("ago = %v", got)
	}
	// Empty-parts branch (zero duration).
	if got := (Duration{}).Since(base); !got.Equal(base) {
		t.Errorf("empty since = %v", got)
	}
	// Fractional calendar part contributes exact seconds.
	if got := Days(1.5).Since(base); !got.Equal(base.AddDate(0, 0, 1).Add(12 * time.Hour)) {
		t.Errorf("fractional since = %v", got)
	}
}

func TestParseValid(t *testing.T) {
	cases := []struct {
		in   string
		want string // Inspect
	}{
		{"P1Y2M3DT4H5M6S", "1 year, 2 months, 3 days, 4 hours, 5 minutes, and 6 seconds"},
		{"PT90S", "90 seconds"},
		{"P1W", "1 week"},
		{"-P1D", "-1 days"},
		{"+PT1H", "1 hour"},
		{"PT1,5H", "1.5 hours"},
		{"P1YT2H", "1 year and 2 hours"},
	}
	for _, c := range cases {
		d, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", c.in, err)
			continue
		}
		if got := d.Inspect(); got != c.want {
			t.Errorf("Parse(%q).Inspect = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseValueRoundTrip(t *testing.T) {
	d, _ := Parse("PT1H")
	if d.Value() != 3600 {
		t.Errorf("value = %v", d.Value())
	}
}

func TestParseErrors(t *testing.T) {
	bad := []string{
		"",        // empty -> start with no marker (finishes empty? handled as empty duration)
		"X",       // no P marker
		"P",       // empty duration
		"PY",      // bad date component (no digits)
		"P1YT",    // empty time part
		"P1Y2W",   // weeks mixed with years
		"P1W2D",   // weeks mixed with days
		"P1.5Y2M", // only last part fractional
		"P1H",     // H is a time designator, invalid in date mode
		"PT",      // empty time
		"PTX",     // bad time component
		"P1Y2M3W", // weeks mixed with months
	}
	for _, in := range bad {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) expected error", in)
		} else if !errors.Is(err, ErrParse) {
			t.Errorf("Parse(%q) error %v is not ErrParse", in, err)
		}
	}
}

func TestScanComponentEdges(t *testing.T) {
	// Fractional with missing fractional digits.
	if _, _, _, ok := scanComponent("1.D", 0, "YMWD"); ok {
		t.Error("expected failure on empty fraction")
	}
	// Sign only, no digits.
	if _, _, _, ok := scanComponent("-D", 0, "YMWD"); ok {
		t.Error("expected failure on no digits")
	}
	// Digits but designator not allowed.
	if _, _, _, ok := scanComponent("5H", 0, "YMWD"); ok {
		t.Error("expected failure on disallowed designator")
	}
	// Digits at end with no designator.
	if _, _, _, ok := scanComponent("5", 0, "YMWD"); ok {
		t.Error("expected failure on missing designator")
	}
}
