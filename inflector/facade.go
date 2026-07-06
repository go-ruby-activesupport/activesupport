// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package inflector

// The package-level functions operate on DefaultLocale (Rails' English rules),
// mirroring the module methods on ActiveSupport::Inflector. Use the methods on a
// cloned *Inflections when you need custom rules.

// Pluralize returns the plural of word using the default English rules.
func Pluralize(word string) string { return DefaultLocale.Pluralize(word) }

// Singularize returns the singular of word using the default English rules.
func Singularize(word string) string { return DefaultLocale.Singularize(word) }

// Camelize converts word to UpperCamelCase.
func Camelize(word string) string { return DefaultLocale.Camelize(word, true) }

// CamelizeLower converts word to lowerCamelCase.
func CamelizeLower(word string) string { return DefaultLocale.Camelize(word, false) }

// Underscore converts word to snake_case.
func Underscore(word string) string { return DefaultLocale.Underscore(word) }

// Humanize turns an attribute name into a human-friendly, capitalized phrase.
func Humanize(word string) string { return DefaultLocale.Humanize(word, true, false) }

// Titleize turns a phrase into a pretty, capitalized title.
func Titleize(word string) string { return DefaultLocale.Titleize(word, false) }

// Tableize turns a model name into its table name.
func Tableize(className string) string { return DefaultLocale.Tableize(className) }

// Classify turns a table name into a class name.
func Classify(tableName string) string { return DefaultLocale.Classify(tableName) }

// ForeignKey builds a foreign-key column name (with a "_id" suffix) from a class
// name.
func ForeignKey(className string) string { return DefaultLocale.ForeignKey(className, true) }
