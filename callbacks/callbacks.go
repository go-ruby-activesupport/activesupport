// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package callbacks ports ActiveSupport::Callbacks — before/after/around hook
// chains run around a body of work.
//
// A Chain collects Before, After and Around callbacks and runs them around a
// block with Rails' exact ordering: before callbacks in registration order,
// then the around callbacks nested (first-registered outermost) wrapping the
// block and the after callbacks (which run in reverse registration order). A
// before callback that returns false halts the chain — equivalent to Ruby's
// throw(:abort): the remaining before callbacks, the around callbacks and the
// block are skipped, but the after callbacks still run, and Run reports halted.
package callbacks

// kind identifies a callback position.
type kind int

const (
	beforeKind kind = iota
	afterKind
	aroundKind
)

type callback struct {
	kind   kind
	before func() bool
	after  func()
	around func(inner func())
}

// Chain is a set of callbacks for one event (ActiveSupport::Callbacks chain).
type Chain struct {
	callbacks []callback
}

// New returns an empty Chain.
func New() *Chain { return &Chain{} }

// Before registers a before callback. Returning false halts the chain.
func (c *Chain) Before(fn func() bool) *Chain {
	c.callbacks = append(c.callbacks, callback{kind: beforeKind, before: fn})
	return c
}

// After registers an after callback (run in reverse registration order).
func (c *Chain) After(fn func()) *Chain {
	c.callbacks = append(c.callbacks, callback{kind: afterKind, after: fn})
	return c
}

// Around registers an around callback wrapping the block; it must call inner to
// proceed. First-registered around is outermost.
func (c *Chain) Around(fn func(inner func())) *Chain {
	c.callbacks = append(c.callbacks, callback{kind: aroundKind, around: fn})
	return c
}

// Run executes the chain around block and reports whether it was halted by a
// before callback (run_callbacks). When halted, block and the around callbacks
// are skipped but the after callbacks still run.
func (c *Chain) Run(block func()) (halted bool) {
	var arounds []func(func())
	var afters []func()
	for _, cb := range c.callbacks {
		switch cb.kind {
		case aroundKind:
			arounds = append(arounds, cb.around)
		case afterKind:
			afters = append(afters, cb.after)
		}
	}

	runAfters := func() {
		for i := len(afters) - 1; i >= 0; i-- {
			afters[i]()
		}
	}

	for _, cb := range c.callbacks {
		if cb.kind == beforeKind {
			if !cb.before() {
				runAfters()
				return true
			}
		}
	}

	inner := func() {
		block()
		runAfters()
	}
	runArounds(arounds, inner)
	return false
}

// runArounds nests the around callbacks, first outermost, around inner.
func runArounds(arounds []func(func()), inner func()) {
	if len(arounds) == 0 {
		inner()
		return
	}
	arounds[0](func() { runArounds(arounds[1:], inner) })
}
