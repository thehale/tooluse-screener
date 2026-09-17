// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"strings"
	"unicode/utf8"
)

func Spoken(text string) string {
	words := strings.Fields(text)
	for len(words) > 0 && isAnAssignment(words[0]) {
		words = words[1:]
	}
	return strings.Join(words, " ")
}

func isAnAssignment(word string) bool {
	name, _, assigned := strings.Cut(word, "=")
	return assigned && isIdentifier(name)
}

func isIdentifier(name string) bool {
	return name != "" && opensAName(name[0]) && !strings.ContainsFunc(name, isNotPartOfAName)
}

func opensAName(letter byte) bool {
	return letter == '_' ||
		('a' <= letter && letter <= 'z') ||
		('A' <= letter && letter <= 'Z')
}

func isNotPartOfAName(letter rune) bool {
	return letter >= utf8.RuneSelf || (!opensAName(byte(letter)) && !isDigit(byte(letter)))
}

func isDigit(letter byte) bool {
	return '0' <= letter && letter <= '9'
}
