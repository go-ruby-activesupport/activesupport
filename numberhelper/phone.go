// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package numberhelper

import (
	"regexp"
	"strings"
)

var (
	phoneAreaCode   = regexp.MustCompile(`([0-9]{1,3})([0-9]{3})([0-9]{4}$)`)
	phoneNoAreaCode = regexp.MustCompile(`([0-9]{0,3})([0-9]{3})([0-9]{4})$`)
)

// NumberToPhone formats a number as a phone number, e.g. "(123) 555-1234"
// (number_to_phone).
func NumberToPhone(n any, opts ...Options) string {
	o := firstOpt(opts)
	delimiter := orStr(o.Delimiter, "-")
	number := strings.TrimSpace(numToString(n))

	var body string
	if o.AreaCode {
		pat := phoneAreaCode
		if o.Pattern != nil {
			pat = o.Pattern
		}
		body = pat.ReplaceAllString(number, "(${1}) ${2}"+delimiter+"${3}")
	} else {
		pat := phoneNoAreaCode
		if o.Pattern != nil {
			pat = o.Pattern
		}
		body = pat.ReplaceAllString(number, "${1}"+delimiter+"${2}"+delimiter+"${3}")
		if delimiter != "" && strings.HasPrefix(body, delimiter) {
			body = body[1:] // Ruby slice!(0, 1) removes a single leading char
		}
	}

	cc := ""
	if s := blankToStr(o.CountryCode); s != "" {
		cc = "+" + s + delimiter
	}
	ext := ""
	if s := blankToStr(o.Extension); s != "" {
		ext = " x " + s
	}
	return cc + body + ext
}

// blankToStr renders a country-code/extension value to its string, treating nil
// and "" as blank.
func blankToStr(v any) string {
	if v == nil {
		return ""
	}
	return numToString(v)
}
