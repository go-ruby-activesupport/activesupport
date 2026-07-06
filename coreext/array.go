// Copyright (c) the go-ruby-activesupport/activesupport authors
//
// SPDX-License-Identifier: BSD-3-Clause

package coreext

// A Ruby Array maps to a Go []any; Ruby nil padding is represented by Go nil.

// ArrayBlank reports whether the array is empty (Array#blank?).
func ArrayBlank(a []any) bool { return len(a) == 0 }

// ArrayFrom returns the elements from index pos to the end (negative counts from
// the end), reporting false when pos is past the end (Array#from).
func ArrayFrom(a []any, pos int) ([]any, bool) {
	if pos < 0 {
		pos += len(a)
	}
	if pos < 0 || pos > len(a) {
		return nil, false
	}
	return a[pos:], true
}

// ArrayTo returns the elements from the start through index pos inclusive
// (negative counts from the end) (Array#to).
func ArrayTo(a []any, pos int) []any {
	if pos < 0 {
		pos += len(a)
	}
	if pos < 0 {
		return []any{}
	}
	if pos >= len(a) {
		return a
	}
	return a[:pos+1]
}

// nth returns the element at index i, or nil when out of range — the shared
// implementation of Second…Fifth.
func nth(a []any, i int) any {
	if i < 0 || i >= len(a) {
		return nil
	}
	return a[i]
}

// Second returns the second element or nil (Array#second).
func Second(a []any) any { return nth(a, 1) }

// Third returns the third element or nil (Array#third).
func Third(a []any) any { return nth(a, 2) }

// Fourth returns the fourth element or nil (Array#fourth).
func Fourth(a []any) any { return nth(a, 3) }

// Fifth returns the fifth element or nil (Array#fifth).
func Fifth(a []any) any { return nth(a, 4) }

// InGroups splits a into number groups, padding shorter groups with fill
// (Array#in_groups). Use InGroupsNoFill to leave short groups unpadded.
func InGroups(a []any, number int, fill any) [][]any {
	return inGroups(a, number, fill, true)
}

// InGroupsNoFill splits a into number groups without padding
// (Array#in_groups(n, false)).
func InGroupsNoFill(a []any, number int) [][]any {
	return inGroups(a, number, nil, false)
}

func inGroups(a []any, number int, fill any, doFill bool) [][]any {
	division := len(a) / number
	modulo := len(a) % number
	groups := make([][]any, 0, number)
	start := 0
	for index := 0; index < number; index++ {
		length := division
		if modulo > 0 && modulo > index {
			length++
		}
		g := append([]any(nil), a[start:start+length]...)
		if doFill && modulo > 0 && length == division {
			g = append(g, fill)
		}
		groups = append(groups, g)
		start += length
	}
	return groups
}

// InGroupsOf splits a into consecutive groups of number elements, padding the
// final group with fill (Array#in_groups_of).
func InGroupsOf(a []any, number int, fill any) [][]any {
	padding := (number - len(a)%number) % number
	collection := append([]any(nil), a...)
	for i := 0; i < padding; i++ {
		collection = append(collection, fill)
	}
	return sliceEvery(collection, number)
}

// InGroupsOfNoFill splits a into groups of number elements without padding the
// final group (Array#in_groups_of(n, false)).
func InGroupsOfNoFill(a []any, number int) [][]any {
	return sliceEvery(append([]any(nil), a...), number)
}

func sliceEvery(a []any, number int) [][]any {
	var out [][]any
	for i := 0; i < len(a); i += number {
		end := i + number
		if end > len(a) {
			end = len(a)
		}
		out = append(out, a[i:end])
	}
	return out
}

// Split divides a into subarrays separated by every element equal to sep
// (Array#split with a value).
func Split(a []any, sep any) [][]any {
	out := [][]any{{}}
	for _, e := range a {
		if e == sep {
			out = append(out, []any{})
		} else {
			out[len(out)-1] = append(out[len(out)-1], e)
		}
	}
	return out
}

// ExtractOptions removes and returns a trailing options hash, mirroring
// Array#extract_options!. It returns the remaining elements and the options (an
// empty map when the last element is not a map[any]any).
func ExtractOptions(a []any) ([]any, map[any]any) {
	if n := len(a); n > 0 {
		if opts, ok := a[n-1].(map[any]any); ok {
			return a[:n-1], opts
		}
	}
	return a, map[any]any{}
}

// ToSentence joins items into a human sentence using the default English
// connectors ("a, b, and c") (Array#to_sentence).
func ToSentence(items []any) string {
	return ToSentenceWith(items, ", ", " and ", ", and ")
}

// ToSentenceWith joins items using explicit connectors (Array#to_sentence with
// :words_connector / :two_words_connector / :last_word_connector).
func ToSentenceWith(items []any, wordsConnector, twoWordsConnector, lastWordConnector string) string {
	strs := make([]string, len(items))
	for i, it := range items {
		strs[i] = toS(it)
	}
	switch len(strs) {
	case 0:
		return ""
	case 1:
		return strs[0]
	case 2:
		return strs[0] + twoWordsConnector + strs[1]
	default:
		head := strs[:len(strs)-1]
		out := head[0]
		for _, h := range head[1:] {
			out += wordsConnector + h
		}
		return out + lastWordConnector + strs[len(strs)-1]
	}
}
