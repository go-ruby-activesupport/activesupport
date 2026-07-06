// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package inflector

// DefaultLocale holds Rails' built-in English inflections, ported verbatim from
// activesupport's default_inflections.rb. The Ruby replacement backreferences
// (\1, \2) are written here in Go's Expand syntax (${1}, ${2}); every /.../i rule
// passes caseInsensitive=true. The rules are registered in file order; because
// each registration prepends, the resulting evaluation order matches MRI exactly.
//
// Callers that need extra rules should Clone() this instance rather than mutating
// it, so the shared default stays pristine.
var DefaultLocale = buildDefault()

func buildDefault() *Inflections {
	inf := NewInflections()

	inf.Plural(false, `$`, "s")
	inf.Plural(true, `s$`, "s")
	inf.Plural(true, `^(ax|test)is$`, "${1}es")
	inf.Plural(true, `(octop|vir)us$`, "${1}i")
	inf.Plural(true, `(octop|vir)i$`, "${1}i")
	inf.Plural(true, `(alias|status)$`, "${1}es")
	inf.Plural(true, `(bu)s$`, "${1}ses")
	inf.Plural(true, `(buffal|tomat)o$`, "${1}oes")
	inf.Plural(true, `([ti])um$`, "${1}a")
	inf.Plural(true, `([ti])a$`, "${1}a")
	inf.Plural(true, `sis$`, "ses")
	inf.Plural(true, `(?:([^f])fe|([lr])f)$`, "${1}${2}ves")
	inf.Plural(true, `(hive)$`, "${1}s")
	inf.Plural(true, `([^aeiouy]|qu)y$`, "${1}ies")
	inf.Plural(true, `(x|ch|ss|sh)$`, "${1}es")
	inf.Plural(true, `(matr|vert|ind)(?:ix|ex)$`, "${1}ices")
	inf.Plural(true, `^(m|l)ouse$`, "${1}ice")
	inf.Plural(true, `^(m|l)ice$`, "${1}ice")
	inf.Plural(true, `^(ox)$`, "${1}en")
	inf.Plural(true, `^(oxen)$`, "${1}")
	inf.Plural(true, `(quiz)$`, "${1}zes")

	inf.Singular(true, `s$`, "")
	inf.Singular(true, `(ss)$`, "${1}")
	inf.Singular(true, `(n)ews$`, "${1}ews")
	inf.Singular(true, `([ti])a$`, "${1}um")
	inf.Singular(true, `((a)naly|(b)a|(d)iagno|(p)arenthe|(p)rogno|(s)ynop|(t)he)(sis|ses)$`, "${1}sis")
	inf.Singular(true, `(^analy)(sis|ses)$`, "${1}sis")
	inf.Singular(true, `([^f])ves$`, "${1}fe")
	inf.Singular(true, `(hive)s$`, "${1}")
	inf.Singular(true, `(tive)s$`, "${1}")
	inf.Singular(true, `([lr])ves$`, "${1}f")
	inf.Singular(true, `([^aeiouy]|qu)ies$`, "${1}y")
	inf.Singular(true, `(s)eries$`, "${1}eries")
	inf.Singular(true, `(m)ovies$`, "${1}ovie")
	inf.Singular(true, `(x|ch|ss|sh)es$`, "${1}")
	inf.Singular(true, `^(m|l)ice$`, "${1}ouse")
	inf.Singular(true, `(bus)(es)?$`, "${1}")
	inf.Singular(true, `(o)es$`, "${1}")
	inf.Singular(true, `(shoe)s$`, "${1}")
	inf.Singular(true, `(cris|test)(is|es)$`, "${1}is")
	inf.Singular(true, `^(a)x[ie]s$`, "${1}xis")
	inf.Singular(true, `(octop|vir)(us|i)$`, "${1}us")
	inf.Singular(true, `(alias|status)(es)?$`, "${1}")
	inf.Singular(true, `^(ox)en`, "${1}")
	inf.Singular(true, `(vert|ind)ices$`, "${1}ex")
	inf.Singular(true, `(matr)ices$`, "${1}ix")
	inf.Singular(true, `(quiz)zes$`, "${1}")
	inf.Singular(true, `(database)s$`, "${1}")

	inf.Irregular("person", "people")
	inf.Irregular("man", "men")
	inf.Irregular("child", "children")
	inf.Irregular("sex", "sexes")
	inf.Irregular("move", "moves")
	inf.Irregular("zombie", "zombies")

	inf.Uncountable("equipment", "information", "rice", "money", "species",
		"series", "fish", "sheep", "jeans", "police")

	return inf
}
