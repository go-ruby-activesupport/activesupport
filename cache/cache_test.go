// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package cache

import (
	"reflect"
	"testing"
	"time"
)

// setClock installs a mutable clock and returns a pointer to advance it.
func setClock(t *testing.T, start time.Time) *time.Time {
	t.Helper()
	cur := start
	orig := now
	now = func() time.Time { return cur }
	t.Cleanup(func() { now = orig })
	return &cur
}

func TestReadWriteExistDelete(t *testing.T) {
	s := NewMemoryStore()
	if !s.Write("a", 1) {
		t.Error("write returns true")
	}
	if v, ok := s.Read("a"); !ok || v != 1 {
		t.Errorf("read a = %v %v", v, ok)
	}
	if _, ok := s.Read("missing"); ok {
		t.Error("missing should not be found")
	}
	if !s.Exist("a") {
		t.Error("exist a")
	}
	if !s.Delete("a") {
		t.Error("delete existing returns true")
	}
	if s.Delete("a") {
		t.Error("delete absent returns false")
	}
	if s.Exist("a") {
		t.Error("a should be gone")
	}
}

func TestFetch(t *testing.T) {
	s := NewMemoryStore()
	if v := s.Fetch("b", func() any { return 42 }); v != 42 {
		t.Errorf("fetch miss = %v", v)
	}
	if v, _ := s.Read("b"); v != 42 {
		t.Errorf("stored = %v", v)
	}
	if v := s.Fetch("b", func() any { return 99 }); v != 42 {
		t.Errorf("fetch hit = %v", v)
	}
	if v := s.Fetch("b", func() any { return 7 }, FetchOptions{Force: true}); v != 7 {
		t.Errorf("fetch force = %v", v)
	}
	if v, _ := s.Read("b"); v != 7 {
		t.Errorf("after force = %v", v)
	}
}

func TestIncrementDecrement(t *testing.T) {
	s := NewMemoryStore()
	s.Write("a", 1)
	if s.Increment("a") != 2 {
		t.Error("inc 1")
	}
	if s.Increment("a", 5) != 7 {
		t.Error("inc 5")
	}
	if s.Decrement("a") != 6 {
		t.Error("dec 1")
	}
	if v, _ := s.Read("a"); v != int64(6) {
		t.Errorf("read after = %v", v)
	}
	// Missing key initialises from zero.
	if s.Increment("nope") != 1 {
		t.Error("inc missing = 1")
	}
	if s.Decrement("neg2", 2) != -2 {
		t.Error("dec missing = -2")
	}
	// Non-integer value coerces to zero base.
	s.Write("str", "x")
	if s.Increment("str", 3) != 3 {
		t.Error("inc non-int base 0")
	}
}

func TestExpiry(t *testing.T) {
	clk := setClock(t, time.Unix(1000, 0))
	s := NewMemoryStore()
	s.Write("a", 1, WriteOptions{ExpiresIn: 10 * time.Second})
	if _, ok := s.Read("a"); !ok {
		t.Error("should be live")
	}
	*clk = clk.Add(11 * time.Second)
	if _, ok := s.Read("a"); ok {
		t.Error("should be expired")
	}
	// Fetch with expiry.
	*clk = time.Unix(2000, 0)
	s.Fetch("b", func() any { return 5 }, FetchOptions{ExpiresIn: 3 * time.Second})
	if !s.Exist("b") {
		t.Error("b live")
	}
	*clk = clk.Add(4 * time.Second)
	if s.Exist("b") {
		t.Error("b expired")
	}
	// Increment preserves remaining expiry.
	*clk = time.Unix(3000, 0)
	s.Write("c", 1, WriteOptions{ExpiresIn: 100 * time.Second})
	s.Increment("c")
	*clk = clk.Add(101 * time.Second)
	if s.Exist("c") {
		t.Error("c expiry should be preserved through increment")
	}
}

func TestClearAndCleanup(t *testing.T) {
	clk := setClock(t, time.Unix(0, 0))
	s := NewMemoryStore()
	s.Write("keep", 1)
	s.Write("gone", 2, WriteOptions{ExpiresIn: time.Second})
	*clk = clk.Add(2 * time.Second)
	s.Cleanup()
	if _, ok := s.data["gone"]; ok {
		t.Error("cleanup should drop expired")
	}
	if _, ok := s.data["keep"]; !ok {
		t.Error("cleanup should keep live")
	}
	s.Clear()
	if len(s.data) != 0 {
		t.Error("clear empties")
	}
}

func TestMulti(t *testing.T) {
	s := NewMemoryStore()
	s.Write("b", 7)
	got := s.ReadMulti("b", "x")
	if !reflect.DeepEqual(got, map[string]any{"b": 7}) {
		t.Errorf("read_multi = %v", got)
	}
	s.WriteMulti(map[string]any{"m": 1, "n": 2})
	if !reflect.DeepEqual(s.ReadMulti("m", "n"), map[string]any{"m": 1, "n": 2}) {
		t.Errorf("write_multi/read_multi")
	}
	fm := s.FetchMulti([]string{"p", "q"}, func(k string) any { return k + "!" })
	if !reflect.DeepEqual(fm, map[string]any{"p": "p!", "q": "q!"}) {
		t.Errorf("fetch_multi = %v", fm)
	}
	if v, _ := s.Read("p"); v != "p!" {
		t.Error("fetch_multi should store")
	}
	// Hit path and force path.
	s.Write("h", "cached")
	fm2 := s.FetchMulti([]string{"h"}, func(k string) any { return "new" })
	if fm2["h"] != "cached" {
		t.Error("fetch_multi hit")
	}
	fm3 := s.FetchMulti([]string{"h"}, func(k string) any { return "forced" }, FetchOptions{Force: true})
	if fm3["h"] != "forced" {
		t.Error("fetch_multi force")
	}
	s.WriteMulti(map[string]any{"z": 9}, WriteOptions{ExpiresIn: time.Hour})
	if !s.Exist("z") {
		t.Error("write_multi with opts")
	}
}

func TestToInt64(t *testing.T) {
	if toInt64(int64(5)) != 5 || toInt64(3) != 3 || toInt64("x") != 0 {
		t.Error("toInt64")
	}
}
