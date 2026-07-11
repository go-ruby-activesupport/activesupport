// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package cache ports ActiveSupport::Cache::MemoryStore — an in-process,
// concurrency-safe key/value cache with per-entry expiry.
//
// It mirrors the store's observable behaviour: Read/Write/Delete/Exist, Fetch
// (compute-and-store on miss, with a Force recompute), Increment/Decrement
// (initialising a missing key from zero), Clear, Cleanup (drop expired entries),
// and the multi variants ReadMulti/WriteMulti/FetchMulti. Expiry is driven by an
// injectable clock so tests are deterministic.
package cache

import (
	"sync"
	"time"
)

// now is the time source, overridable in tests (a seam).
var now = time.Now

type entry struct {
	value     any
	expiresAt time.Time
	hasExpiry bool
}

func (e entry) expired(at time.Time) bool {
	return e.hasExpiry && !at.Before(e.expiresAt)
}

// Store is an in-memory cache (ActiveSupport::Cache::MemoryStore).
type Store struct {
	mu   sync.Mutex
	data map[string]entry
}

// WriteOptions carries per-write settings.
type WriteOptions struct {
	// ExpiresIn expires the entry after the given duration (zero means never).
	ExpiresIn time.Duration
}

// FetchOptions carries per-fetch settings.
type FetchOptions struct {
	ExpiresIn time.Duration
	// Force recomputes and overwrites even on a hit.
	Force bool
}

// NewMemoryStore returns an empty store.
func NewMemoryStore() *Store {
	return &Store{data: map[string]entry{}}
}

// Write stores value under key and returns true (Cache#write).
func (s *Store) Write(key string, value any, opts ...WriteOptions) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writeLocked(key, value, firstWrite(opts))
	return true
}

func (s *Store) writeLocked(key string, value any, o WriteOptions) {
	e := entry{value: value}
	if o.ExpiresIn > 0 {
		e.expiresAt = now().Add(o.ExpiresIn)
		e.hasExpiry = true
	}
	s.data[key] = e
}

// Read returns the value for key and whether a live entry was found
// (Cache#read).
func (s *Store) Read(key string) (any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked(key)
}

func (s *Store) readLocked(key string) (any, bool) {
	e, ok := s.data[key]
	if !ok {
		return nil, false
	}
	if e.expired(now()) {
		delete(s.data, key)
		return nil, false
	}
	return e.value, true
}

// Exist reports whether a live entry exists for key (Cache#exist?).
func (s *Store) Exist(key string) bool {
	_, ok := s.Read(key)
	return ok
}

// Fetch returns the cached value for key, or computes it with block, stores it
// and returns it (Cache#fetch). With Force set, block is always run.
func (s *Store) Fetch(key string, block func() any, opts ...FetchOptions) any {
	o := firstFetch(opts)
	s.mu.Lock()
	defer s.mu.Unlock()
	if !o.Force {
		if v, ok := s.readLocked(key); ok {
			return v
		}
	}
	v := block()
	s.writeLocked(key, v, WriteOptions{ExpiresIn: o.ExpiresIn})
	return v
}

// Delete removes key, returning whether it existed as a live entry (Cache#delete).
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.readLocked(key)
	delete(s.data, key)
	return ok
}

// Increment adds amount (default 1) to the integer at key, initialising a
// missing key from zero, and returns the new value (Cache#increment).
func (s *Store) Increment(key string, amount ...int64) int64 {
	return s.modify(key, delta(amount, 1))
}

// Decrement subtracts amount (default 1) from the integer at key (Cache#decrement).
func (s *Store) Decrement(key string, amount ...int64) int64 {
	return s.modify(key, -delta(amount, 1))
}

func (s *Store) modify(key string, by int64) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	base := int64(0)
	if v, ok := s.readLocked(key); ok {
		base = toInt64(v)
	}
	nv := base + by
	// Preserve any existing expiry on the entry.
	o := WriteOptions{}
	if e, ok := s.data[key]; ok && e.hasExpiry {
		o.ExpiresIn = e.expiresAt.Sub(now())
	}
	s.writeLocked(key, nv, o)
	return nv
}

// Clear removes every entry (Cache#clear).
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = map[string]entry{}
}

// Cleanup drops expired entries (Cache#cleanup).
func (s *Store) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	at := now()
	for k, e := range s.data {
		if e.expired(at) {
			delete(s.data, k)
		}
	}
}

// ReadMulti returns the live values for the given keys, omitting misses
// (Cache#read_multi).
func (s *Store) ReadMulti(keys ...string) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[string]any{}
	for _, k := range keys {
		if v, ok := s.readLocked(k); ok {
			out[k] = v
		}
	}
	return out
}

// WriteMulti stores several entries at once (Cache#write_multi).
func (s *Store) WriteMulti(values map[string]any, opts ...WriteOptions) {
	o := firstWrite(opts)
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range values {
		s.writeLocked(k, v, o)
	}
}

// FetchMulti returns cached values for keys, computing any misses with block,
// storing them, and returning the full set (Cache#fetch_multi).
func (s *Store) FetchMulti(keys []string, block func(key string) any, opts ...FetchOptions) map[string]any {
	o := firstFetch(opts)
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[string]any{}
	for _, k := range keys {
		if !o.Force {
			if v, ok := s.readLocked(k); ok {
				out[k] = v
				continue
			}
		}
		v := block(k)
		s.writeLocked(k, v, WriteOptions{ExpiresIn: o.ExpiresIn})
		out[k] = v
	}
	return out
}

func firstWrite(opts []WriteOptions) WriteOptions {
	if len(opts) > 0 {
		return opts[0]
	}
	return WriteOptions{}
}

func firstFetch(opts []FetchOptions) FetchOptions {
	if len(opts) > 0 {
		return opts[0]
	}
	return FetchOptions{}
}

func delta(amount []int64, def int64) int64 {
	if len(amount) > 0 {
		return amount[0]
	}
	return def
}

func toInt64(v any) int64 {
	switch t := v.(type) {
	case int:
		return int64(t)
	case int64:
		return t
	default:
		return 0
	}
}
