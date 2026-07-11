// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package notifications

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

func rubyScript(t *testing.T, bin, script string) string {
	t.Helper()
	full := "$stdout.binmode\nrequire \"active_support\"\nrequire \"active_support/all\"\n" + script
	out, err := exec.Command(bin, "-e", full).CombinedOutput()
	if err != nil {
		t.Fatalf("ruby error: %v\n%s", err, out)
	}
	return string(out)
}

// TestOracleNotificationsContract confirms the subscribe/instrument observable
// contract matches Rails: the subscriber sees the event name and payload value,
// and instrument returns the block's value.
func TestOracleNotificationsContract(t *testing.T) {
	bin := rubyWithActiveSupport(t)

	n := New()
	var seenName string
	var seenSQL any
	n.Subscribe(Exact("sql.active_record"), func(e Event) {
		seenName = e.Name
		seenSQL = e.Payload["sql"]
	})
	ret := n.Instrument("sql.active_record", map[string]any{"sql": "SELECT 1"}, func() any { return 7 })
	goOut := fmt.Sprintf("%s|%v|%v", seenName, seenSQL, ret)

	script := `
seen_name = nil
seen_sql = nil
ActiveSupport::Notifications.subscribe("sql.active_record") do |name, start, finish, id, payload|
  seen_name = name
  seen_sql = payload[:sql]
end
ret = ActiveSupport::Notifications.instrument("sql.active_record", sql: "SELECT 1") { 7 }
print "#{seen_name}|#{seen_sql}|#{ret}"
`
	if want := rubyScript(t, bin, script); goOut != want {
		t.Errorf("contract mismatch: go %q, ruby %q", goOut, want)
	}
}
