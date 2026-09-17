// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

// Package git reads a git command past its global options.
package git

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/thehale/tooluse-screener/internal/lists"
)

type Invocation struct {
	Subcommand  string
	Arguments   []string
	Directories []string
}

// Read returns an Invocation with no subcommand for a command that is
// not git's.
func Read(command string) Invocation {
	words := afterEnvironmentAssignments(strings.Fields(command))
	if len(words) == 0 || words[0] != "git" {
		return Invocation{}
	}
	directories, spoken := afterGlobalOptions(words[1:])
	return spokenAs(spoken, directories)
}

func (i Invocation) IsA(subcommand string) bool {
	return i.Subcommand != "" && i.Subcommand == subcommand
}

// Carries counts an option anywhere, since git permutes them, and a
// bare word only first, where it is git's own word.
func (i Invocation) Carries(words []string) bool {
	return lists.Every(words, i.carries)
}

func (i Invocation) carries(word string) bool {
	if strings.HasPrefix(word, "-") {
		return slices.Contains(i.Arguments, word)
	}
	return len(i.Arguments) > 0 && i.Arguments[0] == word
}

func afterEnvironmentAssignments(words []string) []string {
	for len(words) > 0 && isAnAssignment(words[0]) {
		words = words[1:]
	}
	return words
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

func afterGlobalOptions(words []string) (directories, spoken []string) {
	for len(words) > 0 && strings.HasPrefix(words[0], "-") {
		option, value, joined := strings.Cut(words[0], "=")
		if option == "-C" {
			directories = append(directories, directoryFrom(value, joined, words))
		}
		words = words[stride(option, joined, len(words)):]
	}
	return directories, words
}

func directoryFrom(value string, joined bool, words []string) string {
	switch {
	case joined:
		return value
	case len(words) > 1:
		return words[1]
	default:
		return ""
	}
}

func stride(option string, joined bool, remaining int) int {
	if joined || !takesAValue(option) || remaining < 2 {
		return 1
	}
	return 2
}

func takesAValue(option string) bool {
	withValues := []string{"-C", "-c", "--git-dir", "--namespace", "--work-tree"}
	return slices.Contains(withValues, option)
}

func spokenAs(spoken, directories []string) Invocation {
	if len(spoken) == 0 {
		return Invocation{Directories: directories}
	}
	return Invocation{Subcommand: spoken[0], Arguments: spoken[1:], Directories: directories}
}
