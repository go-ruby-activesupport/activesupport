// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package coreext

import "fmt"

// A Ruby Hash maps to a Go map[any]any so string and Symbol keys coexist.

// Symbol represents a Ruby Symbol (e.g. :name). It is distinct from string so
// symbolize/stringify round-trips are observable, exactly as in Ruby.
type Symbol string

// HashBlank reports whether the hash is empty (Hash#blank?).
func HashBlank(h map[any]any) bool { return len(h) == 0 }

// DeepMerge returns a new hash merging other into h recursively: when both
// values for a key are hashes they are deep-merged, otherwise other wins
// (Hash#deep_merge).
func DeepMerge(h, other map[any]any) map[any]any {
	out := shallowCopy(h)
	for k, v := range other {
		if existing, ok := out[k]; ok {
			if em, ok1 := existing.(map[any]any); ok1 {
				if vm, ok2 := v.(map[any]any); ok2 {
					out[k] = DeepMerge(em, vm)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

// DeepMergeInto merges other into h in place and returns h (Hash#deep_merge!).
func DeepMergeInto(h, other map[any]any) map[any]any {
	merged := DeepMerge(h, other)
	for k := range h {
		delete(h, k)
	}
	for k, v := range merged {
		h[k] = v
	}
	return h
}

// DeepDup returns a recursive copy of h, duplicating nested hashes and slices
// (Hash#deep_dup).
func DeepDup(h map[any]any) map[any]any {
	out := make(map[any]any, len(h))
	for k, v := range h {
		out[k] = deepDupValue(v)
	}
	return out
}

func deepDupValue(v any) any {
	switch t := v.(type) {
	case map[any]any:
		return DeepDup(t)
	case []any:
		s := make([]any, len(t))
		for i, e := range t {
			s[i] = deepDupValue(e)
		}
		return s
	default:
		return v
	}
}

func shallowCopy(h map[any]any) map[any]any {
	out := make(map[any]any, len(h))
	for k, v := range h {
		out[k] = v
	}
	return out
}

// Except returns a copy of h without the given keys (Hash#except).
func Except(h map[any]any, keys ...any) map[any]any {
	drop := make(map[any]bool, len(keys))
	for _, k := range keys {
		drop[k] = true
	}
	out := make(map[any]any, len(h))
	for k, v := range h {
		if !drop[k] {
			out[k] = v
		}
	}
	return out
}

// Slice returns a copy of h containing only the given keys that are present
// (Hash#slice).
func Slice(h map[any]any, keys ...any) map[any]any {
	out := make(map[any]any)
	for _, k := range keys {
		if v, ok := h[k]; ok {
			out[k] = v
		}
	}
	return out
}

// ReverseMerge returns other merged with h, where h's values win on conflict
// (Hash#reverse_merge).
func ReverseMerge(h, other map[any]any) map[any]any {
	out := shallowCopy(other)
	for k, v := range h {
		out[k] = v
	}
	return out
}

// DeepTransformValues returns a copy of h with fn applied to every value,
// recursing into nested hashes and slices (Hash#deep_transform_values).
func DeepTransformValues(h map[any]any, fn func(any) any) map[any]any {
	out := make(map[any]any, len(h))
	for k, v := range h {
		out[k] = deepTransformValue(v, fn)
	}
	return out
}

func deepTransformValue(v any, fn func(any) any) any {
	switch t := v.(type) {
	case map[any]any:
		return DeepTransformValues(t, fn)
	case []any:
		s := make([]any, len(t))
		for i, e := range t {
			s[i] = deepTransformValue(e, fn)
		}
		return s
	default:
		return fn(v)
	}
}

// SymbolizeKeys returns a copy of h with string keys converted to Symbol
// (Hash#symbolize_keys). Non-string keys are left unchanged.
func SymbolizeKeys(h map[any]any) map[any]any {
	out := make(map[any]any, len(h))
	for k, v := range h {
		out[symKey(k)] = v
	}
	return out
}

// StringifyKeys returns a copy of h with every key converted to its string form
// (Hash#stringify_keys).
func StringifyKeys(h map[any]any) map[any]any {
	out := make(map[any]any, len(h))
	for k, v := range h {
		out[strKey(k)] = v
	}
	return out
}

// DeepSymbolizeKeys is SymbolizeKeys applied recursively to nested hashes
// (Hash#deep_symbolize_keys).
func DeepSymbolizeKeys(h map[any]any) map[any]any {
	out := make(map[any]any, len(h))
	for k, v := range h {
		out[symKey(k)] = deepKeyTransform(v, symKey)
	}
	return out
}

// DeepStringifyKeys is StringifyKeys applied recursively to nested hashes
// (Hash#deep_stringify_keys).
func DeepStringifyKeys(h map[any]any) map[any]any {
	out := make(map[any]any, len(h))
	for k, v := range h {
		out[strKey(k)] = deepKeyTransform(v, strKey)
	}
	return out
}

func deepKeyTransform(v any, conv func(any) any) any {
	switch t := v.(type) {
	case map[any]any:
		out := make(map[any]any, len(t))
		for k, vv := range t {
			out[conv(k)] = deepKeyTransform(vv, conv)
		}
		return out
	case []any:
		s := make([]any, len(t))
		for i, e := range t {
			s[i] = deepKeyTransform(e, conv)
		}
		return s
	default:
		return v
	}
}

// symKey mirrors Ruby's `key.to_sym rescue key`: only strings (and existing
// Symbols) become Symbols; other keys are kept.
func symKey(k any) any {
	switch t := k.(type) {
	case string:
		return Symbol(t)
	case Symbol:
		return t
	default:
		return k
	}
}

// strKey mirrors Ruby's `key.to_s`.
func strKey(k any) any {
	switch t := k.(type) {
	case string:
		return t
	case Symbol:
		return string(t)
	default:
		return fmt.Sprintf("%v", k)
	}
}

// AssertValidKeys returns an error if h contains a key not in valid, mirroring
// Hash#assert_valid_keys (which raises ArgumentError).
func AssertValidKeys(h map[any]any, valid ...any) error {
	allowed := make(map[any]bool, len(valid))
	for _, k := range valid {
		allowed[k] = true
	}
	for k := range h {
		if !allowed[k] {
			return fmt.Errorf("Unknown key: %v. Valid keys are: %s", k, joinKeys(valid))
		}
	}
	return nil
}

func joinKeys(keys []any) string {
	out := ""
	for i, k := range keys {
		if i > 0 {
			out += ", "
		}
		out += fmt.Sprintf("%v", k)
	}
	return out
}
