// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package hwia

import (
	"fmt"
	"os/exec"
	"testing"
)

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

// TestOracleHWIA confirms the observable behaviours against MRI's
// HashWithIndifferentAccess.
func TestOracleHWIA(t *testing.T) {
	bin := rubyWithActiveSupport(t)
	// h = { a: 1, b: { c: 2 } }
	build := `ActiveSupport::HashWithIndifferentAccess.new({a: 1, b: {c: 2}})`
	h := NewFrom(map[string]any{"a": 1, "b": map[string]any{"c": 2}})

	cases := []struct{ got, expr string }{
		{fmt.Sprint(h.Get("a")), build + `["a"].to_s`},
		{fmt.Sprint(h.Get("b").(*Hash).Get("c")), build + `[:b][:c].to_s`},
		{fmt.Sprint(h.KeyQ("a")), build + `.key?(:a).to_s`},
		{fmt.Sprint(h.KeyQ("z")), build + `.key?("z").to_s`},
		{fmt.Sprint(h.FetchDefault("z", 9)), build + `.fetch(:z, 9).to_s`},
		{fmt.Sprint(h.Keys()), rubyArray(build + `.keys`)},
		{fmt.Sprint(h.Slice("a").Keys()), rubyArray(build + `.slice(:a).keys`)},
		{fmt.Sprint(h.Except("a").Keys()), rubyArray(build + `.except(:a).keys`)},
		{fmt.Sprint(h.Merge(NewFrom(map[string]any{"d": 4})).Keys()), rubyArray(build + `.merge(d: 4).keys`)},
	}
	for _, c := range cases {
		if want := ruby(t, bin, c.expr); c.got != want {
			t.Errorf("%s => %q, ruby %q", c.expr, c.got, want)
		}
	}
}

// rubyArray renders a Ruby array of strings as Go's fmt.Sprint of a []string
// would (e.g. [a b c]).
func rubyArray(expr string) string {
	return `"[" + (` + expr + `).map(&:to_s).join(" ") + "]"`
}
