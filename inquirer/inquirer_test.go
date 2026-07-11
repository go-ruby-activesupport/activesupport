// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package inquirer

import (
	"os/exec"
	"testing"
)

func TestStringInquirer(t *testing.T) {
	s := New("production")
	if !s.Is("production") {
		t.Error("production? should be true")
	}
	if s.Is("development") {
		t.Error("development? should be false")
	}
	if s.String() != "production" {
		t.Errorf("String() = %q", s.String())
	}
}

func TestArrayInquirer(t *testing.T) {
	a := NewArray("phone", "tablet")
	if !a.Is("phone") {
		t.Error("phone? should be true")
	}
	if a.Is("desktop") {
		t.Error("desktop? should be false")
	}
	if !a.Any("phone", "desktop") {
		t.Error("any?(phone, desktop) should be true")
	}
	if a.Any("desktop", "watch") {
		t.Error("any?(desktop, watch) should be false")
	}
	if !a.Any() {
		t.Error("any? with no candidates should be true for non-empty")
	}
	if NewArray().Any() {
		t.Error("any? should be false for empty")
	}
}

// --- differential oracle ----------------------------------------------------

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

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestOracleInquirer(t *testing.T) {
	bin := rubyWithActiveSupport(t)
	s := New("production")
	cases := []struct{ got, expr string }{
		{boolStr(s.Is("production")), `ActiveSupport::StringInquirer.new("production").production?`},
		{boolStr(s.Is("development")), `ActiveSupport::StringInquirer.new("production").development?`},
	}
	a := NewArray("phone", "tablet")
	cases = append(cases,
		struct{ got, expr string }{boolStr(a.Is("phone")), `ActiveSupport::ArrayInquirer.new([:phone, :tablet]).phone?`},
		struct{ got, expr string }{boolStr(a.Is("desktop")), `ActiveSupport::ArrayInquirer.new([:phone, :tablet]).desktop?`},
		struct{ got, expr string }{boolStr(a.Any("phone", "desktop")), `ActiveSupport::ArrayInquirer.new([:phone, :tablet]).any?("phone", "desktop")`},
		struct{ got, expr string }{boolStr(a.Any("desktop", "watch")), `ActiveSupport::ArrayInquirer.new([:phone, :tablet]).any?(:desktop, :watch)`},
	)
	for _, c := range cases {
		if want := ruby(t, bin, c.expr); c.got != want {
			t.Errorf("%s => %q, ruby %q", c.expr, c.got, want)
		}
	}
}
