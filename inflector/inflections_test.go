// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package inflector

import "testing"

// TestIrregularElseBranch covers Irregular's different-first-letter branch
// (cow/kine), which the built-in irregulars never exercise.
func TestIrregularElseBranch(t *testing.T) {
	I := NewInflections()
	// Give it the baseline "$"→"s" rule so unrelated words still pluralize, then a
	// cross-letter irregular.
	I.Plural(false, `$`, "s")
	I.Singular(true, `s$`, "")
	I.Irregular("cow", "kine")

	if got := I.Pluralize("cow"); got != "kine" {
		t.Errorf("Pluralize(cow) = %q, want kine", got)
	}
	if got := I.Singularize("kine"); got != "cow" {
		t.Errorf("Singularize(kine) = %q, want cow", got)
	}
	// The generated /K(?i)ine$/ rule normalizes case, exactly as MRI does.
	if got := I.Pluralize("KINE"); got != "Kine" {
		t.Errorf("Pluralize(KINE) = %q, want Kine", got)
	}
}

// TestHumanRule covers the humans-rule application in Humanize.
func TestHumanRule(t *testing.T) {
	I := DefaultLocale.Clone()
	I.Human(true, `(.*)_cnt$`, "${1}_count")
	if got := I.Humanize("jobs_cnt", true, false); got != "Jobs count" {
		t.Errorf("Humanize with human rule = %q, want %q", got, "Jobs count")
	}
}

// TestUncountableRegistration covers custom uncountables and the cached pattern.
func TestUncountableRegistration(t *testing.T) {
	I := NewInflections()
	I.Plural(false, `$`, "s")
	I.Uncountable("Foobar")
	// First call builds the cached regex; second reuses it.
	if got := I.Pluralize("foobar"); got != "foobar" {
		t.Errorf("Pluralize(foobar) = %q, want foobar", got)
	}
	if got := I.Pluralize("FOOBAR"); got != "FOOBAR" {
		t.Errorf("Pluralize(FOOBAR) = %q, want FOOBAR", got)
	}
	if got := I.Pluralize("other"); got != "others" {
		t.Errorf("Pluralize(other) = %q, want others", got)
	}
}

// TestEmptyInflectionsNoUncountables covers isUncountable's zero-length fast path.
func TestEmptyInflectionsNoUncountables(t *testing.T) {
	I := NewInflections()
	I.Plural(false, `$`, "s")
	if got := I.Pluralize("thing"); got != "things" {
		t.Errorf("Pluralize(thing) = %q, want things", got)
	}
	// No rules and no uncountables: singularize returns the word unchanged.
	empty := NewInflections()
	if got := empty.Singularize("things"); got != "things" {
		t.Errorf("empty Singularize = %q", got)
	}
}

// TestCloneIsolation ensures Clone produces an independent copy.
func TestCloneIsolation(t *testing.T) {
	c := DefaultLocale.Clone()
	c.Acronym("XML")
	c.Uncountable("bespoke")
	c.Plural(true, `^zzz$`, "zzzs")
	c.Human(true, `^x$`, "X")
	if _, ok := DefaultLocale.acronyms["xml"]; ok {
		t.Error("Clone leaked acronym into DefaultLocale")
	}
	// The default still pluralizes "bespoke" normally.
	if got := Pluralize("bespoke"); got != "bespokes" {
		t.Errorf("DefaultLocale mutated: Pluralize(bespoke) = %q", got)
	}
	if len(c.plurals) == len(DefaultLocale.plurals) {
		t.Error("Clone did not add plural rule independently")
	}
	// Cloning an instance that already has acronyms exercises the map-copy loop.
	c2 := c.Clone()
	if c2.acronyms["xml"] != "XML" {
		t.Errorf("Clone did not copy acronyms: %v", c2.acronyms)
	}
}
