// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import "strings"

func Literal(word string) string {
	unquoted, quoted := withoutQuotes(word)
	switch {
	case strings.ContainsAny(unquoted, expanding):
		return ""
	case quoted && strings.HasPrefix(unquoted, "~"):
		return ""
	default:
		return unquoted
	}
}

const expanding = "$`\\\"'*?[]{}" + Unquotable

func withoutQuotes(word string) (unquoted string, quoted bool) {
	if len(word) >= 2 && strings.ContainsRune(`'"`, rune(word[0])) && word[0] == word[len(word)-1] {
		return word[1 : len(word)-1], true
	} else {
		return word, false
	}
}
