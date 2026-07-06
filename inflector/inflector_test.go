// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package inflector

import "testing"

type pair struct{ in, want string }

func TestPluralize(t *testing.T) {
	cases := []pair{
		{"post", "posts"},
		{"octopus", "octopi"},
		{"virus", "viri"},
		{"alias", "aliases"},
		{"status", "statuses"},
		{"bus", "buses"},
		{"buffalo", "buffaloes"},
		{"tomato", "tomatoes"},
		{"medium", "media"},
		{"datum", "data"},
		{"analysis", "analyses"},
		{"half", "halves"},
		{"wife", "wives"},
		{"hive", "hives"},
		{"category", "categories"},
		{"quiz", "quizzes"},
		{"matrix", "matrices"},
		{"vertex", "vertices"},
		{"index", "indices"},
		{"mouse", "mice"},
		{"louse", "lice"},
		{"ox", "oxen"},
		{"oxen", "oxen"},
		{"news", "news"},
		{"series", "series"},
		{"testis", "testes"},
		{"axis", "axes"},
		{"crisis", "crises"},
		{"shoe", "shoes"},
		{"database", "databases"},
		{"sheep", "sheep"},
		{"money", "money"},
		{"person", "people"},
		{"man", "men"},
		{"child", "children"},
		{"sex", "sexes"},
		{"move", "moves"},
		{"zombie", "zombies"},
		{"comment", "comments"},
		{"house", "houses"},
		{"foot", "foots"},
		{"fish", "fish"},
		{"equipment", "equipment"},
		{"", ""},
		{"CamelOctopus", "CamelOctopi"},
	}
	for _, c := range cases {
		if got := Pluralize(c.in); got != c.want {
			t.Errorf("Pluralize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSingularize(t *testing.T) {
	cases := []pair{
		{"posts", "post"},
		{"octopi", "octopus"},
		{"viruses", "viruse"},
		{"aliases", "alias"},
		{"statuses", "status"},
		{"buses", "bus"},
		{"buffaloes", "buffalo"},
		{"tomatoes", "tomato"},
		{"media", "medium"},
		{"data", "datum"},
		{"analyses", "analysis"},
		{"halves", "half"},
		{"wives", "wife"},
		{"hives", "hive"},
		{"categories", "category"},
		{"quizzes", "quiz"},
		{"matrices", "matrix"},
		{"vertices", "vertex"},
		{"indices", "index"},
		{"mice", "mouse"},
		{"lice", "louse"},
		{"oxen", "ox"},
		{"news", "news"},
		{"series", "series"},
		{"testes", "testis"},
		{"axes", "axis"},
		{"crises", "crisis"},
		{"shoes", "shoe"},
		{"databases", "database"},
		{"sheep", "sheep"},
		{"moneys", "money"},
		{"people", "person"},
		{"men", "man"},
		{"children", "child"},
		{"sexes", "sex"},
		{"moves", "move"},
		{"zombies", "zombie"},
		{"comments", "comment"},
		{"houses", "house"},
		{"feet", "feet"},
		{"fish", "fish"},
		{"movies", "movie"},
		{"word", "word"}, // matches no rule ⇒ returned unchanged
		{"ss", "ss"},
		{"", ""},
	}
	for _, c := range cases {
		if got := Singularize(c.in); got != c.want {
			t.Errorf("Singularize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCamelize(t *testing.T) {
	cases := []pair{
		{"active_model", "ActiveModel"},
		{"active_model/errors", "ActiveModel::Errors"},
		{"post", "Post"},
		{"", ""},
		{"raw_scaled_scorer", "RawScaledScorer"},
	}
	for _, c := range cases {
		if got := Camelize(c.in); got != c.want {
			t.Errorf("Camelize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	// lowerCamelCase.
	if got := CamelizeLower("active_model"); got != "activeModel" {
		t.Errorf("CamelizeLower = %q", got)
	}
	if got := CamelizeLower("active_model/errors"); got != "activeModel::Errors" {
		t.Errorf("CamelizeLower slash = %q", got)
	}
	if got := CamelizeLower(""); got != "" {
		t.Errorf("CamelizeLower empty = %q", got)
	}
	if got := CamelizeLower("/foo"); got != "::Foo" { // first byte non-word ⇒ unchanged before gsub
		t.Errorf("CamelizeLower /foo = %q", got)
	}
}

func TestUnderscore(t *testing.T) {
	cases := []pair{
		{"ActiveModel", "active_model"},
		{"ActiveModel::Errors", "active_model/errors"},
		{"already_snake", "already_snake"}, // no [A-Z-]|:: ⇒ early return
		{"fancyCategory", "fancy_category"},
		{"HTTPResponse", "http_response"},
		{"puni-puni", "puni_puni"},
		{"SSLError", "ssl_error"},
	}
	for _, c := range cases {
		if got := Underscore(c.in); got != c.want {
			t.Errorf("Underscore(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestHumanize(t *testing.T) {
	if got := Humanize("employee_salary"); got != "Employee salary" {
		t.Errorf("Humanize = %q", got)
	}
	if got := Humanize("author_id"); got != "Author" {
		t.Errorf("Humanize author_id = %q", got)
	}
	if got := Humanize("_id"); got != "Id" {
		t.Errorf("Humanize _id = %q", got)
	}
	if got := DefaultLocale.Humanize("author_id", false, false); got != "author" {
		t.Errorf("Humanize no-capitalize = %q", got)
	}
	if got := DefaultLocale.Humanize("author_id", true, true); got != "Author id" {
		t.Errorf("Humanize keep_id = %q", got)
	}
	if got := DefaultLocale.Humanize("", true, false); got != "" { // empty ⇒ upcaseFirstAlpha empty branch
		t.Errorf("Humanize empty = %q", got)
	}
	if got := DefaultLocale.Humanize("42", true, false); got != "42" { // digit-first ⇒ non-letter branch
		t.Errorf("Humanize digit = %q", got)
	}
}

func TestTitleize(t *testing.T) {
	cases := []pair{
		{"man from the boondocks", "Man From The Boondocks"},
		{"x-men: the last stand", "X Men: The Last Stand"},
		{"TheManWithoutAPast", "The Man Without A Past"},
		{"raiders_of_the_lost_ark", "Raiders Of The Lost Ark"},
		{"dave's code", "Dave's Code"}, // letter after in-word apostrophe not capitalized
	}
	for _, c := range cases {
		if got := Titleize(c.in); got != c.want {
			t.Errorf("Titleize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	if got := DefaultLocale.Titleize("string_ending_with_id", true); got != "String Ending With Id" {
		t.Errorf("Titleize keep_id = %q", got)
	}
}

func TestTableize(t *testing.T) {
	cases := []pair{
		{"RawScaledScorer", "raw_scaled_scorers"},
		{"ham_and_egg", "ham_and_eggs"},
		{"fancyCategory", "fancy_categories"},
	}
	for _, c := range cases {
		if got := Tableize(c.in); got != c.want {
			t.Errorf("Tableize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestClassify(t *testing.T) {
	cases := []pair{
		{"ham_and_eggs", "HamAndEgg"},
		{"posts", "Post"},
		{"calculus", "Calculu"},  // singularize quirk preserved
		{"public.posts", "Post"}, // schema prefix stripped
	}
	for _, c := range cases {
		if got := Classify(c.in); got != c.want {
			t.Errorf("Classify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestForeignKey(t *testing.T) {
	if got := ForeignKey("Message"); got != "message_id" {
		t.Errorf("ForeignKey = %q", got)
	}
	if got := ForeignKey("Admin::Post"); got != "post_id" {
		t.Errorf("ForeignKey namespaced = %q", got)
	}
	if got := DefaultLocale.ForeignKey("Message", false); got != "messageid" {
		t.Errorf("ForeignKey no-underscore = %q", got)
	}
}

func TestDasherize(t *testing.T) {
	if got := Dasherize("puni_puni"); got != "puni-puni" {
		t.Errorf("Dasherize = %q", got)
	}
}

func TestDemodulize(t *testing.T) {
	cases := []pair{
		{"ActiveSupport::Inflector::Inflections", "Inflections"},
		{"Inflections", "Inflections"},
		{"::Inflections", "Inflections"},
		{"", ""},
	}
	for _, c := range cases {
		if got := Demodulize(c.in); got != c.want {
			t.Errorf("Demodulize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDeconstantize(t *testing.T) {
	cases := []pair{
		{"Net::HTTP", "Net"},
		{"::Net::HTTP", "::Net"},
		{"String", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := Deconstantize(c.in); got != c.want {
			t.Errorf("Deconstantize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestOrdinal(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{1, "st"}, {2, "nd"}, {3, "rd"}, {4, "th"},
		{11, "th"}, {12, "th"}, {13, "th"},
		{21, "st"}, {22, "nd"}, {23, "rd"},
		{111, "th"}, {1002, "nd"}, {1003, "rd"},
		{-11, "th"}, {-1021, "st"}, {0, "th"},
	}
	for _, c := range cases {
		if got := Ordinal(c.n); got != c.want {
			t.Errorf("Ordinal(%d) = %q, want %q", c.n, got, c.want)
		}
	}
	if got := Ordinalize(1002); got != "1002nd" {
		t.Errorf("Ordinalize = %q", got)
	}
	if got := Ordinalize(-11); got != "-11th" {
		t.Errorf("Ordinalize negative = %q", got)
	}
}

func TestTransliterate(t *testing.T) {
	if got := Transliterate("plain ascii", "?"); got != "plain ascii" {
		t.Errorf("Transliterate ascii = %q", got)
	}
	if got := Transliterate("Ærøskøbing", "?"); got != "AEroskobing" {
		t.Errorf("Transliterate = %q", got)
	}
	if got := Transliterate("a字z", "?"); got != "a?z" { // CJK not in table ⇒ replacement
		t.Errorf("Transliterate miss = %q", got)
	}
	if got := Transliterate("a字z", "_"); got != "a_z" {
		t.Errorf("Transliterate custom replacement = %q", got)
	}
}

func TestParameterize(t *testing.T) {
	cases := []struct {
		in, sep string
		pres    bool
		want    string
	}{
		{"Donald E. Knuth", "-", false, "donald-e-knuth"},
		{"^très|Jolie-- ", "-", false, "tres-jolie"},
		{"Donald E. Knuth", "_", false, "donald_e_knuth"},
		{"Donald E. Knuth", "-", true, "Donald-E-Knuth"},
		{"^très|Jolie__ ", "-", false, "tres-jolie__"},
		{"^très_Jolie-- ", ".", false, "tres_jolie--"},
		{"a  b  c", "..", false, "a..b..c"}, // multi-char separator branch
		{"Foo Bar", "", false, "foobar"},    // empty separator skips collapse/trim
	}
	for _, c := range cases {
		if got := Parameterize(c.in, c.sep, c.pres); got != c.want {
			t.Errorf("Parameterize(%q, %q, %v) = %q, want %q", c.in, c.sep, c.pres, got, c.want)
		}
	}
}

func TestConstantize(t *testing.T) {
	resolve := func(name string) (any, bool) {
		if name == "Known" {
			return 42, true
		}
		return nil, false
	}
	if v, err := Constantize("Known", resolve); err != nil || v != 42 {
		t.Errorf("Constantize known = %v, %v", v, err)
	}
	v, err := Constantize("Missing", resolve)
	if err == nil || v != nil {
		t.Fatalf("Constantize missing = %v, %v", v, err)
	}
	var ne *NameError
	if ne2, ok := err.(*NameError); !ok || ne2.Name != "Missing" {
		t.Errorf("expected *NameError{Missing}, got %v", err)
	} else {
		ne = ne2
	}
	if ne.Error() != "uninitialized constant Missing" {
		t.Errorf("NameError.Error = %q", ne.Error())
	}
	if got := SafeConstantize("Known", resolve); got != 42 {
		t.Errorf("SafeConstantize known = %v", got)
	}
	if got := SafeConstantize("Missing", resolve); got != nil {
		t.Errorf("SafeConstantize missing = %v", got)
	}
}

// TestAcronyms exercises the acronym-aware paths of camelize/underscore/humanize
// on a cloned inflection set, covering every branch of the manual scanners.
func TestAcronyms(t *testing.T) {
	I := DefaultLocale.Clone()
	I.Acronym("HTML")
	I.Acronym("HTTP")
	I.Acronym("SSL")

	und := []pair{
		{"HTMLPage", "html_page"}, // acronym at position 0
		{"MyHTML", "my_html"},     // acronym preceded by alnum
		{"my_HTML", "my_html"},    // acronym preceded by '_' (no acronym match)
		{"Foo::HTML", "foo/html"}, // acronym preceded by non-word '/'
		{"HTMLish", "htm_lish"},   // acronym followed by lowercase ⇒ not treated as acronym
		{"SSLError", "ssl_error"},
	}
	for _, c := range und {
		if got := I.Underscore(c.in); got != c.want {
			t.Errorf("Underscore(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	cam := []pair{
		{"html", "HTML"},          // all-lower-digit shortcut, acronym hit
		{"html_page", "HTMLPage"}, // upperFirstToken acronym hit
		{"my_html", "MyHTML"},     // segment acronym via gsub
		{"ssl_error", "SSLError"},
		{"foo_bar", "FooBar"}, // upperFirstToken acronym miss
	}
	for _, c := range cam {
		if got := I.Camelize(c.in, true); got != c.want {
			t.Errorf("Camelize(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	// lowerFirstToken acronym-prefix path.
	if got := I.Camelize("HTMLThing", false); got != "htmlThing" {
		t.Errorf("CamelizeLower HTMLThing = %q", got)
	}
	// lowerFirstToken where the acronym is followed by a lowercase letter, so the
	// acronym branch is skipped and only the first rune is lowered.
	if got := I.Camelize("SSLes", false); got != "sSLes" {
		t.Errorf("CamelizeLower SSLes = %q, want %q", got, "sSLes")
	}

	if got := I.Humanize("ssl_error", true, false); got != "SSL error" {
		t.Errorf("Humanize acronym = %q", got)
	}
	if got := I.Titleize("ssl_error", false); got != "SSL Error" {
		t.Errorf("Titleize acronym = %q", got)
	}
}

// TestSortedAcronymsTieBreak covers the equal-length tie-break in sortedAcronyms.
func TestSortedAcronymsTieBreak(t *testing.T) {
	I := NewInflections()
	I.Acronym("SSL")
	I.Acronym("XML")
	got := I.sortedAcronyms()
	if len(got) != 2 || got[0] != "SSL" || got[1] != "XML" {
		t.Errorf("sortedAcronyms tie-break = %v", got)
	}
}
