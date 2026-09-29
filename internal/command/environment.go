// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"slices"
	"strings"
	"unicode/utf8"
)

func Spoken(text string) string {
	words := strings.Fields(text)
	return strings.Join(words[len(Assignments(text)):], " ")
}

func Assignments(text string) []string {
	words := strings.Fields(text)
	leading := 0
	for leading < len(words) && isAnAssignment(words[leading]) {
		leading++
	}
	return words[:leading]
}

func Retargets(name string) bool {
	return slices.Contains(retargeting, name) || strings.HasPrefix(name, "GIT_CONFIG")
}

var retargeting = []string{
	"GIT_DIR", "GIT_WORK_TREE", "GIT_NAMESPACE", "GIT_COMMON_DIR",
	"GH_REPO", "GH_HOST", "GH_CONFIG_DIR",
	"HOME", "XDG_CONFIG_HOME",
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
