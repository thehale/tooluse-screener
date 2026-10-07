// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"slices"
	"strings"
	"unicode/utf8"
)

func (c Command) WithoutAssignments() Command {
	return Command{Text: WithoutAssignments(c.Text), Moved: c.Moved}
}

func WithoutAssignments(text string) string {
	words := strings.Fields(text)
	return strings.Join(words[len(Assignments(text)):], " ")
}

func Assignments(text string) []string {
	words := strings.Fields(text)
	assignmentCount := 0
	for assignmentCount < len(words) && isAnAssignment(words[assignmentCount]) {
		assignmentCount++
	}
	return words[:assignmentCount]
}

func IsRetargeting(name string) bool {
	return slices.Contains(retargetingNames, name) || strings.HasPrefix(name, "GIT_CONFIG")
}

var retargetingNames = []string{
	"GIT_DIR", "GIT_WORK_TREE", "GIT_NAMESPACE", "GIT_COMMON_DIR",
	"GH_REPO", "GH_HOST", "GH_CONFIG_DIR",
	"HOME", "XDG_CONFIG_HOME",
}

func isAnAssignment(word string) bool {
	name, _, isAssignment := strings.Cut(word, "=")
	return isAssignment && isIdentifier(name)
}

func isIdentifier(name string) bool {
	return name != "" && isNameStart(name[0]) && !strings.ContainsFunc(name, isNotPartOfAName)
}

func isNameStart(letter byte) bool {
	return letter == '_' ||
		('a' <= letter && letter <= 'z') ||
		('A' <= letter && letter <= 'Z')
}

func isNotPartOfAName(letter rune) bool {
	return letter >= utf8.RuneSelf || (!isNameStart(byte(letter)) && !isDigit(byte(letter)))
}

func isDigit(letter byte) bool {
	return '0' <= letter && letter <= '9'
}
