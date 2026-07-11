// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package cache

import (
	"fmt"
	"os/exec"
	"strings"
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

func rubyScript(t *testing.T, bin, script string) string {
	t.Helper()
	full := "$stdout.binmode\nrequire \"active_support\"\nrequire \"active_support/all\"\n" + script
	out, err := exec.Command(bin, "-e", full).CombinedOutput()
	if err != nil {
		t.Fatalf("ruby error: %v\n%s", err, out)
	}
	return string(out)
}

// TestOracleMemoryStore replays a mixed operation sequence and confirms the
// observable results match MRI's MemoryStore step-for-step.
func TestOracleMemoryStore(t *testing.T) {
	bin := rubyWithActiveSupport(t)

	s := NewMemoryStore()
	var log []string
	add := func(v any) { log = append(log, fmt.Sprint(v)) }

	add(s.Fetch("b", func() any { return 42 }))
	add(s.Fetch("b", func() any { return 99 }))
	s.Write("a", 1)
	add(s.Increment("a"))
	add(s.Increment("a", 5))
	add(s.Decrement("a"))
	add(s.Increment("nope"))
	add(s.Delete("a"))
	add(s.Exist("a"))
	add(len(s.ReadMulti("b", "x")))
	goOut := strings.Join(log, ",")

	script := `
s = ActiveSupport::Cache::MemoryStore.new
log = []
log << s.fetch("b"){ 42 }
log << s.fetch("b"){ 99 }
s.write("a", 1)
log << s.increment("a")
log << s.increment("a", 5)
log << s.decrement("a")
log << s.increment("nope")
log << s.delete("a")
log << s.exist?("a")
log << s.read_multi("b","x").size
print log.join(",")
`
	if want := rubyScript(t, bin, script); goOut != want {
		t.Errorf("sequence mismatch:\n go   %q\n ruby %q", goOut, want)
	}
}
