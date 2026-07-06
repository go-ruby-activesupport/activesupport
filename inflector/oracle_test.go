// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package inflector

import (
	"os/exec"
	"strings"
	"testing"
)

// rubyWithActiveSupport locates a ruby that can `require "active_support"`, and
// skips the test otherwise. This is the MRI differential oracle: on the CI
// ubuntu/macos lanes the activesupport gem is installed, so these run; on the
// qemu cross-arch lanes and Windows (no target-arch ruby/gem) they skip, and the
// deterministic suite alone holds coverage at 100%.
func rubyWithActiveSupport(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("ruby")
	if err != nil {
		t.Skip("ruby not on PATH; skipping ActiveSupport oracle")
	}
	check := `require "active_support"; require "active_support/inflector/transliterate"; print "ok"`
	if out, err := exec.Command(path, "-e", check).CombinedOutput(); err != nil || string(out) != "ok" {
		t.Skipf("activesupport gem unavailable: %v %s", err, out)
	}
	return path
}

// rubyLines runs a Ruby script and returns its stdout lines.
func rubyLines(t *testing.T, bin, script string) []string {
	t.Helper()
	full := "$stdout.binmode\nrequire \"active_support\"\nrequire \"active_support/inflector/transliterate\"\nI = ActiveSupport::Inflector\n" + script
	out, err := exec.Command(bin, "-e", full).CombinedOutput()
	if err != nil {
		t.Fatalf("ruby error: %v\n%s", err, out)
	}
	return strings.Split(strings.TrimRight(string(out), "\n"), "\n")
}

// oracleWords drives the word-transform methods; oraclePhrases drives the
// case/namespace transforms. Together they hit every default inflection rule.
var oracleWords = []string{
	"post", "octopus", "virus", "alias", "status", "bus", "buffalo", "tomato",
	"medium", "datum", "analysis", "half", "wife", "hive", "category", "quiz",
	"matrix", "vertex", "index", "mouse", "louse", "ox", "oxen", "news", "series",
	"testis", "axis", "crisis", "shoe", "database", "sheep", "money", "person",
	"man", "child", "sex", "move", "zombie", "comment", "house", "foot", "fish",
	"CamelOctopus",
}

var oraclePhrases = []string{
	"active_record", "active_record/errors", "ActiveRecord", "ActiveRecord::Errors",
	"raw_scaled_scorer", "fancyCategory", "ham_and_egg", "employee_salary",
	"author_id", "_id", "SSLError", "TheManWithoutAPast", "x-men: the last stand",
	"raiders_of_the_lost_ark", "puni_puni", "Message", "Admin::Post", "Net::HTTP",
	"::Net::HTTP", "String", "ActiveSupport::Inflector::Inflections", "Ærøskøbing",
	"Donald E. Knuth", "^très|Jolie-- ",
}

func TestOracleInflector(t *testing.T) {
	bin := rubyWithActiveSupport(t)

	// Build one Ruby script that prints every transform, tab-separated, in the
	// same order our Go loop produces, then compare line by line.
	var sb strings.Builder
	sb.WriteString("W=%w[" + strings.Join(oracleWords, " ") + "]\n")
	sb.WriteString("W.each{|w| puts I.pluralize(w)}\n")
	sb.WriteString("W.each{|w| puts I.singularize(w)}\n")
	// Phrases are passed via a heredoc-free array literal built from %q to keep
	// spaces and punctuation intact.
	quoted := make([]string, len(oraclePhrases))
	for i, p := range oraclePhrases {
		quoted[i] = "%q{" + p + "}"
	}
	sb.WriteString("P=[" + strings.Join(quoted, ",") + "]\n")
	for _, m := range []string{
		"camelize(x)", "camelize(x,false)", "underscore(x)", "humanize(x)",
		"titleize(x)", "tableize(x)", "classify(x)", "dasherize(x)",
		"demodulize(x)", "deconstantize(x)", "foreign_key(x)",
		"parameterize(x)", "transliterate(x)",
	} {
		sb.WriteString("P.each{|x| puts I." + m + "}\n")
	}
	sb.WriteString("(-25..135).each{|n| puts I.ordinalize(n)}\n")

	want := rubyLines(t, bin, sb.String())

	var got []string
	for _, w := range oracleWords {
		got = append(got, Pluralize(w))
	}
	for _, w := range oracleWords {
		got = append(got, Singularize(w))
	}
	apply := []func(string) string{
		func(x string) string { return Camelize(x) },
		func(x string) string { return CamelizeLower(x) },
		Underscore, Humanize, Titleize, Tableize, Classify, Dasherize,
		Demodulize, Deconstantize, ForeignKey,
		func(x string) string { return Parameterize(x, "-", false) },
		func(x string) string { return Transliterate(x, "?") },
	}
	for _, fn := range apply {
		for _, p := range oraclePhrases {
			got = append(got, fn(p))
		}
	}
	for n := -25; n <= 135; n++ {
		got = append(got, Ordinalize(n))
	}

	if len(got) != len(want) {
		t.Fatalf("line count mismatch: go=%d ruby=%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("line %d: go=%q ruby=%q", i, got[i], want[i])
		}
	}
}
