// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package hwia ports ActiveSupport::HashWithIndifferentAccess and
// ActiveSupport::OrderedOptions.
//
// A Hash stores every key as a string, so string and symbol access are
// interchangeable — the "indifferent" access Rails is named for. Since Go has no
// Symbol distinct from string, keys are plain strings and the indifference is
// intrinsic. On assignment, nested Ruby hashes (map[string]any / map[any]any)
// and arrays of hashes are recursively converted to *Hash, exactly as Rails'
// convert_value does, so h.Get("a").(*Hash).Get("b") works after setting a
// nested map. Insertion order is preserved (Ruby hashes are ordered).
package hwia

import "github.com/go-ruby-activesupport/activesupport/coreext"

// Hash is a string-keyed, insertion-ordered map with indifferent (string/symbol)
// access and recursive nested-hash conversion (HashWithIndifferentAccess).
type Hash struct {
	m    map[string]any
	keys []string
}

// New returns an empty Hash.
func New() *Hash {
	return &Hash{m: map[string]any{}}
}

// NewFrom builds a Hash from a plain map, converting values recursively.
func NewFrom(src map[string]any) *Hash {
	h := New()
	for k, v := range src {
		h.Set(k, v)
	}
	return h
}

// convertValue recursively converts nested hashes to *Hash and maps over arrays,
// matching HashWithIndifferentAccess#convert_value.
func convertValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return NewFrom(t)
	case map[any]any:
		converted := map[string]any{}
		var order []string
		for k, val := range t {
			key := stringifyKey(k)
			if _, ok := converted[key]; !ok {
				order = append(order, key)
			}
			converted[key] = val
		}
		h := New()
		for _, k := range order {
			h.Set(k, converted[k])
		}
		return h
	case *Hash:
		return t
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = convertValue(e)
		}
		return out
	default:
		return v
	}
}

// stringifyKey renders a map key (string or coreext.Symbol) as a string.
func stringifyKey(k any) string {
	switch t := k.(type) {
	case string:
		return t
	case coreext.Symbol:
		return string(t)
	default:
		return ""
	}
}

// Set stores value under key (converting nested hashes), preserving order.
func (h *Hash) Set(key string, value any) {
	if _, ok := h.m[key]; !ok {
		h.keys = append(h.keys, key)
	}
	h.m[key] = convertValue(value)
}

// Get returns the value for key, or nil when absent.
func (h *Hash) Get(key string) any { return h.m[key] }

// Fetch returns the value for key and whether it was present.
func (h *Hash) Fetch(key string) (any, bool) {
	v, ok := h.m[key]
	return v, ok
}

// FetchDefault returns the value for key, or def when absent (Hash#fetch(k, d)).
func (h *Hash) FetchDefault(key string, def any) any {
	if v, ok := h.m[key]; ok {
		return v
	}
	return def
}

// KeyQ reports whether key is present (Hash#key?).
func (h *Hash) KeyQ(key string) bool {
	_, ok := h.m[key]
	return ok
}

// Delete removes key, returning its old value and whether it existed.
func (h *Hash) Delete(key string) (any, bool) {
	v, ok := h.m[key]
	if !ok {
		return nil, false
	}
	delete(h.m, key)
	for i, k := range h.keys {
		if k == key {
			h.keys = append(h.keys[:i], h.keys[i+1:]...)
			break
		}
	}
	return v, true
}

// Len returns the number of entries.
func (h *Hash) Len() int { return len(h.keys) }

// Keys returns the keys in insertion order.
func (h *Hash) Keys() []string {
	out := make([]string, len(h.keys))
	copy(out, h.keys)
	return out
}

// Values returns the values in insertion order.
func (h *Hash) Values() []any {
	out := make([]any, len(h.keys))
	for i, k := range h.keys {
		out[i] = h.m[k]
	}
	return out
}

// Each iterates entries in insertion order.
func (h *Hash) Each(fn func(key string, value any)) {
	for _, k := range h.keys {
		fn(k, h.m[k])
	}
}

// Dup returns a shallow copy (values shared).
func (h *Hash) Dup() *Hash {
	out := New()
	for _, k := range h.keys {
		out.keys = append(out.keys, k)
		out.m[k] = h.m[k]
	}
	return out
}

// Merge returns a new Hash with other's entries overriding the receiver's
// (Hash#merge).
func (h *Hash) Merge(other *Hash) *Hash {
	out := h.Dup()
	other.Each(func(k string, v any) { out.Set(k, v) })
	return out
}

// Update merges other into the receiver in place and returns it (Hash#update).
func (h *Hash) Update(other *Hash) *Hash {
	other.Each(func(k string, v any) { h.Set(k, v) })
	return h
}

// Slice returns a new Hash with only the given (present) keys (Hash#slice).
func (h *Hash) Slice(keys ...string) *Hash {
	out := New()
	for _, k := range keys {
		if v, ok := h.m[k]; ok {
			out.Set(k, v)
		}
	}
	return out
}

// Except returns a new Hash without the given keys (Hash#except).
func (h *Hash) Except(keys ...string) *Hash {
	drop := map[string]bool{}
	for _, k := range keys {
		drop[k] = true
	}
	out := New()
	for _, k := range h.keys {
		if !drop[k] {
			out.Set(k, h.m[k])
		}
	}
	return out
}

// ValuesAt returns the values for the given keys, in order (nil for absent keys).
func (h *Hash) ValuesAt(keys ...string) []any {
	out := make([]any, len(keys))
	for i, k := range keys {
		out[i] = h.m[k]
	}
	return out
}

// ToHash returns a plain map copy with nested *Hash values recursively flattened
// back to map[string]any (HashWithIndifferentAccess#to_hash).
func (h *Hash) ToHash() map[string]any {
	out := make(map[string]any, len(h.keys))
	for _, k := range h.keys {
		out[k] = toPlain(h.m[k])
	}
	return out
}

func toPlain(v any) any {
	switch t := v.(type) {
	case *Hash:
		return t.ToHash()
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = toPlain(e)
		}
		return out
	default:
		return v
	}
}
