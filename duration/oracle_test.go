// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package duration

import (
	"os/exec"
	"testing"
)

// rubyWithActiveSupport gates the oracle on a ruby that can require the
// activesupport gem (installed on the CI ubuntu/macos lanes), skipping otherwise
// so the deterministic suite alone keeps coverage at 100% on the other lanes.
func rubyWithActiveSupport(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("ruby")
	if err != nil {
		t.Skip("ruby not on PATH; skipping ActiveSupport oracle")
	}
	check := `require "active_support"; require "active_support/core_ext"; print "ok"`
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

// TestOracleInspect diffs Duration#inspect byte-for-byte against MRI.
func TestOracleInspect(t *testing.T) {
	bin := rubyWithActiveSupport(t)
	cases := []struct {
		got  string
		expr string
	}{
		{Days(1).Inspect(), "1.day.inspect"},
		{Days(1).Add(Hours(2)).Inspect(), "(1.day + 2.hours).inspect"},
		{Weeks(3).Add(Days(4)).Inspect(), "(3.weeks + 4.days).inspect"},
		{Years(1).Add(Months(2)).Add(Days(3)).Inspect(), "(1.year + 2.months + 3.days).inspect"},
		{Seconds(90).Inspect(), "90.seconds.inspect"},
		{Hours(1.5).Inspect(), "1.5.hours.inspect"},
		{Days(-1).Inspect(), "(-1).day.inspect"},
		{Seconds(0).Inspect(), "0.seconds.inspect"},
		{Hours(2).Sub(Minutes(30)).Inspect(), "(2.hours - 30.minutes).inspect"},
		{Build(90).Inspect(), "ActiveSupport::Duration.build(90).inspect"},
		{Build(3661).Inspect(), "ActiveSupport::Duration.build(3661).inspect"},
		{Build(-3661).Inspect(), "ActiveSupport::Duration.build(-3661).inspect"},
	}
	for _, c := range cases {
		if want := ruby(t, bin, c.expr); c.got != want {
			t.Errorf("%s => %q, ruby %q", c.expr, c.got, want)
		}
	}
}

// TestOracleIso8601 diffs Duration#iso8601 against MRI.
func TestOracleIso8601(t *testing.T) {
	bin := rubyWithActiveSupport(t)
	cases := []struct {
		got  string
		expr string
	}{
		{Days(30).Iso8601(), "30.days.iso8601"},
		{Seconds(90).Iso8601(), "90.seconds.iso8601"},
		{Hours(1.5).Iso8601(), "1.5.hours.iso8601"},
		{Weeks(3).Add(Days(4)).Iso8601(), "(3.weeks + 4.days).iso8601"},
		{Years(1).Add(Weeks(2)).Iso8601(), "(1.year + 2.weeks).iso8601"},
		{Months(1).Add(Weeks(1)).Iso8601(), "(1.month + 1.week).iso8601"},
		{Weeks(2).Iso8601(), "2.weeks.iso8601"},
		{Seconds(-90).Iso8601(), "(-90).seconds.iso8601"},
		{Years(1).Add(Months(2)).Add(Days(3)).Add(Hours(4)).Add(Minutes(5)).Add(Seconds(6)).Iso8601(),
			"(1.year + 2.months + 3.days + 4.hours + 5.minutes + 6.seconds).iso8601"},
	}
	for _, c := range cases {
		if want := ruby(t, bin, c.expr); c.got != want {
			t.Errorf("%s => %q, ruby %q", c.expr, c.got, want)
		}
	}
}

// TestOracleParse diffs Duration.parse round-trips against MRI.
func TestOracleParse(t *testing.T) {
	bin := rubyWithActiveSupport(t)
	inputs := []string{"P1Y2M3DT4H5M6S", "PT90S", "P1W", "P1YT2H"}
	for _, in := range inputs {
		d, err := Parse(in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", in, err)
			continue
		}
		want := ruby(t, bin, `ActiveSupport::Duration.parse(`+strconvQuote(in)+`).inspect`)
		if d.Inspect() != want {
			t.Errorf("Parse(%q).Inspect = %q, ruby %q", in, d.Inspect(), want)
		}
	}
}

// TestOracleInUnits diffs the in_<unit> conversions against MRI.
func TestOracleInUnits(t *testing.T) {
	bin := rubyWithActiveSupport(t)
	if got, want := formatFloat(Minutes(90).InHours()), ruby(t, bin, "90.minutes.in_hours.to_s"); got != want {
		t.Errorf("in_hours %q vs %q", got, want)
	}
	if got, want := formatFloat(Seconds(45).InMinutes()), ruby(t, bin, "45.seconds.in_minutes.to_s"); got != want {
		t.Errorf("in_minutes %q vs %q", got, want)
	}
}

// strconvQuote returns a Ruby string literal for s (inputs are ASCII).
func strconvQuote(s string) string { return "\"" + s + "\"" }

// formatFloat renders a float like Ruby's Float#to_s for the tested conversions.
func formatFloat(f float64) string {
	// The tested values are exact ratios that Ruby prints as e.g. "1.5"/"0.75".
	return formatNum(f) + rubyFloatSuffix(f)
}

// rubyFloatSuffix appends ".0" when the value is integral, matching Float#to_s.
func rubyFloatSuffix(f float64) string {
	if f == float64(int64(f)) {
		return ".0"
	}
	return ""
}
