// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package activesupport is the module root for go-ruby-activesupport: a pure-Go
// (no cgo), MRI-faithful reimplementation of Ruby on Rails' ActiveSupport.
//
// This v0.1 foundation ships the self-contained, highest-value core:
//
//   - inflector — ActiveSupport::Inflector (pluralize/singularize, camelize/
//     underscore, humanize/titleize, tableize/classify, foreign_key,
//     ordinalize, parameterize, transliterate, and the full inflection-rule
//     registration API), verified byte-for-byte against the activesupport gem.
//   - coreext   — the highest-value String/Array/Hash/Object/Enumerable
//     core-extension helpers, as plain Go functions the rbgo binding maps onto
//     the corresponding Ruby monkey-patches.
//
// See the README for the roadmap covering the deferred subsystems (Concern,
// callbacks, HashWithIndifferentAccess, TimeZone/Duration, Notifications, Cache,
// MessageEncryptor/Verifier, JSON encoding, delegate, and more).
package activesupport

// Version is the module version.
const Version = "0.1.0"
