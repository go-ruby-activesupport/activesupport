// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package numberhelper

import (
	"os/exec"
	"testing"
)

// rubyWithActiveSupport gates the oracle on a ruby that can require the
// activesupport gem, skipping otherwise so the deterministic suite alone keeps
// coverage at 100% on the ruby-free lanes.
func rubyWithActiveSupport(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("ruby")
	if err != nil {
		t.Skip("ruby not on PATH; skipping ActiveSupport oracle")
	}
	check := `require "active_support"; require "active_support/all"; print "ok"`
	if out, err := exec.Command(path, "-e", check).CombinedOutput(); err != nil || string(out) != "ok" {
		t.Skipf("activesupport gem unavailable: %v %s", err, out)
	}
	return path
}

func ruby(t *testing.T, bin, expr string) string {
	t.Helper()
	full := "$stdout.binmode\nrequire \"active_support\"\nrequire \"active_support/all\"\nprint(" + expr + ")"
	out, err := exec.Command(bin, "-e", full).CombinedOutput()
	if err != nil {
		t.Fatalf("ruby error for %s: %v\n%s", expr, err, out)
	}
	return string(out)
}

// TestOracleNumberHelper diffs every helper byte-for-byte against MRI's
// ActiveSupport::NumberHelper.
func TestOracleNumberHelper(t *testing.T) {
	bin := rubyWithActiveSupport(t)
	const h = "ActiveSupport::NumberHelper."
	cases := []struct {
		got, expr string
	}{
		{NumberToDelimited(12345678), h + "number_to_delimited(12345678)"},
		{NumberToDelimited(12345678.05), h + "number_to_delimited(12345678.05)"},
		{NumberToDelimited(12345678, Options{Delimiter: StrPtr("."), Separator: StrPtr(",")}),
			h + `number_to_delimited(12345678, delimiter: ".", separator: ",")`},
		{NumberToRounded(111.2345), h + "number_to_rounded(111.2345)"},
		{NumberToRounded(111.2345, Options{Precision: IntPtr(2)}), h + "number_to_rounded(111.2345, precision: 2)"},
		{NumberToRounded(13, Options{Precision: IntPtr(5), Significant: BoolPtr(true)}),
			h + "number_to_rounded(13, precision: 5, significant: true)"},
		{NumberToRounded(389.32314, Options{Precision: IntPtr(0)}), h + "number_to_rounded(389.32314, precision: 0)"},
		{NumberToRounded(1.243, Options{Precision: IntPtr(2), RoundMode: "down"}),
			h + "number_to_rounded(1.243, precision: 2, round_mode: :down)"},
		{NumberToRounded(1.0/3, Options{Precision: IntPtr(2), Significant: BoolPtr(true)}),
			h + "number_to_rounded(1.0/3, precision: 2, significant: true)"},
		{NumberToRounded(1234.5678, Options{Precision: IntPtr(2), Significant: BoolPtr(true)}),
			h + "number_to_rounded(1234.5678, precision: 2, significant: true)"},
		{NumberToRounded(-2.5, Options{Precision: IntPtr(0), RoundMode: "half_up"}),
			h + "number_to_rounded(-2.5, precision: 0, round_mode: :half_up)"},
		{NumberToRounded(2.5, Options{Precision: IntPtr(0), RoundMode: "half_even"}),
			h + "number_to_rounded(2.5, precision: 0, round_mode: :half_even)"},
		{NumberToPercentage(100), h + "number_to_percentage(100)"},
		{NumberToPercentage(100, Options{Precision: IntPtr(0)}), h + "number_to_percentage(100, precision: 0)"},
		{NumberToPercentage(302.24398923423, Options{Precision: IntPtr(5)}),
			h + "number_to_percentage(302.24398923423, precision: 5)"},
		{NumberToCurrency(1234567890.50), h + "number_to_currency(1234567890.50)"},
		{NumberToCurrency(1234567890.506, Options{Precision: IntPtr(3)}),
			h + "number_to_currency(1234567890.506, precision: 3)"},
		{NumberToCurrency(1234567890.50, Options{Unit: StrPtr("&pound;"), Separator: StrPtr(","), Delimiter: StrPtr("")}),
			h + `number_to_currency(1234567890.50, unit: "&pound;", separator: ",", delimiter: "")`},
		{NumberToCurrency(-1234567890.50, Options{NegativeFormat: StrPtr("(%u%n)")}),
			h + `number_to_currency(-1234567890.50, negative_format: "(%u%n)")`},
		{NumberToCurrency(-0.0001), h + "number_to_currency(-0.0001)"},
		{NumberToCurrency(-0.006), h + "number_to_currency(-0.006)"},
		{NumberToCurrency("abc"), h + `number_to_currency("abc")`},
		{NumberToHumanSize(1234567), h + "number_to_human_size(1234567)"},
		{NumberToHumanSize(1234567890123), h + "number_to_human_size(1234567890123)"},
		{NumberToHumanSize(483989, Options{Precision: IntPtr(2)}), h + "number_to_human_size(483989, precision: 2)"},
		{NumberToHumanSize(0), h + "number_to_human_size(0)"},
		{NumberToHumanSize(1024), h + "number_to_human_size(1024)"},
		{NumberToHuman(123456), h + "number_to_human(123456)"},
		{NumberToHuman(1234567), h + "number_to_human(1234567)"},
		{NumberToHuman(489939, Options{Precision: IntPtr(2)}), h + "number_to_human(489939, precision: 2)"},
		{NumberToHuman(1234567, Options{Precision: IntPtr(4), Significant: BoolPtr(false)}),
			h + "number_to_human(1234567, precision: 4, significant: false)"},
		{NumberToHuman(1234567, Options{Units: map[string]string{"unit": "", "thousand": "K"}}),
			h + `number_to_human(1234567, units: {unit: "", thousand: "K"})`},
		{NumberToPhone(5551234), h + "number_to_phone(5551234)"},
		{NumberToPhone(1235551234, Options{AreaCode: true}), h + "number_to_phone(1235551234, area_code: true)"},
		{NumberToPhone(1235551234, Options{Delimiter: StrPtr(" ")}), h + `number_to_phone(1235551234, delimiter: " ")`},
		{NumberToPhone(1235551234, Options{CountryCode: 1}), h + "number_to_phone(1235551234, country_code: 1)"},
		{NumberToPhone(1235551234, Options{AreaCode: true, Extension: 555}),
			h + "number_to_phone(1235551234, area_code: true, extension: 555)"},
		{NumberToPhone("123a456"), h + `number_to_phone("123a456")`},
	}
	for _, c := range cases {
		if want := ruby(t, bin, c.expr); c.got != want {
			t.Errorf("%s => %q, ruby %q", c.expr, c.got, want)
		}
	}
}
