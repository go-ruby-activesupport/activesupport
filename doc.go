// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package activesupport is the module root for go-ruby-activesupport: a pure-Go
// (no cgo), MRI-faithful reimplementation of Ruby on Rails' ActiveSupport.
//
// Sub-packages, each 100%-covered and validated against the activesupport gem:
//
//   - inflector    — ActiveSupport::Inflector (pluralize/singularize, camelize/
//     underscore, humanize/titleize, tableize/classify, foreign_key,
//     ordinalize, parameterize, transliterate, and the full inflection-rule
//     registration API).
//   - coreext      — the highest-value String/Array/Hash/Object/Enumerable
//     core-extension helpers.
//   - duration     — ActiveSupport::Duration (build/inspect/iso8601/parse,
//     part-preserving arithmetic, in_<unit>, calendar-aware since/ago).
//   - numberhelper — ActiveSupport::NumberHelper (delimited/rounded/percentage/
//     currency/human_size/human/phone) with big.Rat BigDecimal rounding.
//   - inquirer     — StringInquirer / ArrayInquirer.
//   - hwia         — HashWithIndifferentAccess + OrderedOptions.
//   - notifications — ActiveSupport::Notifications (instrument/subscribe).
//   - cache        — ActiveSupport::Cache::MemoryStore.
//   - callbacks    — ActiveSupport::Callbacks (before/after/around + halt).
//
// See the README for the subsystems that remain deferred (Concern/delegate and
// other dispatch-bound object-model pieces, TimeWithZone/TimeZone, JSON
// encoding, MessageEncryptor/Verifier).
package activesupport

// Version is the module version.
const Version = "0.2.0"
