// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package numberhelper

import (
	"math"
	"regexp"
	"testing"
)

func eq(t *testing.T, got, want, label string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", label, got, want)
	}
}

func TestNumToString(t *testing.T) {
	eq(t, numToString("x"), "x", "string")
	eq(t, numToString(42), "42", "int")
	eq(t, numToString(int64(42)), "42", "int64")
	eq(t, numToString(1.5), "1.5", "float")
	eq(t, numToString(math.Inf(1)), "Inf", "inf")
	eq(t, numToString(math.Inf(-1)), "-Inf", "-inf")
	eq(t, numToString(math.NaN()), "NaN", "nan")
	eq(t, numToString(true), "", "unsupported")
}

func TestValidNumber(t *testing.T) {
	for _, n := range []any{7, int64(7), 1.5, "1.5", "-3", "12.34"} {
		if _, ok := validNumber(n); !ok {
			t.Errorf("expected %v valid", n)
		}
	}
	for _, n := range []any{"abc", "1e5", "1d2", math.Inf(1), math.NaN(), true} {
		if _, ok := validNumber(n); ok {
			t.Errorf("expected %v invalid", n)
		}
	}
}

func TestDigitCount(t *testing.T) {
	eq2(t, digitCount(0), 1, "zero")
	eq2(t, digitCount(5), 1, "5")
	eq2(t, digitCount(50), 2, "50")
	eq2(t, digitCount(1234), 4, "1234")
	eq2(t, digitCount(0.33), 0, "0.33")
}

func eq2(t *testing.T, got, want int, label string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %d, want %d", label, got, want)
	}
}

func TestRoundModes(t *testing.T) {
	cases := []struct {
		in    string
		place int
		mode  string
		want  string
	}{
		{"2.5", 0, "half_up", "3.0"},
		{"2.5", 0, "half_even", "2.0"},
		{"3.5", 0, "half_even", "4.0"},
		{"2.6", 0, "half_even", "3.0"}, // cmp > 0
		{"2.4", 0, "half_even", "2.0"}, // cmp < 0
		{"2.5", 0, "half_down", "2.0"},
		{"2.6", 0, "half_down", "3.0"},
		{"2.1", 0, "up", "3.0"},
		{"2.9", 0, "down", "2.0"},
		{"-2.1", 0, "ceiling", "-2.0"},
		{"2.1", 0, "ceiling", "3.0"},
		{"-2.1", 0, "floor", "-3.0"},
		{"2.9", 0, "floor", "2.0"},
		{"-2.5", 0, "half_up", "-3.0"},
		{"1.243", 2, "down", "1.24"},
		{"2.0", 0, "half_up", "2.0"}, // exact, no fraction
		{"1234", -2, "half_up", "1200.0"},
		{"0.0", 2, "half_up", "0.0"},
		{"-0.001", 0, "half_up", "0.0"}, // negative zero prevented
	}
	for _, c := range cases {
		if got := roundToF(c.in, c.place, c.mode); got != c.want {
			t.Errorf("roundToF(%q,%d,%s) = %q, want %q", c.in, c.place, c.mode, got, c.want)
		}
	}
}

func TestNormalizeMode(t *testing.T) {
	pairs := map[string]string{
		"": "half_up", "default": "half_up", "banker": "half_even",
		"ceil": "ceiling", "to_zero": "down", "truncate": "down", "floor": "floor",
	}
	for in, want := range pairs {
		if got := normalizeMode(in); got != want {
			t.Errorf("normalizeMode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGroupThousands(t *testing.T) {
	eq(t, groupThousands("12", ","), "12", "short")
	eq(t, groupThousands("1234", ","), "1,234", "1234")
	eq(t, groupThousands("123456", ","), "123,456", "123456")
	eq(t, groupThousands("1234567", ","), "1,234,567", "off1")
	eq(t, groupThousands("-1234", ","), "-1,234", "neg")
	eq(t, groupThousands("1234", ""), "1234", "no delim")
}

func TestStripInsignificantZeros(t *testing.T) {
	eq(t, stripInsignificantZeros("1.200", "."), "1.2", "trailing")
	eq(t, stripInsignificantZeros("1.000", "."), "1", "all zero")
	eq(t, stripInsignificantZeros("1.23", "."), "1.23", "none")
	eq(t, stripInsignificantZeros("100", "."), "100", "no sep")
}

func TestDelimited(t *testing.T) {
	eq(t, NumberToDelimited(12345678), "12,345,678", "int")
	eq(t, NumberToDelimited(12345678.05), "12,345,678.05", "float")
	eq(t, NumberToDelimited(12345678, Options{Delimiter: StrPtr("."), Separator: StrPtr(",")}), "12.345.678", "euro")
	eq(t, NumberToDelimited("abc"), "abc", "invalid")
}

func TestRounded(t *testing.T) {
	eq(t, NumberToRounded(111.2345), "111.235", "default")
	eq(t, NumberToRounded(111.2345, Options{Precision: IntPtr(2)}), "111.23", "p2")
	eq(t, NumberToRounded(389.32314, Options{Precision: IntPtr(0)}), "389", "p0")
	eq(t, NumberToRounded(13, Options{Precision: IntPtr(5), Significant: BoolPtr(true)}), "13.000", "sig5")
	eq(t, NumberToRounded(13.0, Options{Precision: IntPtr(5), Significant: BoolPtr(true), StripInsignificantZeros: BoolPtr(true)}), "13", "sig strip")
	eq(t, NumberToRounded(1234.5678, Options{Precision: IntPtr(2), Significant: BoolPtr(true)}), "1200", "sig2")
	eq(t, NumberToRounded("nonnum"), "nonnum", "invalid")
	eq(t, NumberToRounded(math.Inf(1)), "Inf", "inf")
	eq(t, NumberToRounded(math.Inf(-1)), "-Inf", "-inf")
	eq(t, NumberToRounded(math.NaN()), "NaN", "nan")
	eq(t, NumberToRounded(math.Inf(1), Options{StripInsignificantZeros: BoolPtr(true)}), "Inf", "inf strip")
}

func TestPercentage(t *testing.T) {
	eq(t, NumberToPercentage(100), "100.000%", "default")
	eq(t, NumberToPercentage(100, Options{Precision: IntPtr(0)}), "100%", "p0")
	eq(t, NumberToPercentage(1000, Options{Delimiter: StrPtr("."), Separator: StrPtr(",")}), "1.000,000%", "euro")
	eq(t, NumberToPercentage("abc"), "abc%", "invalid")
}

func TestCurrency(t *testing.T) {
	eq(t, NumberToCurrency(1234567890.50), "$1,234,567,890.50", "default")
	eq(t, NumberToCurrency(1234567890.506, Options{Precision: IntPtr(3)}), "$1,234,567,890.506", "p3")
	eq(t, NumberToCurrency(1234567890.50, Options{Unit: StrPtr("&pound;"), Separator: StrPtr(","), Delimiter: StrPtr("")}), "&pound;1234567890,50", "pound")
	eq(t, NumberToCurrency(-1234567890.50, Options{NegativeFormat: StrPtr("(%u%n)")}), "($1,234,567,890.50)", "neg fmt")
	eq(t, NumberToCurrency(1234567890.50, Options{Format: StrPtr("%n %u")}), "1,234,567,890.50 $", "fmt")
	eq(t, NumberToCurrency("abc"), "$abc", "invalid")
	eq(t, NumberToCurrency("-abc"), "-$abc", "invalid neg")
	eq(t, NumberToCurrency(-0.0001), "$0.00", "tiny neg positive")
	eq(t, NumberToCurrency(-0.006), "-$0.01", "small neg")
	eq(t, NumberToCurrency(1234.5, Options{Significant: BoolPtr(true), Precision: IntPtr(2)}), "$1,200", "significant")
	eq(t, NumberToCurrency(-5, Options{Format: StrPtr("%u%n")}), "-$5.00", "explicit fmt neg")
}

func TestAbsFloat(t *testing.T) {
	if absFloat(-2) != 2 || absFloat(3) != 3 {
		t.Error("absFloat")
	}
}

func TestHumanSize(t *testing.T) {
	eq(t, NumberToHumanSize(1234567), "1.18 MB", "mb")
	eq(t, NumberToHumanSize(1234567890123), "1.12 TB", "tb")
	eq(t, NumberToHumanSize(483989, Options{Precision: IntPtr(2)}), "470 KB", "kb p2")
	eq(t, NumberToHumanSize(0), "0 Bytes", "zero")
	eq(t, NumberToHumanSize(123), "123 Bytes", "bytes")
	eq(t, NumberToHumanSize(1024), "1 KB", "1kb")
	eq(t, NumberToHumanSize(-1234), "-1.21 KB", "neg")
	eq(t, NumberToHumanSize("abc"), "abc", "invalid")
	// Beyond TB, the exponent is capped at TB (Rails raises; we cap).
	if got := NumberToHumanSize(1024.0 * 1024 * 1024 * 1024 * 1024 * 1024); got == "" {
		t.Error("expected capped output")
	}
}

func TestHuman(t *testing.T) {
	eq(t, NumberToHuman(123456), "123 Thousand", "thousand")
	eq(t, NumberToHuman(1234567), "1.23 Million", "million")
	eq(t, NumberToHuman(489939, Options{Precision: IntPtr(2)}), "490 Thousand", "p2")
	eq(t, NumberToHuman(1234567, Options{Precision: IntPtr(4), Significant: BoolPtr(false)}), "1.2346 Million", "nosig")
	eq(t, NumberToHuman(0), "0", "zero")
	eq(t, NumberToHuman(123), "123", "small")
	eq(t, NumberToHuman(-123456), "-123 Thousand", "neg")
	eq(t, NumberToHuman(1234567, Options{Units: map[string]string{"unit": "", "thousand": "K", "bogus": "Z"}}), "1230 K", "units override")
	eq(t, NumberToHuman(1000000000000000000000.0), "1000000 Quadrillion", "quad")
	eq(t, NumberToHuman("abc"), "abc", "invalid")
	eq(t, NumberToHuman(1234567, Options{Format: StrPtr("%n%u")}), "1.23Million", "format")
}

func TestPhone(t *testing.T) {
	eq(t, NumberToPhone(5551234), "555-1234", "basic")
	eq(t, NumberToPhone(1235551234, Options{AreaCode: true}), "(123) 555-1234", "area")
	eq(t, NumberToPhone(1235551234, Options{Delimiter: StrPtr(" ")}), "123 555 1234", "delim")
	eq(t, NumberToPhone(1235551234, Options{CountryCode: 1}), "+1-123-555-1234", "cc")
	eq(t, NumberToPhone("123a456"), "123a456", "invalid")
	eq(t, NumberToPhone(1235551234, Options{AreaCode: true, Extension: 555}), "(123) 555-1234 x 555", "ext")
	eq(t, NumberToPhone(1235551234, Options{Delimiter: StrPtr("")}), "1235551234", "no delim")
	// Custom patterns (area and no-area).
	pat := regexp.MustCompile(`([0-9]{3})([0-9]{3})([0-9]{4})$`)
	eq(t, NumberToPhone(1235551234, Options{Pattern: pat}), "123-555-1234", "custom pattern")
	eq(t, NumberToPhone(1235551234, Options{AreaCode: true, Pattern: pat}), "(123) 555-1234", "custom area pattern")
	eq(t, NumberToPhone(1235551234, Options{Extension: nil}), "123-555-1234", "nil ext")
}

func TestBlankToStr(t *testing.T) {
	eq(t, blankToStr(nil), "", "nil")
	eq(t, blankToStr(5), "5", "int")
}
