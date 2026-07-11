// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package hwia

import (
	"fmt"

	"github.com/go-ruby-activesupport/activesupport/coreext"
)

// OrderedOptions ports ActiveSupport::OrderedOptions: a symbol-keyed hash whose
// Ruby dot-accessors (o.boy = "John"; o.boy) become Set/Get, and whose bang
// accessor (o.dog!) — which raises when the value is blank — becomes GetBang.
// Missing keys read as nil, never an error, matching Rails.
type OrderedOptions struct {
	h *Hash
}

// NewOrderedOptions returns an empty OrderedOptions.
func NewOrderedOptions() *OrderedOptions { return &OrderedOptions{h: New()} }

// Set stores value under key (the "o.key = value" accessor).
func (o *OrderedOptions) Set(key string, value any) { o.h.Set(key, value) }

// Get returns the value for key, or nil when absent (the "o.key" accessor).
func (o *OrderedOptions) Get(key string) any { return o.h.Get(key) }

// GetBang returns the value for key, or an error when it is blank (the "o.key!"
// accessor, which raises KeyError ":key is blank" in Rails).
func (o *OrderedOptions) GetBang(key string) (any, error) {
	v := o.h.Get(key)
	if coreext.Blank(v) {
		return nil, fmt.Errorf(":%s is blank", key)
	}
	return v, nil
}

// KeyQ reports whether key is present (Hash#key?).
func (o *OrderedOptions) KeyQ(key string) bool { return o.h.KeyQ(key) }

// Delete removes key, returning its old value and whether it existed.
func (o *OrderedOptions) Delete(key string) (any, bool) { return o.h.Delete(key) }

// Keys returns the keys in insertion order.
func (o *OrderedOptions) Keys() []string { return o.h.Keys() }
