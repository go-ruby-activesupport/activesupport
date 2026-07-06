// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package coreext

import (
	"os/exec"
	"strings"
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
	full := "$stdout.binmode\nrequire \"active_support\"\nrequire \"active_support/core_ext\"\nprint(" + expr + ")"
	out, err := exec.Command(bin, "-e", full).CombinedOutput()
	if err != nil {
		t.Fatalf("ruby error for %s: %v\n%s", expr, err, out)
	}
	return string(out)
}

// TestOracleCoreExt diffs the String and Array helpers against MRI byte-for-byte.
func TestOracleCoreExt(t *testing.T) {
	bin := rubyWithActiveSupport(t)

	cases := []struct {
		got  string
		expr string
	}{
		{Squish("  foo\n\t bar   baz  "), `"  foo\n\t bar   baz  ".squish`},
		{Truncate("Once upon a time in a world far far away", 27, "", ""),
			`"Once upon a time in a world far far away".truncate(27)`},
		{Truncate("Once upon a time in a world far far away", 27, "", " "),
			`"Once upon a time in a world far far away".truncate(27, separator: " ")`},
		{Truncate("Once upon a time", 10, "…", " "),
			`"Once upon a time".truncate(10, omission: "…", separator: " ")`},
		{TruncateWords("Once upon a time in a world far far away", 4, ""),
			`"Once upon a time in a world far far away".truncate_words(4)`},
		{First("hello", 3), `"hello".first(3)`},
		{Last("hello", 3), `"hello".last(3)`},
		{To("hello", 2), `"hello".to(2)`},
		{Remove("foo bar foo", "foo "), `"foo bar foo".remove("foo ")`},
		{Indent("foo\n  bar", 2, "", false), `"foo\n  bar".indent(2)`},
		{StripHeredoc("    line1\n      line2\n    line3\n"),
			`"    line1\n      line2\n    line3\n".strip_heredoc`},
		{ToSentence([]any{"a", "b", "c"}), `["a","b","c"].to_sentence`},
		{ToSentence([]any{"a", "b"}), `["a","b"].to_sentence`},
		{ToSentence([]any{"a"}), `["a"].to_sentence`},
		{Parameterize("Donald E. Knuth"), `"Donald E. Knuth".parameterize`},
		{Pluralize("octopus"), `"octopus".pluralize`},
	}
	for _, c := range cases {
		if want := ruby(t, bin, c.expr); c.got != want {
			t.Errorf("%s: go=%q ruby=%q", c.expr, c.got, want)
		}
	}

	// String#at returns nil out of range; compare its inspect.
	if _, ok := At("hello", 99); ok {
		t.Error("At 99 should be OOB")
	}
	if want := ruby(t, bin, `"hello".at(99).inspect`); want != "nil" {
		t.Errorf("At OOB oracle mismatch: %q", want)
	}

	// in_groups_of rendered as a flat string for a byte-for-byte compare.
	gotGroups := renderGroups(InGroupsOf([]any{1, 2, 3, 4, 5, 6, 7}, 3, nil))
	if want := ruby(t, bin, `[1,2,3,4,5,6,7].in_groups_of(3).map{|g| g.map{|x| x.nil? ? "_" : x }.join(",")}.join("|")`); gotGroups != want {
		t.Errorf("in_groups_of oracle: go=%q ruby=%q", gotGroups, want)
	}
}

func renderGroups(groups [][]any) string {
	rows := make([]string, len(groups))
	for i, g := range groups {
		cells := make([]string, len(g))
		for j, v := range g {
			if v == nil {
				cells[j] = "_"
			} else {
				cells[j] = toS(v)
			}
		}
		rows[i] = strings.Join(cells, ",")
	}
	return strings.Join(rows, "|")
}
