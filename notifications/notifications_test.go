// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package notifications

import (
	"regexp"
	"testing"
	"time"
)

// fakeClock returns a controllable clock and a restore function.
func fakeClock(t *testing.T, times ...time.Time) {
	t.Helper()
	orig := clock
	i := 0
	clock = func() time.Time {
		tm := times[i]
		if i < len(times)-1 {
			i++
		}
		return tm
	}
	t.Cleanup(func() { clock = orig })
}

func TestInstrumentAndSubscribe(t *testing.T) {
	t0 := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(5 * time.Millisecond)
	fakeClock(t, t0, t1)

	n := New()
	var got Event
	calls := 0
	n.Subscribe(Exact("sql.query"), func(e Event) { got = e; calls++ })
	n.Subscribe(Exact("other"), func(e Event) { t.Error("should not fire") })

	res := n.Instrument("sql.query", map[string]any{"sql": "SELECT 1"}, func() any { return 42 })
	if res != 42 {
		t.Errorf("result = %v", res)
	}
	if calls != 1 {
		t.Fatalf("calls = %d", calls)
	}
	if got.Name != "sql.query" || got.Payload["sql"] != "SELECT 1" {
		t.Errorf("event = %+v", got)
	}
	if got.Duration() != 5*time.Millisecond || got.DurationMs() != 5 {
		t.Errorf("duration = %v / %v", got.Duration(), got.DurationMs())
	}
	if got.TransactionID == "" {
		t.Error("missing transaction id")
	}
}

func TestNilPayload(t *testing.T) {
	n := New()
	var seen bool
	n.Subscribe(All(), func(e Event) {
		seen = true
		if e.Payload == nil {
			t.Error("payload should be non-nil")
		}
	})
	n.Instrument("evt", nil, func() any { return nil })
	if !seen {
		t.Error("subscriber not called")
	}
}

func TestMatchers(t *testing.T) {
	n := New()
	var names []string
	n.Subscribe(Pattern(regexp.MustCompile(`\.query$`)), func(e Event) { names = append(names, e.Name) })
	n.Subscribe(All(), func(e Event) {})
	n.Instrument("sql.query", nil, func() any { return nil })
	n.Instrument("cache.read", nil, func() any { return nil })
	if len(names) != 1 || names[0] != "sql.query" {
		t.Errorf("regexp matches = %v", names)
	}
}

func TestUnsubscribe(t *testing.T) {
	n := New()
	calls := 0
	sub := n.Subscribe(All(), func(e Event) { calls++ })
	n.Instrument("a", nil, func() any { return nil })
	n.Unsubscribe(sub)
	n.Instrument("a", nil, func() any { return nil })
	if calls != 1 {
		t.Errorf("calls = %d", calls)
	}
	// Unsubscribing an unknown handle is a no-op.
	n.Unsubscribe(&Subscription{})
}

func TestSubscribed(t *testing.T) {
	n := New()
	calls := 0
	n.Subscribed(All(), func(e Event) { calls++ }, func() {
		n.Instrument("inside", nil, func() any { return nil })
	})
	n.Instrument("outside", nil, func() any { return nil })
	if calls != 1 {
		t.Errorf("calls = %d", calls)
	}
}

func TestInstrumentPanicStillPublishes(t *testing.T) {
	n := New()
	var got Event
	fired := false
	n.Subscribe(All(), func(e Event) { got = e; fired = true })

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected re-raised panic")
		}
		if !fired {
			t.Error("event should have been published on panic")
		}
		if got.Payload["exception_object"] != r {
			t.Errorf("exception_object = %v, want %v", got.Payload["exception_object"], r)
		}
		exc, ok := got.Payload["exception"].([]any)
		if !ok || len(exc) != 2 {
			t.Errorf("exception entry = %v", got.Payload["exception"])
		}
	}()
	n.Instrument("boom", map[string]any{}, func() any { panic("kaboom") })
}

func TestPackageLevelDefaultBus(t *testing.T) {
	calls := 0
	sub := Subscribe(Exact("pkg.evt"), func(e Event) { calls++ })
	defer Unsubscribe(sub)
	if got := Instrument("pkg.evt", nil, func() any { return "ok" }); got != "ok" {
		t.Errorf("result = %v", got)
	}
	Subscribed(Exact("pkg.tmp"), func(e Event) { calls++ }, func() {
		Instrument("pkg.tmp", nil, func() any { return nil })
	})
	if calls != 2 {
		t.Errorf("calls = %d", calls)
	}
}
