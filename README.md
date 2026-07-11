<p align="center"><img src="https://raw.githubusercontent.com/go-ruby-activesupport/brand/main/social/go-ruby-activesupport-activesupport.png" alt="go-ruby-activesupport/activesupport" width="720"></p>

# activesupport — go-ruby-activesupport

[![Docs](https://img.shields.io/badge/docs-mkdocs--material-DC2626)](https://go-ruby-activesupport.github.io/docs/)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26.4%2B-00ADD8)](https://go.dev/dl/)
[![Coverage](https://img.shields.io/badge/coverage-100%25-1a7f37)](#tests--coverage)

**A pure-Go (no cgo) reimplementation of Ruby on Rails'
[ActiveSupport](https://guides.rubyonrails.org/active_support_core_extensions.html)**
— faithful to MRI down to byte-for-byte output, with **no Ruby runtime**.

ActiveSupport is enormous. It began as a self-contained core (the **Inflector**
and the **core-extension helpers**) and now also ships the **object-model**,
**time**, **instrumentation/data** and **number** subsystems documented below —
each pure-Go, 100%-covered, and validated byte-for-byte against the
`activesupport` gem. It is the ActiveSupport backend for
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

### `duration` — `ActiveSupport::Duration`

Single-unit constructors (`Years`…`Seconds`), `Build` decomposition, humanised
`Inspect` (`"1 year, 2 months, and 3 days"`), ISO 8601 `Iso8601`/`Parse` (with
Rails' exact validation — weeks may not mix with other date parts, only the last
part may be fractional), part-preserving arithmetic (`Add`/`Sub`/`Mul`/`Neg`),
`Cmp`/`Equal`, the `In<Unit>` conversions, and calendar-aware `Since`/`Ago`.

```go
duration.Days(1).Add(duration.Hours(2)).Inspect() // "1 day and 2 hours"
duration.Hours(1.5).Iso8601()                     // "PT1.5H"
d, _ := duration.Parse("P1Y2M3DT4H5M6S")          // full ISO 8601 parse
```

### `numberhelper` — `ActiveSupport::NumberHelper`

`NumberToDelimited`/`Rounded`/`Percentage`/`Currency`/`HumanSize`/`Human`/`Phone`.
Decimal rounding runs on `math/big` rationals (not binary floats) so every
BigDecimal round mode (`half_up`/`half_even`/`half_down`/`up`/`down`/`ceiling`/
`floor`), significant-digit precision and strip-insignificant-zeros behave
exactly like Rails, including the invalid-input passthrough (`"$abc"`).

```go
numberhelper.NumberToCurrency(1234567890.50)                    // "$1,234,567,890.50"
numberhelper.NumberToHuman(1234567)                             // "1.23 Million"
numberhelper.NumberToRounded(1234.5678, numberhelper.Options{
    Precision:   numberhelper.IntPtr(2),
    Significant: numberhelper.BoolPtr(true),
}) // "1200"
```

### `inquirer` — `StringInquirer` / `ArrayInquirer`

Ruby's `method_missing` `"<name>?"` predicate becomes an explicit `Is(name)`;
`ArrayInquirer#any?` becomes `Any(candidates...)`.

### `hwia` — `HashWithIndifferentAccess` + `OrderedOptions`

String-keyed, insertion-ordered, with recursive nested-hash conversion (including
`map[any]any` and arrays of hashes); `Get`/`Set`/`Fetch`/`Delete`/`KeyQ`/`Merge`/
`Update`/`Dup`/`Slice`/`Except`/`ValuesAt`/`ToHash`, plus `OrderedOptions` with
its blank-raising bang accessor.

### `notifications` — `ActiveSupport::Notifications`

An instrumentation bus: `Instrument(name, payload, block)` publishing an `Event`
(`Start`/`Finish`/`Duration`/`TransactionID`) to subscribers matched by `Exact`
name, `Pattern` (regexp) or `All`; `Unsubscribe` and scoped `Subscribed`. The
event is published even when the block panics, matching Rails' ensure semantics.

### `cache` — `ActiveSupport::Cache::MemoryStore`

Concurrency-safe in-memory cache: `Read`/`Write`/`Exist`/`Delete`, `Fetch`
(compute-on-miss + `Force`), `Increment`/`Decrement` (zero-initialising,
expiry-preserving), `Clear`, `Cleanup`, and the `*Multi` variants. Per-entry
expiry is driven by an injectable clock.

### `callbacks` — `ActiveSupport::Callbacks`

`Before`/`After`/`Around` chains run around a block with Rails' exact ordering
(befores in order, arounds nested first-outermost, afters in reverse) and halt
semantics (a `Before` returning false = `throw :abort`).

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

The following are **still** deferred (they need the Ruby object model / method
dispatch, or large i18n/locale/crypto surfaces that no pure-Go core can honestly
claim yet), and are called out explicitly rather than half-shipped:

- **Object model needing dispatch**: `Concern`, `delegate`, `Configurable`,
  `Module#delegate_missing_to` — these mix into the Ruby module/method system and
  belong in the rbgo binding, not a standalone Go package.
- **Time-with-zone**: `TimeWithZone` and the full `TimeZone` MAPPING (the
  `Duration` calendar/format core and `Since`/`Ago` anchoring ship in the
  [`duration`](duration) package; the friendly-name↔IANA zone table and a
  zone-aware wrapper type remain).
- **Data/crypto**: JSON encoding (`as_json`/`to_json`),
  `MessageEncryptor`/`MessageVerifier`, `Benchmark`.
- **Remaining core-ext long tail**: the less-used `Date`/`Time`/`DateTime`,
  `Range` and `Integer` extensions beyond those already in `coreext`.

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

## WebAssembly

Being pure Go (CGO=0), this library also compiles to **WebAssembly** — both
`GOOS=js GOARCH=wasm` (browser / Node.js) and `GOOS=wasip1 GOARCH=wasm` (WASI).
CI builds both targets on every push, alongside the six 64-bit native/qemu arches.

```sh
GOOS=js     GOARCH=wasm go build ./...   # browser / Node
GOOS=wasip1 GOARCH=wasm go build ./...   # WASI (wasmtime, wasmer, wasmedge, …)
```
