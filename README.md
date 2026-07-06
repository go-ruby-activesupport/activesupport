<p align="center"><img src="https://raw.githubusercontent.com/go-ruby-activesupport/brand/main/social/go-ruby-activesupport-activesupport.png" alt="go-ruby-activesupport/activesupport" width="720"></p>

# activesupport — go-ruby-activesupport

[![Docs](https://img.shields.io/badge/docs-mkdocs--material-DC2626)](https://go-ruby-activesupport.github.io/docs/)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26.4%2B-00ADD8)](https://go.dev/dl/)
[![Coverage](https://img.shields.io/badge/coverage-100%25-1a7f37)](#tests--coverage)

**A pure-Go (no cgo) reimplementation of Ruby on Rails'
[ActiveSupport](https://guides.rubyonrails.org/active_support_core_extensions.html)**
— faithful to MRI down to byte-for-byte output, with **no Ruby runtime**.

ActiveSupport is enormous. This is the **v0.1 foundation**: the self-contained,
highest-value core that most other gems depend on — the **Inflector** and the
**core-extension helpers** — with the rest explicitly deferred and roadmapped
below. It is the ActiveSupport backend for
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby), but is a
**standalone, reusable** module — a sibling of
[go-ruby-set](https://github.com/go-ruby-set/set),
[go-ruby-regexp](https://github.com/go-ruby-regexp/regexp) and
[go-ruby-erb](https://github.com/go-ruby-erb/erb).

> **MRI-faithful, not Composition-Oriented.** This mirrors Rails' *observable*
> behaviour, quirks and all (`pluralize("virus") == "viri"`,
> `classify("calculus") == "Calculu"`). Reach for this when you want Rails
> semantics.

## What's in v0.1

### `inflector` — `ActiveSupport::Inflector`

The keystone that dozens of gems build on. Ported **verbatim** from the
activesupport gem's `default_inflections.rb`, so the rules, irregulars,
uncountables and the transliteration table match Rails exactly.

| Area | Functions |
| --- | --- |
| Number | `Pluralize`, `Singularize` |
| Case | `Camelize`, `CamelizeLower`, `Underscore`, `Dasherize` |
| Display | `Humanize`, `Titleize` |
| Rails naming | `Tableize`, `Classify`, `ForeignKey` |
| Constants | `Demodulize`, `Deconstantize`, `Constantize`, `SafeConstantize` |
| Numbers | `Ordinal`, `Ordinalize` |
| URLs / ASCII | `Parameterize`, `Transliterate` |
| Registration | `Inflections.Plural`/`Singular`/`Irregular`/`Uncountable`/`Acronym`/`Human`, `Clone` |

```go
import "github.com/go-ruby-activesupport/activesupport/inflector"

inflector.Pluralize("octopus")        // "octopi"
inflector.Singularize("mice")         // "mouse"
inflector.Camelize("active_model")    // "ActiveModel"
inflector.Underscore("HTTPResponse")  // "http_response"
inflector.Titleize("x-men: the last stand") // "X Men: The Last Stand"
inflector.Tableize("RawScaledScorer") // "raw_scaled_scorers"
inflector.Classify("posts")           // "Post"
inflector.ForeignKey("Admin::Post")   // "post_id"
inflector.Ordinalize(1002)            // "1002nd"
inflector.Parameterize("Donald E. Knuth", "-", false) // "donald-e-knuth"
inflector.Transliterate("Ærøskøbing", "?")            // "AEroskobing"

// Custom rules — Clone the default locale so the shared instance stays pristine.
inf := inflector.DefaultLocale.Clone()
inf.Acronym("HTML")
inf.Underscore("MyHTML")  // "my_html"
inf.Camelize("my_html", true) // "MyHTML"
```

**Seams.** Ruby object semantics live behind explicit, injectable seams so the
package stays runtime-free:

- `Constantize(name, resolve)` / `SafeConstantize(name, resolve)` take a
  `Resolver func(name string) (any, bool)` — the rbgo binding supplies one backed
  by the Ruby object space; tests supply a fake.

### `coreext` — core-extension helpers

The most-used monkey-patch helpers, as plain Go functions the rbgo binding maps
onto the Ruby methods. Representation: Ruby `String → string`, `Array → []any`,
`Hash → map[any]any` (with a `Symbol` type so string/symbol keys coexist).

- **String**: `StringBlank`/`StringPresent`/`StringPresence`, `Squish`,
  `StripHeredoc`, `Truncate`, `TruncateWords`, `First`/`Last`/`At`/`From`/`To`,
  `Remove`, `Indent`, `StartsWith`/`EndsWith`, plus the inflector delegations
  (`Pluralize`, `Camelize`, `Titleize`, `Parameterize`, …).
- **Array**: `ArrayBlank`, `ArrayFrom`/`ArrayTo`, `Second`…`Fifth`, `InGroups`,
  `InGroupsOf` (+ `NoFill` variants), `Split`, `ToSentence`, `ExtractOptions`.
- **Hash**: `HashBlank`, `DeepMerge`(`Into`), `DeepDup`, `Except`, `Slice`,
  `ReverseMerge`, `DeepTransformValues`, `SymbolizeKeys`/`StringifyKeys` (+ deep
  variants), `AssertValidKeys`.
- **Object / Numeric**: `Blank`/`Present`/`Presence`, `Try` (via a `Dispatcher`
  seam), `Ordinal`/`Ordinalize`, `MultipleOf`.
- **Enumerable**: `IndexBy`, `Many`/`ManyBy`, `Exclude`, `Sum`, `Pluck`, `Pick`.

```go
import "github.com/go-ruby-activesupport/activesupport/coreext"

coreext.Squish("  a\n  b  c ")                 // "a b c"
coreext.Truncate("Once upon a time", 10, "…", " ") // "Once upon…"
coreext.ToSentence([]any{"a", "b", "c"})       // "a, b, and c"
coreext.DeepMerge(h1, h2)                       // recursive hash merge
coreext.Blank(nil)                              // true
```

## Fidelity basis

Every shipped surface is validated against **MRI Ruby with the `activesupport`
gem** by a differential oracle (`oracle_test.go` in each package): the Go output
is diffed **byte-for-byte** against the gem over a broad word/phrase list —
`pluralize`/`singularize`/`camelize`/`underscore`/`humanize`/`titleize`/
`tableize`/`classify`/`parameterize`/`transliterate`/`ordinalize` and the
core-ext helpers. The default inflection rules, irregulars, uncountables and the
transliteration table are ported **verbatim** from the gem's source. The oracle
runs on the CI ubuntu/macos lanes (gem installed) and skips where Ruby is absent;
the deterministic suite alone holds coverage at 100%, so the Windows and qemu
cross-arch lanes still pass the gate.

> Ruby's lookbehind/lookahead regexes (used by `camelize`/`underscore`/
> `titleize`) have no RE2 equivalent, so those are reimplemented by hand-written
> scanners whose **observable output** matches MRI exactly (verified by the
> oracle). Note: `transliterate` assumes NFC-normalised UTF-8 input (the
> realistic case for precomposed Latin text).

## Roadmap — deferred subsystems

v0.1 is deliberately the self-contained core. The following are **planned for
later phases** and are **not** in this release:

- **Phase 2 — object model**: `Concern`, `callbacks`, `delegate`, `Configurable`,
  `HashWithIndifferentAccess`, `StringInquirer`/`ArrayInquirer`.
- **Phase 3 — time**: `Duration`, `TimeWithZone`, `TimeZone`, time-travel
  helpers, `Date`/`Time`/`DateTime` core-ext.
- **Phase 4 — instrumentation & data**: `Notifications`, `Cache`, `Benchmark`,
  JSON encoding (`as_json`/`to_json`), `MessageEncryptor`/`MessageVerifier`.
- **Phase 5 — number & remaining core-ext**: `NumberHelper` (`to_s(:delimited/
  :rounded/:human/…)`), the long tail of String/Enumerable/Range extensions.

## Install

```sh
go get github.com/go-ruby-activesupport/activesupport
```

## Tests & coverage

```sh
GOWORK=off CGO_ENABLED=0 go test ./...
```

The suite enforces **100% line coverage** (every inflection rule, irregular and
acronym path, and every core-ext helper branch), runs the MRI differential oracle
where Ruby is present, and cross-compiles on all six 64-bit Go targets
(amd64/arm64/riscv64/loong64/ppc64le/s390x — the last two big-endian).

## License

BSD-3-Clause. Copyright (c) 2026, the go-ruby-activesupport/activesupport authors.
