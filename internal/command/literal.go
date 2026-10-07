// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import "strings"

func Literal(word string) string {
	inner, isQuoted := withoutQuotes(word)
	switch {
	case strings.ContainsAny(inner, expanding):
		return ""
	case isQuoted && strings.HasPrefix(inner, "~"):
		return ""
	default:
		return inner
	}
}

const expanding = "$`\\\"'*?[]{}" + Unquotable

func withoutQuotes(word string) (inner string, isQuoted bool) {
	if len(word) >= 2 && strings.ContainsRune(`'"`, rune(word[0])) && word[0] == word[len(word)-1] {
		return word[1 : len(word)-1], true
	} else {
		return word, false
	}
}
