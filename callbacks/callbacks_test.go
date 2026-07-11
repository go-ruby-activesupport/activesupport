// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package callbacks

import (
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestFullChainOrder(t *testing.T) {
	var log []string
	c := New().
		Before(func() bool { log = append(log, "before_a"); return true }).
		Before(func() bool { log = append(log, "before_b"); return true }).
		Around(func(inner func()) { log = append(log, "around_c_in"); inner(); log = append(log, "around_c_out") }).
		After(func() { log = append(log, "after_d") }).
		After(func() { log = append(log, "after_e") })

	if halted := c.Run(func() { log = append(log, "BLOCK") }); halted {
		t.Error("should not halt")
	}
	want := []string{"before_a", "before_b", "around_c_in", "BLOCK", "after_e", "after_d", "around_c_out"}
	if !reflect.DeepEqual(log, want) {
		t.Errorf("order = %v", log)
	}
}

func TestHalt(t *testing.T) {
	var log []string
	c := New().
		Before(func() bool { log = append(log, "b1"); return true }).
		Before(func() bool { log = append(log, "b2"); return false }).
		Before(func() bool { log = append(log, "b3"); return true }).
		After(func() { log = append(log, "a1") }).
		Around(func(inner func()) { log = append(log, "ar_in"); inner(); log = append(log, "ar_out") })

	if halted := c.Run(func() { log = append(log, "BLOCK") }); !halted {
		t.Error("should halt")
	}
	want := []string{"b1", "b2", "a1"}
	if !reflect.DeepEqual(log, want) {
		t.Errorf("halt order = %v", log)
	}
}

func TestNestedArounds(t *testing.T) {
	var log []string
	c := New().
		Around(func(inner func()) { log = append(log, "o1_in"); inner(); log = append(log, "o1_out") }).
		Around(func(inner func()) { log = append(log, "o2_in"); inner(); log = append(log, "o2_out") })
	c.Run(func() { log = append(log, "BLOCK") })
	want := []string{"o1_in", "o2_in", "BLOCK", "o2_out", "o1_out"}
	if !reflect.DeepEqual(log, want) {
		t.Errorf("nested arounds = %v", log)
	}
}

func TestEmptyChain(t *testing.T) {
	ran := false
	if halted := New().Run(func() { ran = true }); halted || !ran {
		t.Error("empty chain should just run the block")
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

func rubyScript(t *testing.T, bin, script string) string {
	t.Helper()
	full := "$stdout.binmode\nrequire \"active_support\"\nrequire \"active_support/all\"\n" + script
	out, err := exec.Command(bin, "-e", full).CombinedOutput()
	if err != nil {
		t.Fatalf("ruby error: %v\n%s", err, out)
	}
	return string(out)
}

func TestOracleCallbackOrder(t *testing.T) {
	bin := rubyWithActiveSupport(t)

	// Go: full chain order.
	var log []string
	New().
		Before(func() bool { log = append(log, "before_a"); return true }).
		Before(func() bool { log = append(log, "before_b"); return true }).
		Around(func(inner func()) { log = append(log, "around_c_in"); inner(); log = append(log, "around_c_out") }).
		After(func() { log = append(log, "after_d") }).
		After(func() { log = append(log, "after_e") }).
		Run(func() { log = append(log, "BLOCK") })
	goOrder := strings.Join(log, ",")

	script := `
class Rec
  include ActiveSupport::Callbacks
  define_callbacks :save
  attr_reader :log
  def initialize; @log=[]; end
  set_callback :save, :before, -> { log << "before_a" }
  set_callback :save, :before, -> { log << "before_b" }
  set_callback :save, :around, ->(o, blk){ o.log << "around_c_in"; blk.call; o.log << "around_c_out" }
  set_callback :save, :after,  -> { log << "after_d" }
  set_callback :save, :after,  -> { log << "after_e" }
  def save; run_callbacks(:save){ log << "BLOCK" }; end
end
r = Rec.new; r.save; print r.log.join(",")
`
	if want := rubyScript(t, bin, script); goOrder != want {
		t.Errorf("order mismatch: go %q ruby %q", goOrder, want)
	}

	// Go: halted chain.
	var hlog []string
	New().
		Before(func() bool { hlog = append(hlog, "b1"); return true }).
		Before(func() bool { hlog = append(hlog, "b2"); return false }).
		Before(func() bool { hlog = append(hlog, "b3"); return true }).
		After(func() { hlog = append(hlog, "a1") }).
		Around(func(inner func()) { hlog = append(hlog, "ar_in"); inner(); hlog = append(hlog, "ar_out") }).
		Run(func() { hlog = append(hlog, "BLOCK") })
	goHalt := strings.Join(hlog, ",")

	haltScript := `
class Rec2
  include ActiveSupport::Callbacks
  define_callbacks :save
  attr_reader :log
  def initialize; @log=[]; end
  set_callback :save, :before, -> { log << "b1" }
  set_callback :save, :before, -> { log << "b2"; throw :abort }
  set_callback :save, :before, -> { log << "b3" }
  set_callback :save, :after,  -> { log << "a1" }
  set_callback :save, :around, ->(o,blk){ o.log << "ar_in"; blk.call; o.log << "ar_out" }
  def save; run_callbacks(:save){ log << "BLOCK" }; end
end
r = Rec2.new; r.save; print r.log.join(",")
`
	if want := rubyScript(t, bin, haltScript); goHalt != want {
		t.Errorf("halt mismatch: go %q ruby %q", goHalt, want)
	}
}
