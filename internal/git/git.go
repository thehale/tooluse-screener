// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package git

import (
	"slices"
	"strings"

	commands "github.com/thehale/tooluse-screener/internal/command"
	"github.com/thehale/tooluse-screener/internal/lists"
)

type Invocation struct {
	Git         bool
	Subcommand  string
	Arguments   []string
	Directories []string
	Global      []string
	Unread      []string
}

func Read(command string) Invocation {
	if words := strings.Fields(commands.WithoutAssignments(command)); len(words) == 0 || words[0] != "git" {
		return Invocation{Unread: unreadNames(command, nil)}
	} else {
		return invocationOf(command, words[1:])
	}
}

func invocationOf(command string, words []string) Invocation {
	directories, global, rest := afterGlobalOptions(append(twinOptions(command), words...))
	subcommand, arguments := subcommandAndArguments(rest)
	return Invocation{
		Git:         true,
		Subcommand:  subcommand,
		Arguments:   arguments,
		Directories: directories,
		Global:      global,
		Unread:      unreadNames(command, twins),
	}
}

func (i Invocation) IsA(subcommand string) bool {
	return i.Subcommand != "" && i.Subcommand == subcommand
}

func (i Invocation) HasAll(words []string) bool {
	return lists.Every(words, i.has)
}

func (i Invocation) has(word string) bool {
	if strings.HasPrefix(word, "-") {
		return slices.Contains(i.Arguments, word)
	} else {
		return len(i.Arguments) > 0 && i.Arguments[0] == word
	}
}

func afterGlobalOptions(words []string) (directories, global, rest []string) {
	var pointings []pointing
	for len(words) > 0 && strings.HasPrefix(words[0], "-") {
		option, value, isJoined := strings.Cut(words[0], "=")
		if isDirectoryOption(option) {
			pointings = append(pointings, pointing{option, directoryFrom(value, isJoined, words)})
		}
		global = append(global, option)
		words = words[stride(option, isJoined, len(words)):]
	}
	return directoriesOf(pointings), global, words
}

func isDirectoryOption(option string) bool {
	pointing := []string{"-C", "--git-dir", "--work-tree"}
	return slices.Contains(pointing, option)
}

func directoryFrom(value string, isJoined bool, words []string) string {
	switch {
	case isJoined:
		return commands.Literal(value)
	case len(words) > 1:
		return commands.Literal(words[1])
	default:
		return ""
	}
}

func stride(option string, isJoined bool, wordsLeft int) int {
	if isJoined || !hasAValue(option) || wordsLeft < 2 {
		return 1
	} else {
		return 2
	}
}

func hasAValue(option string) bool {
	withValues := []string{"-C", "-c", "--git-dir", "--namespace", "--work-tree"}
	return slices.Contains(withValues, option)
}

func subcommandAndArguments(rest []string) (subcommand string, arguments []string) {
	if len(rest) == 0 {
		return "", nil
	} else {
		return rest[0], rest[1:]
	}
}
