// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package notifications ports ActiveSupport::Notifications — an instrumentation
// bus where producers Instrument named events carrying a payload and timing, and
// consumers Subscribe to event names (exact, by regexp, or all).
//
// Ruby's block form (Notifications.instrument("name", payload) { ... }) maps to
// Instrument(name, payload, func() any { ... }); subscribers receive a fully
// populated Event with Start/Finish/Duration. Even when the instrumented block
// panics, the event is still published (the payload gains an "exception" entry)
// before the panic propagates, matching Rails' ensure semantics.
package notifications

import (
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
	"time"
)

// clock is the time source, overridable in tests (a seam).
var clock = time.Now

// Event is a single instrumented occurrence.
type Event struct {
	Name          string
	Start         time.Time
	Finish        time.Time
	TransactionID string
	Payload       map[string]any
}

// Duration returns the wall-clock span of the event.
func (e Event) Duration() time.Duration { return e.Finish.Sub(e.Start) }

// DurationMs returns the event duration in fractional milliseconds, like Rails'
// Event#duration.
func (e Event) DurationMs() float64 {
	return float64(e.Finish.Sub(e.Start)) / float64(time.Millisecond)
}

// Matcher decides whether a subscription applies to an event name.
type Matcher interface{ matches(name string) bool }

type exactMatcher string

func (m exactMatcher) matches(name string) bool { return string(m) == name }

type regexpMatcher struct{ re *regexp.Regexp }

func (m regexpMatcher) matches(name string) bool { return m.re.MatchString(name) }

type allMatcher struct{}

func (allMatcher) matches(string) bool { return true }

// Exact matches events whose name equals s.
func Exact(s string) Matcher { return exactMatcher(s) }

// Pattern matches events whose name matches re.
func Pattern(re *regexp.Regexp) Matcher { return regexpMatcher{re} }

// All matches every event.
func All() Matcher { return allMatcher{} }

// Subscription is a live subscription handle used to Unsubscribe.
type Subscription struct {
	matcher Matcher
	handler func(Event)
	id      uint64
}

// Notifier is an instrumentation bus (ActiveSupport::Notifications fanout).
type Notifier struct {
	mu   sync.RWMutex
	subs []*Subscription
}

var idCounter uint64
var txCounter uint64

// New returns an empty Notifier.
func New() *Notifier { return &Notifier{} }

// Subscribe registers handler for events accepted by matcher and returns a
// handle for Unsubscribe.
func (n *Notifier) Subscribe(matcher Matcher, handler func(Event)) *Subscription {
	sub := &Subscription{matcher: matcher, handler: handler, id: atomic.AddUint64(&idCounter, 1)}
	n.mu.Lock()
	n.subs = append(n.subs, sub)
	n.mu.Unlock()
	return sub
}

// Unsubscribe removes a subscription (Notifications.unsubscribe).
func (n *Notifier) Unsubscribe(sub *Subscription) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for i, s := range n.subs {
		if s == sub {
			n.subs = append(n.subs[:i], n.subs[i+1:]...)
			return
		}
	}
}

// Subscribed registers handler for the duration of block, unsubscribing after
// (Notifications.subscribed).
func (n *Notifier) Subscribed(matcher Matcher, handler func(Event), block func()) {
	sub := n.Subscribe(matcher, handler)
	defer n.Unsubscribe(sub)
	block()
}

// Instrument runs block as a named event, publishing an Event to matching
// subscribers with the block's timing. It returns block's result. If block
// panics, the payload gains "exception"/"exception_object" entries, the event is
// still published, and the panic is re-raised.
func (n *Notifier) Instrument(name string, payload map[string]any, block func() any) (result any) {
	if payload == nil {
		payload = map[string]any{}
	}
	start := clock()
	tx := fmt.Sprintf("%d", atomic.AddUint64(&txCounter, 1))

	publish := func() {
		ev := Event{Name: name, Start: start, Finish: clock(), TransactionID: tx, Payload: payload}
		n.mu.RLock()
		matched := make([]*Subscription, 0, len(n.subs))
		for _, s := range n.subs {
			if s.matcher.matches(name) {
				matched = append(matched, s)
			}
		}
		n.mu.RUnlock()
		for _, s := range matched {
			s.handler(ev)
		}
	}

	panicked := true
	defer func() {
		if panicked {
			r := recover()
			payload["exception"] = []any{fmt.Sprintf("%T", r), fmt.Sprint(r)}
			payload["exception_object"] = r
			publish()
			panic(r)
		}
	}()
	result = block()
	panicked = false
	publish()
	return result
}

// --- package-level default bus ----------------------------------------------

var defaultNotifier = New()

// Subscribe registers handler on the default bus.
func Subscribe(matcher Matcher, handler func(Event)) *Subscription {
	return defaultNotifier.Subscribe(matcher, handler)
}

// Unsubscribe removes a subscription from the default bus.
func Unsubscribe(sub *Subscription) { defaultNotifier.Unsubscribe(sub) }

// Subscribed runs block with a temporary subscription on the default bus.
func Subscribed(matcher Matcher, handler func(Event), block func()) {
	defaultNotifier.Subscribed(matcher, handler, block)
}

// Instrument publishes a named event on the default bus.
func Instrument(name string, payload map[string]any, block func() any) any {
	return defaultNotifier.Instrument(name, payload, block)
}
